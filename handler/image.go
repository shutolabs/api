package handler

import (
	"net/http"
	"strings"

	"shuto-api/config"
	"shuto-api/security"
	"shuto-api/utils"
)

// ImageHandler processes image transformation requests
// @Summary Process and transform an image
// @Description Get an image with optional transformations applied
// @Tags image
// @Accept  json
// @Produce  image/jpeg,image/png,image/webp
// @Param   path     path    string     true        "Path to the image file"
// @Param   w        query   int        false       "Output image width in pixels"
// @Param   h        query   int        false       "Output image height in pixels"
// @Param   fit      query   string     false       "Resize mode: clip, crop, fill" Enums(clip,crop,fill)
// @Param   fm       query   string     false       "Output format: jpg, jpeg, png, webp" Enums(jpg,jpeg,png,webp)
// @Param   q        query   int        false       "Compression quality (1-100)"
// @Param   dpr      query   number     false       "Device pixel ratio (1-3)"
// @Param   blur     query   int        false       "Gaussian blur intensity (0-100)"
// @Param   dl       query   bool       false       "Force download instead of display"
// @Success 200 {file}  []byte
// @Failure 400 {object} utils.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized - Invalid signature"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - Invalid signature"
// @Failure 404 {object} utils.ErrorResponse "Image not found"
// @Failure 410 {object} utils.ErrorResponse "Gone - Token expired"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /image/{path} [get]
func ImageHandler(w http.ResponseWriter, r *http.Request, imgUtils utils.ImageUtils, rclone utils.Rclone, domainConfig config.DomainConfigManager) {
	if r.Method != http.MethodGet {
		utils.WriteInvalidRequestError(w, "Method not allowed", r.Method)
		return
	}

	domain := utils.GetDomainFromRequest(r)
	path := strings.TrimPrefix(r.URL.Path, "/"+config.ApiVersion+"/image/")
	if path == "" {
		utils.WriteInvalidPathError(w, "Path is required")
		return
	}
	
	cfg, err := domainConfig.GetDomainConfig(domain)
	if err != nil {
		utils.WriteInvalidDomainError(w, domain)
		return
	}

	// Validate signed URL if security is enabled
	if cfg.Security.Mode != "" {
		if err := security.ValidateSignedURLFromConfig(path, r.URL.Query(), cfg.Security.Secrets, cfg.Security.ValidityWindow); err != nil {
			switch err {
			case security.ErrKeyNotFound:
				utils.WriteUnauthorizedError(w, "Invalid security key")
			case security.ErrExpiredURL:
				utils.WriteExpiredTokenError(w)
			default:
				utils.WriteInvalidSignatureError(w)
			}
			return
		}
	}

	// Check if path is a directory
	files, err := rclone.ListPath(path, domain)
	if err == nil && len(files) > 0 && files[0].IsDir {
		utils.WriteInvalidRequestError(w, "Cannot serve directory as image", path)
		return
	}

	imgData, err := rclone.FetchImage(path, domain)
	if err != nil {
		utils.WriteNotFoundError(w, "Failed to fetch image", err.Error())
		return
	}

	options := utils.ParseImageOptionsFromRequest(r)

	// If no format is specified, automatically select the best format based on browser support
	if options.Format == "" {
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "image/webp") {
			options.Format = "webp"
		} else if strings.Contains(accept, "image/avif") {
			options.Format = "avif"
		} else {
			options.Format = "jpg" // Default to JPEG as fallback
		}
	}

	modifiedImg, err := imgUtils.TransformImage(imgData, options)
	if err != nil {
		utils.WriteInternalError(w, "Failed to transform image", err.Error())
		return
	}

	mimeType, err := imgUtils.GetMimeType(modifiedImg)
	if err != nil {
		utils.WriteInternalError(w, "Failed to get MIME type", err.Error())
		return
	}

	if options.ForceDownload {
		w.Header().Set("Content-Disposition", "attachment")
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(modifiedImg)
}

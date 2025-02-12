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
// @Failure 400 {string} string "Invalid parameters"
// @Failure 404 {string} string "Image not found"
// @Failure 500 {string} string "Internal server error"
// @Router /image/{path} [get]
func ImageHandler(w http.ResponseWriter, r *http.Request, imgUtils utils.ImageUtils, rclone utils.Rclone, domainConfig config.DomainConfigManager) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	domain := utils.GetDomainFromRequest(r)
	path := strings.TrimPrefix(r.URL.Path, "/"+config.ApiVersion+"/image/")
	if path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}
	
	cfg, err := domainConfig.GetDomainConfig(domain)
	if err != nil {
		utils.Error("Failed to get domain config", "error", err, "domain", domain)
		http.Error(w, "Invalid domain", http.StatusBadRequest)
		return
	}

	// Validate signed URL if security is enabled
	if cfg.Security.Mode != "" {
		if err := security.ValidateSignedURLFromConfig(path, r.URL.Query(), cfg.Security.Secrets, cfg.Security.ValidityWindow); err != nil {
			utils.Error("Invalid signed URL", "error", err, "path", path)
			var status int
			switch err {
			case security.ErrKeyNotFound:
				status = http.StatusUnauthorized
			case security.ErrExpiredURL:
				status = http.StatusGone
			default:
				status = http.StatusForbidden
			}
			http.Error(w, err.Error(), status)
			return
		}
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

	utils.Debug("Processing image request", 
		"domain", domain,
		"path", path,
		"options", options,
	)

	// Check if path is a directory
	files, err := rclone.ListPath(path, domain)
	if err == nil && len(files) > 0 && files[0].IsDir {
		// If the path points to a directory
		utils.Error("Cannot serve directory as image", "path", path)
		http.Error(w, "Cannot serve directory as image", http.StatusBadRequest)
		return
	}

	imgData, err := rclone.FetchImage(path, domain)
	if err != nil {
		utils.Error("Failed to fetch image", "error", err, "path", path)
		http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
		return
	}

	modifiedImg, err := imgUtils.TransformImage(imgData, options)
	if err != nil {
		utils.Error("Failed to transform image", "error", err, "options", options)
		http.Error(w, "Failed to transform image", http.StatusInternalServerError)
		return
	}

	mimeType, err := imgUtils.GetMimeType(modifiedImg)
	if err != nil {
		utils.Error("Failed to get MIME type", "error", err)
		http.Error(w, "Failed to process image", http.StatusInternalServerError)
		return
	}

	utils.Debug("Image processed successfully",
		"path", path,
		"mimeType", mimeType,
		"size", len(modifiedImg),
	)

	if options.ForceDownload {
		w.Header().Set("Content-Disposition", "attachment")
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(modifiedImg)
}

package handler

import (
	"net/http"
	"strings"

	"shuto-api/config"
	"shuto-api/security"
	"shuto-api/utils"
)

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

	utils.Debug("Processing image request", 
		"domain", domain,
		"path", path,
		"options", options,
	)

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

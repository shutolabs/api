package handler

import (
	"net/http"
	"strconv"
	"strings"

	"shuto-api/config"
	"shuto-api/security"
	"shuto-api/utils"
)

// ImageHandler processes image transformations based on query parameters
func ImageHandler(w http.ResponseWriter, r *http.Request, imgUtils utils.ImageUtils, rclone utils.Rclone, domainConfig config.DomainConfigManager) {
	domain := utils.GetDomainFromRequest(r)
	
	// Get domain configuration
	cfg, err := domainConfig.GetDomainConfig(domain)
	if err != nil {
		utils.Error("Failed to get domain config", "error", err, "domain", domain)
		http.Error(w, "Invalid domain", http.StatusBadRequest)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/"+config.ApiVersion+"/image/")

	// Validate signed URL if security is enabled
	if cfg.Security.Mode != "" {
		// Convert config secrets to security.SecretKey
		keys := make([]security.SecretKey, len(cfg.Security.Secrets))
		for i, secret := range cfg.Security.Secrets {
			keys[i] = security.SecretKey{
				ID:     secret.KeyID,
				Secret: []byte(secret.Secret),
			}
		}

		signer, err := security.NewURLSigner(keys, cfg.Security.ValidityWindow, "")
		if err != nil {
			utils.Error("Failed to create URL signer", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err := signer.ValidateSignedURL(path, r.URL.Query()); err != nil {
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
	
	width, _ := strconv.Atoi(r.URL.Query().Get("w"))
	height, _ := strconv.Atoi(r.URL.Query().Get("h"))
	fit := r.URL.Query().Get("fit")
	if fit == "" {
		fit = "clip" // Default as per spec
	}
	
	dpr, err := strconv.ParseFloat(r.URL.Query().Get("dpr"), 64)
	if err != nil || dpr == 0 {
		dpr = 1.0 // Default as per spec
	}
	if dpr > 3.0 {
		dpr = 3.0 // Max value as per spec
	}
	
	format := r.URL.Query().Get("fm")
	quality, err := strconv.Atoi(r.URL.Query().Get("q"))
	if err != nil || quality == 0 {
		quality = 75 // Default as per spec
	}
	
	blur, _ := strconv.Atoi(r.URL.Query().Get("blur"))
	forceDownload := r.URL.Query().Get("dl") == "1"

	options := utils.ImageTransformOptions{
		Width:         width,
		Height:        height,
		Fit:          fit,
		Format:       format,
		Quality:      quality,
		Dpr:          dpr,
		Blur:         blur,
		ForceDownload: forceDownload,
	}

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

	if forceDownload {
		w.Header().Set("Content-Disposition", "attachment")
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(modifiedImg)
}

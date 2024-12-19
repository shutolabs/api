package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"shuto-api/config"
	"shuto-api/utils"
)

// ImageHandler processes image transformations based on query parameters
func ImageHandler(w http.ResponseWriter, r *http.Request, imgUtils utils.ImageUtils, rclone utils.Rclone) {
	// Get domain configuration
	domain := utils.GetDomainFromRequest(r)

	// Extract path and parameters
	path := strings.TrimPrefix(r.URL.Path, "/"+config.ApiVersion+"/image/")
	
	// Parse parameters according to spec
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
		"remoteAddr", r.RemoteAddr,
		"queryParams", r.URL.RawQuery,
	)

	utils.Debug("Image transform options",
		"width", width,
		"height", height,
		"fit", fit,
		"format", format,
		"quality", quality,
		"dpr", dpr,
	)

	// Fetch image data using rclone with domain-specific config
	imgData, err := rclone.FetchImage(path, domain)
	if err != nil {
		utils.Error("Failed to fetch image",
			"error", err,
			"path", path,
			"domain", domain,
		)
		http.Error(w, fmt.Sprintf("Failed to fetch image: %v", err), http.StatusInternalServerError)
		return
	}

	// Transform the image
	modifiedImg, err := imgUtils.TransformImage(imgData, options)
	if err != nil {
		utils.Error("Failed to transform image",
			"error", err,
			"path", path,
			"options", options,
		)
		http.Error(w, fmt.Sprintf("Failed to transform image: %v", err), http.StatusInternalServerError)
		return
	}

	mimeType, err := imgUtils.GetMimeType(modifiedImg)
	if err != nil {
		utils.Error("Failed to get image MIME type",
			"error", err,
			"path", path,
		)
		http.Error(w, fmt.Sprintf("Failed to get image format: %v", err), http.StatusInternalServerError)
		return
	}

	utils.Debug("Successfully processed image",
		"path", path,
		"mimeType", mimeType,
		"outputSize", len(modifiedImg),
	)

	// Set response headers and write the modified image as an HTTP response
	if forceDownload {
		w.Header().Set("Content-Disposition", "attachment")
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(modifiedImg)
}

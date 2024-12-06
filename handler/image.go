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

	// Fetch image data using rclone with domain-specific config
	imgData, err := rclone.FetchImage(path, domain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch image: %v", err), http.StatusInternalServerError)
		return
	}

	// Transform the image
	modifiedImg, err := imgUtils.TransformImage(imgData, options)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to transform image: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers and write the modified image as an HTTP response
	mimeType, err := imgUtils.GetMimeType(modifiedImg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get image format: %v", err), http.StatusInternalServerError)
		return
	}

	// Set download header if requested
	if forceDownload {
		w.Header().Set("Content-Disposition", "attachment")
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(modifiedImg)
}

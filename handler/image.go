package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"shuto-api/service"
	"shuto-api/utils"
)

// ImageHandler processes image transformations based on query parameters
func ImageHandler(w http.ResponseWriter, r *http.Request) {
	// Extract path and parameters
	path := strings.TrimPrefix(r.URL.Path, "/image/")
	width, _ := strconv.Atoi(r.URL.Query().Get("w"))
	height, _ := strconv.Atoi(r.URL.Query().Get("h"))
	format := r.URL.Query().Get("format")
	crop := r.URL.Query().Get("crop") == "true"
	fit := r.URL.Query().Get("fit")
	quality, _ := strconv.Atoi(r.URL.Query().Get("quality"))
	dpr, _ := strconv.ParseFloat(r.URL.Query().Get("dpr"), 64)

	if dpr == 0 {
		dpr = 1.0
	}

	options := service.ImageTransformOptions{
		Width:   uint(width),
		Height:  uint(height),
		Crop:    crop,
		Format:  format,
		Quality: uint(quality),
		Dpr:     dpr,
		Fit:     fit,
	}

	// Fetch image data using rclone
	imgData, err := utils.FetchImage(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch image: %v", err), http.StatusInternalServerError)
		return
	}

	// Transform the image
	modifiedImg, err := service.TransformImage(imgData, options)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to transform image: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers and write the modified image as an HTTP response
	format, err = utils.GetMimeType(imgData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get image format: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", format)
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(modifiedImg)
}

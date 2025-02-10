package handler

import (
	"net/http"
	"strconv"
	"strings"

	"shuto-api/config"
	"shuto-api/utils"
)

// ImageHandler processes image transformations based on query parameters
func ImageHandler(w http.ResponseWriter, r *http.Request, imgUtils utils.ImageUtils, rclone utils.Rclone) {
	domain := utils.GetDomainFromRequest(r)
	path := strings.TrimPrefix(r.URL.Path, "/"+config.ApiVersion+"/image/")
	
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

package handler

import (
	"archive/zip"
	"net/http"
	"path/filepath"
	"strings"

	"shuto-api/config"
	"shuto-api/utils"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request, imageUtils utils.ImageUtils, rclone utils.Rclone) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	domain := utils.GetDomainFromRequest(r)
	path := strings.TrimPrefix(r.URL.Path, "/"+config.ApiVersion+"/download/")
	if path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	utils.Debug("Processing download request", "domain", domain, "path", path)

	files, err := rclone.ListPath(path, domain)
	if err != nil {
		utils.Error("Failed to list files", "error", err, "path", path)
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	var totalSize int64
	for _, file := range files {
		if !file.IsDir {
			totalSize += file.Size
		}
	}

	const maxSize = 1 * 1024 * 1024 * 1024 // 1GB
	if totalSize > maxSize {
		utils.Warn("Download size exceeds limit", "size", totalSize, "max", maxSize)
		http.Error(w, "Requested files exceed maximum allowed size", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(path)+".zip\"")

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, file := range files {
		if file.IsDir {
			continue
		}

		filePath := filepath.Join(path, file.Name)
		content, err := rclone.FetchImage(filePath, domain)
		if err != nil {
			utils.Error("Failed to fetch file", "error", err, "file", filePath)
			continue
		}

		f, err := zipWriter.Create(file.Name)
		if err != nil {
			utils.Error("Failed to create zip entry", "error", err, "file", file.Name)
			continue
		}

		if _, err := f.Write(content); err != nil {
			utils.Error("Failed to write to zip", "error", err, "file", file.Name)
			continue
		}

		utils.Debug("Added file to zip", "file", file.Name, "size", len(content))
	}
}
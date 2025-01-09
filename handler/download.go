package handler

import (
	"archive/zip"
	"fmt"
	"net/http"
	"path/filepath"
	"shuto-api/utils"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request, imageUtils utils.ImageUtils, rclone utils.Rclone) {
	domain := utils.GetDomainFromRequest(r)
	if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
	}

	// Extract the folder path from the URL similar to ListHandler
	folderPath := r.URL.Path[len("/v1/download/"):]
	if folderPath == "" {
			http.Error(w, "Folder path is required", http.StatusBadRequest)
			return
	}

	// List all files in the folder
	files, err := rclone.ListPath(folderPath, domain)
	if err != nil {
			utils.Error("Failed to list files", "error", err)
			http.Error(w, "Failed to list files", http.StatusInternalServerError)
			return
	}

	// Calculate total size and enforce limits
	var totalSize int64
	for _, file := range files {
			if !file.IsDir {
					totalSize += file.Size
			}
	}

	// Example limit of 1GB
	const maxSize = 1 * 1024 * 1024 * 1024
	if totalSize > maxSize {
			http.Error(w, "Requested files exceed maximum allowed size", http.StatusBadRequest)
			return
	}

	// Set headers for zip file download
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", filepath.Base(folderPath)))

	// Create zip writer
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Add each file to the zip
	for _, file := range files {
			// Skip directories
			if file.IsDir {
					continue
			}

			// Get file content
			content, err := rclone.FetchImage(filepath.Join(folderPath, file.Name), domain)
			if err != nil {
					utils.Error("Failed to get file content", "error", err, "file", file.Name)
					continue
			}

			// Create file in zip
			f, err := zipWriter.Create(file.Name)
			if err != nil {
					utils.Error("Failed to create file in zip", "error", err, "file", file.Name)
					continue
			}

			// Write content to zip file
			_, err = f.Write(content)
			if err != nil {
					utils.Error("Failed to write file content to zip", "error", err, "file", file.Name)
					continue
			}
	}
}
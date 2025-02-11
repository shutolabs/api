package handler

import (
	"archive/zip"
	"net/http"
	"path/filepath"
	"strings"

	"shuto-api/config"
	"shuto-api/security"
	"shuto-api/utils"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request, imageUtils utils.ImageUtils, rclone utils.Rclone, domainConfig config.DomainConfigManager) {
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

	cfg, err := domainConfig.GetDomainConfig(domain)
	if err != nil {
		utils.Error("Failed to get domain config", "error", err, "domain", domain)
		http.Error(w, "Invalid domain", http.StatusBadRequest)
		return
	}

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

	utils.Debug("Processing download request", "domain", domain, "path", path)

	files, err := rclone.ListPath(path, domain)
	if err != nil {
		utils.Error("Failed to list files", "error", err, "path", path)
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	// If no files found or only one file that's not a directory, treat as single file download
	if len(files) == 0 || (len(files) == 1 && !files[0].IsDir) {
		handleSingleFileDownload(w, r, path, domain, imageUtils, rclone)
		return
	}

	// Handle folder download with multiple files
	handleFolderDownload(w, r, path, files, domain, imageUtils, rclone)
}

func handleSingleFileDownload(w http.ResponseWriter, r *http.Request, path string, domain string, imageUtils utils.ImageUtils, rclone utils.Rclone) {
	content, err := rclone.FetchImage(path, domain)
	if err != nil {
		utils.Error("Failed to fetch file", "error", err, "path", path)
		http.Error(w, "Failed to fetch file", http.StatusInternalServerError)
		return
	}

	// Process image if it's an image file and has transformation parameters
	if utils.IsImageFile(path) && utils.HasImageTransformParams(r) {
		options := utils.ParseImageOptionsFromRequest(r)
		content, err = imageUtils.TransformImage(content, options)
		if err != nil {
			utils.Error("Failed to transform image", "error", err, "options", options)
			http.Error(w, "Failed to transform image", http.StatusInternalServerError)
			return
		}
	}

	mimeType := http.DetectContentType(content)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(path)+"\"")
	w.Write(content)
}

func handleFolderDownload(w http.ResponseWriter, r *http.Request, path string, files []utils.RcloneFile, domain string, imageUtils utils.ImageUtils, rclone utils.Rclone) {
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

	options := utils.ParseImageOptionsFromRequest(r)
	hasTransformParams := utils.HasImageTransformParams(r)

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

		// Process image if it's an image file and has transformation parameters
		if hasTransformParams && utils.IsImageFile(filePath) {
			content, err = imageUtils.TransformImage(content, options)
			if err != nil {
				utils.Error("Failed to transform image", "error", err, "file", filePath)
				continue
			}
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
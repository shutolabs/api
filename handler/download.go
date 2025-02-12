package handler

import (
	"archive/zip"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"

	"shuto-api/config"
	"shuto-api/security"
	"shuto-api/utils"
)

// DownloadHandler handles file download requests
// @Summary Download a file
// @Description Download a file from the specified path
// @Tags download
// @Accept  json
// @Produce  octet-stream
// @Param   path     path    string     true        "Path to the file to download"
// @Success 200 {file}  []byte
// @Failure 400 {object} utils.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized - Invalid signature"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - Invalid signature"
// @Failure 404 {object} utils.ErrorResponse "File not found"
// @Failure 410 {object} utils.ErrorResponse "Gone - Token expired"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /download/{path} [get]
func DownloadHandler(w http.ResponseWriter, r *http.Request, imageUtils utils.ImageUtils, rclone utils.Rclone, domainConfig config.DomainConfigManager) {
	if r.Method != http.MethodGet {
		utils.WriteInvalidRequestError(w, "Method not allowed", r.Method)
		return
	}

	domain := utils.GetDomainFromRequest(r)
	path := strings.TrimPrefix(r.URL.Path, "/"+config.ApiVersion+"/download/")
	if path == "" {
		utils.WriteInvalidPathError(w, "Path is required")
		return
	}

	cfg, err := domainConfig.GetDomainConfig(domain)
	if err != nil {
		utils.WriteInvalidDomainError(w, domain)
		return
	}

	if cfg.Security.Mode != "" {
		if err := security.ValidateSignedURLFromConfig(path, r.URL.Query(), cfg.Security.Secrets, cfg.Security.ValidityWindow); err != nil {
			switch err {
			case security.ErrKeyNotFound:
				utils.WriteUnauthorizedError(w, "Invalid security key")
			case security.ErrExpiredURL:
				utils.WriteExpiredTokenError(w)
			default:
				utils.WriteInvalidSignatureError(w)
			}
			return
		}
	}

	files, err := rclone.ListPath(path, domain)
	if err != nil {
		utils.WriteInternalError(w, "Failed to list files", err.Error())
		return
	}

	// If no files found
	if len(files) == 0 {
		utils.WriteNotFoundError(w, "File or directory not found", path)
		return
	}

	// Handle single file download
	if len(files) == 1 && !files[0].IsDir {
		handleSingleFileDownload(w, r, path, domain, imageUtils, rclone)
		return
	}

	// Check total size before processing folder download
	var totalSize int64
	for _, file := range files {
		if !file.IsDir {
			totalSize += file.Size
		}
	}

	const maxSize = 1 * 1024 * 1024 * 1024 // 1GB
	if totalSize > maxSize {
		utils.WriteInvalidRequestError(w, "Requested files exceed maximum allowed size", fmt.Sprintf("Size: %d bytes, Max: %d bytes", totalSize, maxSize))
		return
	}

	handleFolderDownload(w, r, path, files, domain, imageUtils, rclone)
}

func handleSingleFileDownload(w http.ResponseWriter, r *http.Request, path string, domain string, imageUtils utils.ImageUtils, rclone utils.Rclone) {
	content, err := rclone.FetchImage(path, domain)
	if err != nil {
		utils.WriteNotFoundError(w, "Failed to fetch file", err.Error())
		return
	}

	// Process image if it's an image file and has transformation parameters
	if utils.IsImageFile(path) && utils.HasImageTransformParams(r) {
		options := utils.ParseImageOptionsFromRequest(r)
		content, err = imageUtils.TransformImage(content, options)
		if err != nil {
			utils.WriteInternalError(w, "Failed to transform image", err.Error())
			return
		}
	}

	mimeType := http.DetectContentType(content)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(path)+"\"")
	w.Write(content)
}

func handleFolderDownload(w http.ResponseWriter, r *http.Request, path string, files []utils.RcloneFile, domain string, imageUtils utils.ImageUtils, rclone utils.Rclone) {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(path)+".zip\"")

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	options := utils.ParseImageOptionsFromRequest(r)
	hasTransformParams := utils.HasImageTransformParams(r)

	// Create a channel to receive processed files
	type processedFile struct {
		name    string
		content []byte
		err     error
	}
	results := make(chan processedFile)

	// Create a worker pool to limit concurrent operations
	const maxWorkers = 5
	sem := make(chan struct{}, maxWorkers)
	var activeWorkers int32

	// Start workers for each file
	for _, file := range files {
		if file.IsDir {
			continue
		}

		go func(f utils.RcloneFile) {
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			filePath := filepath.Join(path, f.Name)
			content, err := rclone.FetchImage(filePath, domain)
			if err != nil {
				results <- processedFile{name: f.Name, err: err}
				return
			}

			// Process image if needed
			if hasTransformParams && utils.IsImageFile(filePath) {
				content, err = imageUtils.TransformImage(content, options)
				if err != nil {
					results <- processedFile{name: f.Name, err: err}
					return
				}
			}

			results <- processedFile{name: f.Name, content: content}
		}(file)

		atomic.AddInt32(&activeWorkers, 1)
	}

	// Process results as they come in
	var processedCount int32
	for atomic.LoadInt32(&processedCount) < atomic.LoadInt32(&activeWorkers) {
		result := <-results
		atomic.AddInt32(&processedCount, 1)

		if result.err != nil {
			utils.Debug("Failed to process file", "error", result.err, "file", result.name)
			continue
		}

		f, err := zipWriter.Create(result.name)
		if err != nil {
			utils.Debug("Failed to create zip entry", "error", err, "file", result.name)
			continue
		}

		if _, err := f.Write(result.content); err != nil {
			utils.Debug("Failed to write to zip", "error", err, "file", result.name)
			continue
		}
	}
}
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"shuto-api/config"
	"shuto-api/utils"
)

// Utils defines the methods that can be used by the ListHandler
type Utils interface {
	ListPath(path string) ([]utils.RcloneFile, error)
}

type FileResponse struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
	IsDir    bool   `json:"isDir"`
	Width    int  `json:"width"`
	Height   int  `json:"height"`
}

// ListHandler processes listing of files based on the provided path
func ListHandler(w http.ResponseWriter, r *http.Request, imgUtils utils.ImageUtils, rclone utils.Rclone) {
	domain := utils.GetDomainFromRequest(r)
	path := strings.TrimPrefix(r.URL.Path, "/"+config.ApiVersion+"/list/")
	
	utils.Debug("Processing list request", 
		"domain", domain,
		"path", path,
		"remoteAddr", r.RemoteAddr,
	)

	files, err := rclone.ListPath(path, domain)
	if err != nil {
		utils.Error("Failed to list directory contents",
			"error", err,
			"path", path,
			"domain", domain,
		)
		http.Error(w, fmt.Sprintf("Failed to list directory contents: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response with image dimensions
	response := make([]FileResponse, len(files))
	for i, file := range files {
		newFile := FileResponse{
			Path:     file.Path,
			Size:     file.Size,
			MimeType: file.MimeType,
			IsDir:    file.IsDir,
		}

		// Get image dimensions if the file is an image
		if !file.IsDir && strings.HasPrefix(file.MimeType, "image/") {
			imgPath := path
			if !strings.HasSuffix(path, "/" + file.Path) {
				imgPath = path + "/" + file.Path
			}
			imgData, err := rclone.FetchImage(imgPath, domain)
			if err == nil {
				width, height, err := imgUtils.GetImageDimensions(imgData)
				if err == nil {
					newFile.Width = width
					newFile.Height = height
				}
			}
		}

		response[i] = newFile
	}

	data, err := json.Marshal(response)
	if err != nil {
		utils.Error("Failed to encode JSON response",
			"error", err,
			"filesCount", len(files),
		)
		http.Error(w, fmt.Sprintf("Failed to encode json: %v", err), http.StatusInternalServerError)
		return
	}

	utils.Debug("Successfully listed directory",
		"path", path,
		"filesCount", len(files),
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

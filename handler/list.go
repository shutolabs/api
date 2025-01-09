package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	Width    int  `json:"width,omitempty"`
	Height   int  `json:"height,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
}

// At package level
var (
	metadataCache *utils.Cache[utils.ImageMetadata]
)

func init() {
	var err error
	metadataCache, err = utils.NewCache[utils.ImageMetadata](utils.CacheOptions{
		MaxSize: 1000,
	})
	if err != nil {
		panic(err)
	}
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

			metadata, err := metadataCache.GetCached(utils.GetCachedOptions{
				Key: imgPath,
				TTL: 180 * 24 * time.Hour,
				StaleTime: 1 * time.Hour,
				GetFreshValue: func() (interface{}, error) {
					imgData, err := rclone.FetchImage(imgPath, domain)
					if err != nil {
						return utils.ImageMetadata{}, err
					}
					metadata, err := imgUtils.GetImageMetadata(imgData)
					if err != nil {
						return nil, err
					}
					return metadata, nil
				},
			})

			if err == nil {
				newFile.Width = metadata.Width
				newFile.Height = metadata.Height
				newFile.Keywords = metadata.Keywords
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

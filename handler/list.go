package handler

import (
	"encoding/json"
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
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	MimeType string    `json:"mimeType"`
	IsDir    bool      `json:"isDir"`
	Width    int       `json:"width,omitempty"`
	Height   int       `json:"height,omitempty"`
	Keywords []string  `json:"keywords,omitempty"`
}

var metadataCache *utils.Cache[utils.ImageMetadata]

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
	
	utils.Debug("Processing list request", "domain", domain, "path", path)

	files, err := rclone.ListPath(path, domain)
	if err != nil {
		utils.Error("Failed to list directory", "error", err, "path", path)
		http.Error(w, "Failed to list directory", http.StatusInternalServerError)
		return
	}

	response := make([]FileResponse, len(files))
	for i, file := range files {
		newFile := FileResponse{
			Path:     file.Path,
			Size:     file.Size,
			MimeType: file.MimeType,
			IsDir:    file.IsDir,
		}

		if !file.IsDir && strings.HasPrefix(file.MimeType, "image/") {
			imgPath := path
			if !strings.HasSuffix(path, "/" + file.Path) {
				imgPath = path + "/" + file.Path
			}

			metadata, err := metadataCache.GetCached(utils.GetCachedOptions{
				Key: imgPath,
				TTL: 24 * time.Hour,
				StaleTime: time.Hour,
				GetFreshValue: func() (interface{}, error) {
					imgData, err := rclone.FetchImage(imgPath, domain)
					if err != nil {
						return utils.ImageMetadata{}, err
					}
					return imgUtils.GetImageMetadata(imgData)
				},
			})

			if err == nil {
				newFile.Width = metadata.Width
				newFile.Height = metadata.Height
				newFile.Keywords = metadata.Keywords
			} else {
				utils.Debug("Failed to get image metadata", "error", err, "path", imgPath)
			}
		}

		response[i] = newFile
	}

	data, err := json.Marshal(response)
	if err != nil {
		utils.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	utils.Debug("Directory listed successfully", "path", path, "count", len(files))
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

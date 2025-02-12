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

// ListHandler handles directory listing requests
// @Summary List contents of a directory
// @Description Get a list of files and directories at the specified path
// @Tags list
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param   path     path    string     true        "Path to list contents from"
// @Success 200 {array}  utils.RcloneFile "List of files and directories"
// @Failure 400 {object} utils.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized - Invalid or missing API key"
// @Failure 404 {object} utils.ErrorResponse "Path not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /list/{path} [get]
func ListHandler(w http.ResponseWriter, r *http.Request, imgUtils utils.ImageUtils, rclone utils.Rclone, domainConfig config.DomainConfigManager) {
	domain := utils.GetDomainFromRequest(r)
	path := strings.TrimPrefix(r.URL.Path, "/"+config.ApiVersion+"/list/")
	
	if path == "" {
		utils.WriteInvalidPathError(w, "Path is required")
		return
	}

	utils.Debug("Processing list request", "domain", domain, "path", path)

	// Get domain configuration
	cfg, err := domainConfig.GetDomainConfig(domain)
	if err != nil {
		utils.WriteInvalidDomainError(w, domain)
		return
	}

	if !validateAPIKey(cfg.Security.APIKeys, r.Header.Get("Authorization")) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized\n"))
		return
	}

	files, err := rclone.ListPath(path, domain)
	if err != nil {
		utils.WriteInternalError(w, "Failed to list directory", err.Error())
		return
	}

	if len(files) == 0 {
		utils.WriteNotFoundError(w, "Directory is empty or does not exist", path)
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
		utils.WriteInternalError(w, "Failed to encode response", err.Error())
		return
	}

	utils.Debug("Directory listed successfully", "path", path, "count", len(files))
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func validateAPIKey(apiKeys []config.APIKey, authHeader string) bool {
	if len(apiKeys) == 0 {
		return true // No API keys configured means no authentication required
	}

	if authHeader == "" {
		return false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return false
	}

	providedKey := parts[1]
	for _, apiKey := range apiKeys {
		if apiKey.Key == providedKey {
			return true
		}
	}
	return false
}


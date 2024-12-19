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

// ListHandler processes listing of files based on the provided path
func ListHandler(w http.ResponseWriter, r *http.Request, rclone utils.Rclone) {
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

	data, err := json.Marshal(files)
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

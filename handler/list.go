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
	// Get domain configuration
	domain := utils.GetDomainFromRequest(r)

	// Extract path and parameters
	path := strings.TrimPrefix(r.URL.Path, "/"+config.ApiVersion+"/list/")

	// Fetch file data using the injected utils
	files, err := rclone.ListPath(path, domain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list directory contents: %v", err), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(files)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode json: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

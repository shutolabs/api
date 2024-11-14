package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"shuto-api/utils"
)

// ImageHandler processes image transformations based on query parameters
func ListHandler(w http.ResponseWriter, r *http.Request) {
	// Extract path and parameters
	path := strings.TrimPrefix(r.URL.Path, "/list/")

	// Fetch image data using rclone
	files, err := utils.ListPath(path)
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

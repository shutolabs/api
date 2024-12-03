package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// RcloneFile represents a file in Rclone
type RcloneFile struct {
	Path     string `json:"Path"`
	Name     string `json:"Name"`
	Size     int64  `json:"Size"`
	MimeType string `json:"MimeType"`
	ModTime  string `json:"ModTime"`
	IsDir    bool   `json:"IsDir"`
}

// Rclone interface for Rclone operations
type Rclone interface {
	FetchImage(path string) ([]byte, error)
	ListPath(path string) ([]RcloneFile, error)
}

// rcloneImpl is the concrete implementation of Rclone
type rcloneImpl struct {
	executor CommandExecutor
}

// NewRclone creates a new instance of rcloneImpl
func NewRclone(executor CommandExecutor) *rcloneImpl {
	return &rcloneImpl{executor: executor}
}

func (r *rcloneImpl) FetchImage(path string) ([]byte, error) {
	_, runningLocally := os.LookupEnv("RUNNING_LOCALLY")
	remote := "server:"
	if runningLocally {
		remote = "test:/"
	}
	output, err := r.executor.Execute("rclone", "cat", remote+path)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image with rclone: %v", err)
	}
	return output, nil
}

func (r *rcloneImpl) ListPath(path string) ([]RcloneFile, error) {
	_, runningLocally := os.LookupEnv("RUNNING_LOCALLY")
	remote := "server:"
	if runningLocally {
		remote = "test:/"
	}
	log.Println(remote + path)
	output, err := r.executor.Execute("rclone", "lsjson", remote+path)
	if err != nil {
		return nil, fmt.Errorf("error executing rclone lsjson: %v", err)
	}

	var files []RcloneFile
	if err := json.Unmarshal(output, &files); err != nil {
		return nil, fmt.Errorf("error parsing JSON output: %v", err)
	}
	return files, nil
}
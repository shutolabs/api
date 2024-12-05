package utils

import (
	"encoding/json"
	"fmt"
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

type RcloneConfig struct {
	Flags []string
	Type string
	Remote string
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

// GetRcloneConfig retrieves the configuration for rclone
func GetRcloneConfig() (*RcloneConfig, error) {
	config := &RcloneConfig{
		Type: "webdav",
		Remote: "webdav",
		Flags: []string{
			"--webdav-vendor=" + os.Getenv("RCLONE_CONFIG_SERVER_VENDOR"),
			"--webdav-url=" + os.Getenv("RCLONE_CONFIG_SERVER_URL"),
			"--webdav-user=" + os.Getenv("RCLONE_CONFIG_SERVER_USER"),
			"--webdav-pass=" + os.Getenv("RCLONE_CONFIG_SERVER_PASS"),
		},
	}
	return config, nil
}

// rcloneCmd executes an rclone command with the given configuration
func (r *rcloneImpl) rcloneCmd(command string, path string) ([]byte, error) {
	config, err := GetRcloneConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get rclone config: %v", err)
	}

	args := append([]string{command, config.Remote + ":" + path}, config.Flags...)
	return r.executor.Execute("rclone", args...)
}

func (r *rcloneImpl) FetchImage(path string) ([]byte, error) {
	output, err := r.rcloneCmd("lsjson", path)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image with rclone: %v", err)
	}
	return output, nil
}

func (r *rcloneImpl) ListPath(path string) ([]RcloneFile, error) {
	output, err := r.rcloneCmd("lsjson", path)
	if err != nil {
		return nil, fmt.Errorf("error executing rclone lsjson: %v", err)
	}

	var files []RcloneFile
	if err := json.Unmarshal(output, &files); err != nil {
		return nil, fmt.Errorf("error parsing JSON output: %v", err)
	}
	return files, nil
}
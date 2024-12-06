package utils

import (
	"encoding/json"
	"fmt"

	"shuto-api/config"
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

// Define the interface first
type Rclone interface {
	FetchImage(path string, domain string) ([]byte, error)
	ListPath(path string, domain string) ([]RcloneFile, error)
}

// MockRclone implements Rclone interface
type MockRclone struct {
	FetchImageFunc func(path string, domain string) ([]byte, error)
	ListPathFunc   func(path string, domain string) ([]RcloneFile, error)
}

// Implement the interface methods
func (m *MockRclone) FetchImage(path string, domain string) ([]byte, error) {
	return m.FetchImageFunc(path, domain)
}

func (m *MockRclone) ListPath(path string, domain string) ([]RcloneFile, error) {
	return m.ListPathFunc(path, domain)
}

type rcloneImpl struct {
	executor      CommandExecutor
	configManager config.DomainConfigManager
}

// NewRclone creates a new instance of rcloneImpl
func NewRclone(executor CommandExecutor, configManager config.DomainConfigManager) *rcloneImpl {
	return &rcloneImpl{
		executor:      executor,
		configManager: configManager,
	}
}

// GetRcloneConfig is now a method of rcloneImpl
func (r *rcloneImpl) getRcloneConfig(domain string) (*config.RcloneConfig, error) {
	domainConfig, err := r.configManager.GetDomainConfig(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain config: %v", err)
	}
	return &domainConfig.Rclone, nil
}

// rcloneCmd now uses the instance method
func (r *rcloneImpl) rcloneCmd(command string, path string, domain string) ([]byte, error) {
	config, err := r.getRcloneConfig(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get rclone config: %v", err)
	}

	args := append([]string{command, config.Remote + ":" + path}, config.Flags...)
	return r.executor.Execute("rclone", args...)
}

func (r *rcloneImpl) FetchImage(path string, domain string) ([]byte, error) {
	output, err := r.rcloneCmd("cat", path, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image with rclone: %v", err)
	}
	return output, nil
}

func (r *rcloneImpl) ListPath(path string, domain string) ([]RcloneFile, error) {
	fmt.Println("Listing path:", path, "for domain:", domain)
	output, err := r.rcloneCmd("lsjson", path, domain)
	if err != nil {
		return nil, fmt.Errorf("error executing rclone lsjson: %v", err)
	}

	var files []RcloneFile
	if err := json.Unmarshal(output, &files); err != nil {
		return nil, fmt.Errorf("error parsing JSON output: %v", err)
	}
	return files, nil
}
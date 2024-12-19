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
	Debug("Getting rclone configuration", "domain", domain)
	domainConfig, err := r.configManager.GetDomainConfig(domain)
	if err != nil {
		Error("Failed to get domain config for %s: %v", domain, err)
		return nil, fmt.Errorf("failed to get domain config: %v", err)
	}
	return &domainConfig.Rclone, nil
}

// rcloneCmd now uses the instance method
func (r *rcloneImpl) rcloneCmd(command string, path string, domain string) ([]byte, error) {
	Debug("Executing rclone command", "command", command, "path", path, "domain", domain)
	
	config, err := r.getRcloneConfig(domain)
	if err != nil {
		Error("Failed to get rclone config: %v", err)
		return nil, fmt.Errorf("failed to get rclone config: %v", err)
	}

	args := append([]string{command, config.Remote + ":" + path}, config.Flags...)
	Debug("Executing rclone", "args", args)
	
	output, err := r.executor.Execute("rclone", args...)
	if err != nil {
		Error("Rclone command failed",
			"error", err)
		return nil, err
	}
	
	Debug("Rclone command completed successfully", "outputSize", len(output))
	return output, nil
}

func (r *rcloneImpl) FetchImage(path string, domain string) ([]byte, error) {
	Debug("Fetching image with rclone",
		"path", path,
		"domain", domain,
	)
	
	output, err := r.rcloneCmd("cat", path, domain)
	if err != nil {
		Error("Failed to fetch image with rclone",
			"error", err,
			"path", path,
			"domain", domain,
		)
		return nil, fmt.Errorf("failed to fetch image with rclone: %w", err)
	}
	
	Debug("Successfully fetched image",
		"path", path,
		"size", len(output),
	)
	return output, nil
}

func (r *rcloneImpl) ListPath(path string, domain string) ([]RcloneFile, error) {
	Debug("Listing path with rclone",
		"path", path,
		"domain", domain,
	)
	
	output, err := r.rcloneCmd("lsjson", path, domain)
	if err != nil {
		Error("Failed to execute rclone lsjson",
			"error", err,
			"path", path,
			"domain", domain,
		)
		return nil, fmt.Errorf("error executing rclone lsjson: %w", err)
	}

	var files []RcloneFile
	if err := json.Unmarshal(output, &files); err != nil {
		Error("Failed to parse JSON output from rclone",
			"error", err,
			"output", string(output),
		)
		return nil, fmt.Errorf("error parsing JSON output: %v", err)
	}

	Debug("Successfully listed path",
		"path", path,
		"filesCount", len(files),
	)
	return files, nil
}
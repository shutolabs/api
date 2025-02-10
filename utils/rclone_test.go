package utils

import (
	"fmt"
	"testing"

	"shuto-api/config"

	"github.com/stretchr/testify/assert"
)

func TestFetchImage(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		ExecuteFunc: func(command string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("mock error")
		},
	}

	mockConfigManager := &config.MockDomainConfigManager{
		GetDomainConfigFunc: func(domain string) (config.DomainConfig, error) {
			return config.DomainConfig{
				Rclone: config.RcloneConfig{
					Remote: "test",
					Flags:  []string{},
				},
			}, nil
		},
	}

	rclone := NewRclone(mockExecutor, mockConfigManager)

	imageData, err := rclone.FetchImage("mock/path", "test")
	assert.Error(t, err)
	assert.Equal(t, "failed to fetch image: rclone command failed: mock error", err.Error())
	assert.Nil(t, imageData)

	// Test config manager error
	mockConfigManager.GetDomainConfigFunc = func(domain string) (config.DomainConfig, error) {
		return config.DomainConfig{}, fmt.Errorf("config error")
	}

	imageData, err = rclone.FetchImage("mock/path", "test")
	assert.Error(t, err)
	assert.NotNil(t, err)
	assert.Nil(t, imageData)
}

func TestListPath(t *testing.T) {
	// Test successful case
	mockExecutor := &MockCommandExecutor{
		ExecuteFunc: func(command string, args ...string) ([]byte, error) {
			return []byte(`[
				{"Path":"file1.jpg","Name":"file1.jpg","Size":1024,"MimeType":"image/jpeg"},
				{"Path":"file2.png","Name":"file2.png","Size":2048,"MimeType":"image/png"},
				{"Path":"file3.gif","Name":"file3.gif","Size":512,"MimeType":"image/gif"}
			]`), nil
		},
	}

	mockConfigManager := &config.MockDomainConfigManager{
		GetDomainConfigFunc: func(domain string) (config.DomainConfig, error) {
			return config.DomainConfig{
				Rclone: config.RcloneConfig{
					Remote: "test",
					Flags:  []string{},
				},
			}, nil
		},
	}

	rclone := NewRclone(mockExecutor, mockConfigManager)

	// Test successful listing
	files, err := rclone.ListPath("mock/path", "test")
	assert.NoError(t, err)
	expectedFiles := []RcloneFile{
		{Path: "file1.jpg", Name: "file1.jpg", Size: 1024, MimeType: "image/jpeg"},
		{Path: "file2.png", Name: "file2.png", Size: 2048, MimeType: "image/png"},
		{Path: "file3.gif", Name: "file3.gif", Size: 512, MimeType: "image/gif"},
	}
	assert.Equal(t, expectedFiles, files)

	// Test executor error
	mockExecutor.ExecuteFunc = func(command string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("mock error")
	}

	files, err = rclone.ListPath("mock/path", "test")
	assert.Error(t, err)
	assert.Equal(t, "failed to list path: rclone command failed: mock error", err.Error())
	assert.Nil(t, files)

	// Test config manager error
	mockConfigManager.GetDomainConfigFunc = func(domain string) (config.DomainConfig, error) {
		return config.DomainConfig{}, fmt.Errorf("config error")
	}

	files, err = rclone.ListPath("mock/path", "test")
	assert.Error(t, err)
	assert.NotNil(t, err)
	assert.Nil(t, files)
}



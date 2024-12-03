package utils

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestFetchImage(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		ExecuteFunc: func(command string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("mock error")
		},
	}

	rclone := NewRclone(mockExecutor)

	imageData, err := rclone.FetchImage("mock/path")
	assert.Error(t, err)
	assert.Equal(t, "failed to fetch image with rclone: mock error", err.Error())
	assert.Nil(t, imageData)

	// Running locally case
	os.Setenv("RUNNING_LOCALLY", "true")
	defer os.Unsetenv("RUNNING_LOCALLY") // Clean up after test

	mockExecutor.ExecuteFunc = func(command string, args ...string) ([]byte, error) {
		return []byte(`mock image data`), nil
	}

	imageData, err = rclone.FetchImage("mock/path")
	assert.Nil(t, err)
	assert.Equal(t, []byte(`mock image data`), imageData)
	assert.NotNil(t, imageData)

}

func TestListPath(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		ExecuteFunc: func(command string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("mock error")
		},
	}
	
	rclone := NewRclone(mockExecutor)
	
	// Error case: Command execution error
	listData, err := rclone.ListPath("mock/path")
	assert.Error(t, err)
	assert.Nil(t, listData)

	// Error case: JSON decoding error
	mockExecutor.ExecuteFunc = func(command string, args ...string) ([]byte, error) {
		return []byte(`invalid json`), nil // Simulate invalid JSON
	}

	listData, err = rclone.ListPath("mock/path")
	assert.Error(t, err)
	assert.Nil(t, listData)
	
	// Successful case
	mockExecutor.ExecuteFunc = func(command string, args ...string) ([]byte, error) {
		return []byte(`[{"Path":"mock/path/file1.txt","Name":"file1.txt","Size":1234,"MimeType":"text/plain","ModTime":"2023-01-01T00:00:00Z","IsDir":false}]`), nil
	}
	listData, err = rclone.ListPath("mock/path")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listData))
	assert.Equal(t, "mock/path/file1.txt", listData[0].Path)
	assert.Equal(t, "file1.txt", listData[0].Name)
	assert.Equal(t, int64(1234), listData[0].Size)
	assert.Equal(t, "text/plain", listData[0].MimeType)
	assert.Equal(t, "2023-01-01T00:00:00Z", listData[0].ModTime)
	assert.Equal(t, false, listData[0].IsDir)

	// Running locally case
	os.Setenv("RUNNING_LOCALLY", "true")
	defer os.Unsetenv("RUNNING_LOCALLY") // Clean up after test

	listData, err = rclone.ListPath("mock/path")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listData))
	assert.Equal(t, "mock/path/file1.txt", listData[0].Path)
	assert.Equal(t, "file1.txt", listData[0].Name)
	assert.Equal(t, int64(1234), listData[0].Size)
	assert.Equal(t, "text/plain", listData[0].MimeType)
	assert.Equal(t, "2023-01-01T00:00:00Z", listData[0].ModTime)
	assert.Equal(t, false, listData[0].IsDir)
}



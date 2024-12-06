package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"shuto-api/config"
	"shuto-api/utils"
)

// MockUtils is a mock implementation of the Utils interface for testing
type MockUtils struct {
	Files []utils.RcloneFile
	Err   error
}

func (m *MockUtils) ListPath(path string) ([]utils.RcloneFile, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	// Test path handling by returning different results for different paths
	if path == "empty" {
		return []utils.RcloneFile{}, nil
	}
	return m.Files, nil
}

func TestListHandler(t *testing.T) {
	// Define a sample time string for consistency
	sampleTimeStr := "2024-01-01T00:00:00Z"
	
	tests := []struct {
		name           string
		path           string
		mockFiles      []byte
		mockListError  error
		expectedStatus int
		expectedBody   string
		checkBody      func(t *testing.T, body []byte)
	}{
		{
			name: "Successful listing with multiple files",
			path: "http://test/v1/list/photos",
			mockFiles: []byte(`[
				{
					"path": "photos/file1.jpg",
					"name": "file1.jpg",
					"size": 1024,
					"mimeType": "image/jpeg",
					"modTime": "` + sampleTimeStr + `",
					"isDir": false
				},
				{
					"path": "photos/dir1",
					"name": "dir1",
					"size": 0,
					"mimeType": "",
					"modTime": "` + sampleTimeStr + `",
					"isDir": true
				}
			]`),
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var files []utils.RcloneFile
				err := json.Unmarshal(body, &files)
				if err != nil {
					t.Fatalf("Failed to unmarshal response body: %v", err)
				}
				if len(files) != 2 {
					t.Errorf("Expected 2 files, got %d", len(files))
				}
				if files[0].Name != "file1.jpg" || files[1].Name != "dir1" {
					t.Errorf("Unexpected file names in response")
				}
			},
		},
		{
			name:           "Empty directory",
			path:           "http://test/v1/list/empty",
			mockFiles:      []byte(`[]`),
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var files []utils.RcloneFile
				err := json.Unmarshal(body, &files)
				if err != nil {
					t.Fatalf("Failed to unmarshal response body: %v", err)
				}
				if len(files) != 0 {
					t.Errorf("Expected empty file list, got %d files", len(files))
				}
			},
		},
		{
			name:           "Error listing files",
			path:           "http://test/v1/list/error",
			mockListError:  fmt.Errorf("error executing rclone lsjson: failed to list directory"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to list directory contents: error executing rclone lsjson: error executing rclone lsjson: failed to list directory\n",
		},
		{
			name: "Root path listing",
			path: "http://test/v1/list/123",
			mockFiles: []byte(`[
				{
					"path": "file1.jpg",
					"name": "file1.jpg",
					"size": 1024,
					"mimeType": "image/jpeg",
					"modTime": "` + sampleTimeStr + `",
					"isDir": false
				}
			]`),
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var files []utils.RcloneFile
				err := json.Unmarshal(body, &files)
				if err != nil {
					t.Fatalf("Failed to unmarshal response body: %v", err)
				}
				if len(files) != 1 {
					t.Errorf("Expected 1 file, got %d", len(files))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &utils.MockCommandExecutor{
				ExecuteFunc: func(command string, args ...string) ([]byte, error) {
					return tt.mockFiles, tt.mockListError
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

			rclone := utils.NewRclone(mockExecutor, mockConfigManager)

			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			ListHandler(rec, req, rclone)

			// Check status code
			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			// Check Content-Type header for successful responses
			if tt.expectedStatus == http.StatusOK {
				contentType := rec.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", contentType)
				}
			}

			// Check body
			if tt.expectedBody != "" {
				if rec.Body.String() != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, rec.Body.String())
				}
			}

			// Run custom body checks if provided
			if tt.checkBody != nil {
				tt.checkBody(t, rec.Body.Bytes())
			}
		})
	}
} 
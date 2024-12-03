package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	// Create a sample time for consistent testing
	sampleTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	sampleTimeStr := sampleTime.Format(time.RFC3339)

	tests := []struct {
		name           string
		path           string
		mockFiles      []utils.RcloneFile
		mockListError  error
		expectedStatus int
		expectedBody   string
		checkBody      func(t *testing.T, body []byte)
	}{
		{
			name: "Successful listing with multiple files",
			path: "/list/photos",
			mockFiles: []utils.RcloneFile{
				{
					Path:     "photos/file1.jpg",
					Name:     "file1.jpg",
					Size:     1024,
					MimeType: "image/jpeg",
					ModTime:  sampleTimeStr,
					IsDir:    false,
				},
				{
					Path:     "photos/dir1",
					Name:     "dir1",
					Size:     0,
					MimeType: "",
					ModTime:  sampleTimeStr,
					IsDir:    true,
				},
			},
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
			path:           "/list/empty",
			mockFiles:      []utils.RcloneFile{},
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
			path:           "/list/error",
			mockListError:  fmt.Errorf("failed to list directory"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to list directory contents: failed to list directory\n",
		},
		{
			name: "Root path listing",
			path: "/list/",
			mockFiles: []utils.RcloneFile{
				{
					Path:     "file1.jpg",
					Name:     "file1.jpg",
					Size:     1024,
					MimeType: "image/jpeg",
					ModTime:  sampleTimeStr,
					IsDir:    false,
				},
			},
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
			mockUtils := &MockUtils{
				Files: tt.mockFiles,
				Err:   tt.mockListError,
			}

			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			ListHandler(rec, req, mockUtils)

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
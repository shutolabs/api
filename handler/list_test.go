package handler

import (
	"encoding/json"
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
	tests := []struct {
		name           string
		path           string
		authHeader     string
		mockFiles      []utils.RcloneFile
		mockListError  error
		mockDomainConfig config.DomainConfig
		mockDomainConfigError error
		expectedStatus int
		expectedBody   string
		checkBody      func(t *testing.T, body []byte)
	}{
		{
			name: "Successful listing with no API key required",
			path: "/v1/list/photos",
			mockFiles: []utils.RcloneFile{
				{Path: "photos/file1.jpg", Size: 1024, MimeType: "image/jpeg", IsDir: false},
				{Path: "photos/dir1", Size: 0, MimeType: "", IsDir: true},
			},
			mockDomainConfig: config.DomainConfig{},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var files []FileResponse
				err := json.Unmarshal(body, &files)
				if err != nil {
					t.Fatalf("Failed to unmarshal response body: %v", err)
				}
				if len(files) != 2 {
					t.Errorf("Expected 2 files, got %d", len(files))
				}
			},
		},
		{
			name: "Successful listing with valid API key",
			path: "/v1/list/photos",
			authHeader: "Bearer test-key-1",
			mockFiles: []utils.RcloneFile{
				{Path: "photos/file1.jpg", Size: 1024, MimeType: "image/jpeg", IsDir: false},
			},
			mockDomainConfig: config.DomainConfig{
				Security: config.SecuritySettings{
					APIKeys: []config.APIKey{
						{Key: "test-key-1", Description: "Test Key 1"},
						{Key: "test-key-2", Description: "Test Key 2"},
					},
				},
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var files []FileResponse
				err := json.Unmarshal(body, &files)
				if err != nil {
					t.Fatalf("Failed to unmarshal response body: %v", err)
				}
				if len(files) != 1 {
					t.Errorf("Expected 1 file, got %d", len(files))
				}
			},
		},
		{
			name: "Failed authentication - missing API key",
			path: "/v1/list/photos",
			mockDomainConfig: config.DomainConfig{
				Security: config.SecuritySettings{
					APIKeys: []config.APIKey{
						{Key: "test-key-1", Description: "Test Key 1"},
					},
				},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: "Unauthorized\n",
		},
		{
			name: "Failed authentication - invalid API key",
			path: "/v1/list/photos",
			authHeader: "Bearer invalid-key",
			mockDomainConfig: config.DomainConfig{
				Security: config.SecuritySettings{
					APIKeys: []config.APIKey{
						{Key: "test-key-1", Description: "Test Key 1"},
					},
				},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: "Unauthorized\n",
		},
		{
			name: "Failed authentication - malformed auth header",
			path: "/v1/list/photos",
			authHeader: "Basic test-key-1",
			mockDomainConfig: config.DomainConfig{
				Security: config.SecuritySettings{
					APIKeys: []config.APIKey{
						{Key: "test-key-1", Description: "Test Key 1"},
					},
				},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: "Unauthorized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRclone := &utils.MockRclone{
				ListPathFunc: func(path string, domain string) ([]utils.RcloneFile, error) {
					if tt.mockListError != nil {
						return nil, tt.mockListError
					}
					return tt.mockFiles, nil
				},
				FetchImageFunc: func(path string, domain string) ([]byte, error) {
					// Return dummy image data for testing
					return []byte("mock-image-data"), nil
				},
			}

			mockDomainConfigManager := &config.MockDomainConfigManager{
				GetDomainConfigFunc: func(domain string) (config.DomainConfig, error) {
					if tt.mockDomainConfigError != nil {
						return config.DomainConfig{}, tt.mockDomainConfigError
					}
					return tt.mockDomainConfig, nil
				},
			}

			mockImageUtils := &MockImageUtils{
				GetImageMetadataFunc: func(data []byte) (utils.ImageMetadata, error) {
					return utils.ImageMetadata{Width: 100, Height: 100}, nil
				},
			}

			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			ListHandler(rec, req, mockImageUtils, mockRclone, mockDomainConfigManager)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedBody != "" {
				if rec.Body.String() != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, rec.Body.String())
				}
			}

			if tt.checkBody != nil {
				tt.checkBody(t, rec.Body.Bytes())
			}
		})
	}
} 
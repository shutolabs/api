package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"shuto-api/config"
	"shuto-api/utils"
)

func TestDownloadHandler(t *testing.T) {
	defaultDomainConfig := func(domain string) (config.DomainConfig, error) {
		return config.DomainConfig{}, nil
	}

	securedDomainConfig := func(domain string) (config.DomainConfig, error) {
		return config.DomainConfig{
			Security: config.SecuritySettings{
				Mode: config.HMACTimebound,
				Secrets: []config.SecretKey{
					{KeyID: "v1", Secret: "test-secret"},
				},
				ValidityWindow: 300,
			},
		}, nil
	}

	tests := []struct {
		name           string
		path           string
		queryParams    map[string]string
		mockFetch      func(string, string) ([]byte, error)
		mockList       func(string, string) ([]utils.RcloneFile, error)
		mockDomainConfig func(string) (config.DomainConfig, error)
		expectedStatus int
		expectedHeaders map[string]string
	}{
		{
			name: "Basic download without security",
			path: "/download/test.jpg",
			mockList: func(path, domain string) ([]utils.RcloneFile, error) {
				return []utils.RcloneFile{
					{Name: "test.jpg", Size: 1024, IsDir: false},
				}, nil
			},
			mockFetch: func(path, domain string) ([]byte, error) {
				return []byte("test-data"), nil
			},
			mockDomainConfig: defaultDomainConfig,
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Content-Type": "application/zip",
				"Content-Disposition": "attachment; filename=\"test.jpg.zip\"",
			},
		},
		{
			name: "Download with security - missing signature",
			path: "/download/test.jpg",
			mockList: func(path, domain string) ([]utils.RcloneFile, error) {
				return []utils.RcloneFile{}, nil
			},
			mockDomainConfig: securedDomainConfig,
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Download with security - expired URL",
			path: "/download/test.jpg",
			queryParams: map[string]string{
				"kid": "v1",
				"ts":  "1000", // Old timestamp
				"sig": "invalid",
			},
			mockList: func(path, domain string) ([]utils.RcloneFile, error) {
				return []utils.RcloneFile{}, nil
			},
			mockDomainConfig: securedDomainConfig,
			expectedStatus: http.StatusGone,
		},
		{
			name: "Download with security - invalid key",
			path: "/download/test.jpg",
			queryParams: map[string]string{
				"kid": "v2", // Non-existent key
				"ts":  "1000",
				"sig": "invalid",
			},
			mockList: func(path, domain string) ([]utils.RcloneFile, error) {
				return []utils.RcloneFile{}, nil
			},
			mockDomainConfig: securedDomainConfig,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Download size exceeds limit",
			path: "/download/large-folder",
			mockList: func(path, domain string) ([]utils.RcloneFile, error) {
				return []utils.RcloneFile{
					{Name: "large.file", Size: 2 * 1024 * 1024 * 1024, IsDir: false}, // 2GB
				}, nil
			},
			mockDomainConfig: defaultDomainConfig,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRclone := &utils.MockRclone{
				FetchImageFunc: tt.mockFetch,
				ListPathFunc:   tt.mockList,
			}

			mockImageUtils := &MockImageUtils{
				TransformImageFunc: func([]byte, utils.ImageTransformOptions) ([]byte, error) {
					return []byte("transformed"), nil
				},
				GetMimeTypeFunc: func([]byte) (string, error) {
					return "image/jpeg", nil
				},
			}

			mockDomainConfig := &MockDomainConfigManager{
				GetDomainConfigFunc: tt.mockDomainConfig,
			}

			req := httptest.NewRequest("GET", tt.path, nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			DownloadHandler(rr, req, mockImageUtils, mockRclone, mockDomainConfig)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				for k, v := range tt.expectedHeaders {
					if rr.Header().Get(k) != v {
						t.Errorf("expected %s header %s, got %s", k, v, rr.Header().Get(k))
					}
				}
			}
		})
	}
} 
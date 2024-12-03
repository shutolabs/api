package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"shuto-api/utils"
)

// MockImageUtils implements utils.ImageUtils interface for testing
type MockImageUtils struct {
	TransformImageFunc func([]byte, utils.ImageTransformOptions) ([]byte, error)
	GetMimeTypeFunc   func([]byte) (string, error)
}

func (m *MockImageUtils) TransformImage(data []byte, opts utils.ImageTransformOptions) ([]byte, error) {
	return m.TransformImageFunc(data, opts)
}

func (m *MockImageUtils) GetMimeType(data []byte) (string, error) {
	return m.GetMimeTypeFunc(data)
}

// MockRclone implements utils.Rclone interface for testing
type MockRclone struct {
	FetchImageFunc func(string) ([]byte, error)
	ListPathFunc   func(string) ([]utils.RcloneFile, error)
}

func (m *MockRclone) FetchImage(path string) ([]byte, error) {
	return m.FetchImageFunc(path)
}

func (m *MockRclone) ListPath(path string) ([]utils.RcloneFile, error) {
	return m.ListPathFunc(path)
}

func TestImageHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		queryParams    map[string]string
		mockFetch      func(string) ([]byte, error)
		mockTransform  func([]byte, utils.ImageTransformOptions) ([]byte, error)
		mockMimeType   func([]byte) (string, error)
		expectedStatus int
		expectedMime   string
	}{
		{
			name: "Successful image processing",
			path: "/image/test.jpg",
			queryParams: map[string]string{
				"w": "100", "h": "100", "format": "jpeg",
				"crop": "true", "fit": "cover", "quality": "80",
				"dpr": "2.0",
			},
			mockFetch: func(path string) ([]byte, error) {
				return []byte("mock-image-data"), nil
			},
			mockTransform: func(data []byte, opts utils.ImageTransformOptions) ([]byte, error) {
				return []byte("mock-transformed-image"), nil
			},
			mockMimeType: func(data []byte) (string, error) {
				return "image/jpeg", nil
			},
			expectedStatus: http.StatusOK,
			expectedMime:   "image/jpeg",
		},
		{
			name: "Failed to fetch image",
			path: "/image/nonexistent.jpg",
			mockFetch: func(path string) ([]byte, error) {
				return nil, fmt.Errorf("image not found")
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Failed to transform image",
			path: "/image/test.jpg",
			mockFetch: func(path string) ([]byte, error) {
				return []byte("mock-image-data"), nil
			},
			mockTransform: func(data []byte, opts utils.ImageTransformOptions) ([]byte, error) {
				return nil, fmt.Errorf("transform error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Failed to get mime type",
			path: "/image/test.jpg",
			mockFetch: func(path string) ([]byte, error) {
				return []byte("mock-image-data"), nil
			},
			mockTransform: func(data []byte, opts utils.ImageTransformOptions) ([]byte, error) {
				return []byte("mock-transformed-image"), nil
			},
			mockMimeType: func(data []byte) (string, error) {
				return "", fmt.Errorf("mime type error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Default DPR value",
			path: "/image/test.jpg",
			queryParams: map[string]string{
				"w": "100", "h": "100",
			},
			mockFetch: func(path string) ([]byte, error) {
				return []byte("mock-image-data"), nil
			},
			mockTransform: func(data []byte, opts utils.ImageTransformOptions) ([]byte, error) {
				if opts.Dpr != 1.0 {
					t.Errorf("expected default DPR 1.0, got %f", opts.Dpr)
				}
				return []byte("mock-transformed-image"), nil
			},
			mockMimeType: func(data []byte) (string, error) {
				return "image/jpeg", nil
			},
			expectedStatus: http.StatusOK,
			expectedMime:   "image/jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRclone := &MockRclone{
				FetchImageFunc: tt.mockFetch,
			}

			mockImageUtils := &MockImageUtils{
				TransformImageFunc: tt.mockTransform,
				GetMimeTypeFunc:   tt.mockMimeType,
			}

			req := httptest.NewRequest("GET", tt.path, nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			ImageHandler(rr, req, mockImageUtils, mockRclone)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				if contentType := rr.Header().Get("Content-Type"); contentType != tt.expectedMime {
					t.Errorf("expected Content-Type %s, got %s", tt.expectedMime, contentType)
				}

				if cacheControl := rr.Header().Get("Cache-Control"); cacheControl != "public, max-age=31536000" {
					t.Errorf("expected Cache-Control header not set correctly, got %s", cacheControl)
				}
			}
		})
	}
}

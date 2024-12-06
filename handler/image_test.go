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

func TestImageHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		queryParams    map[string]string
		mockFetch      func(string, string) ([]byte, error)
		mockTransform  func([]byte, utils.ImageTransformOptions) ([]byte, error)
		mockMimeType   func([]byte) (string, error)
		expectedStatus int
		expectedMime   string
		expectedHeaders map[string]string
	}{
		{
			name: "Basic resize with defaults",
			path: "/image/test.jpg",
			queryParams: map[string]string{
				"w": "100",
				"h": "100",
			},
			mockFetch: func(remote, path string) ([]byte, error) {
				return []byte("mock-image-data"), nil
			},
			mockTransform: func(data []byte, opts utils.ImageTransformOptions) ([]byte, error) {
				if opts.Width != 100 || opts.Height != 100 {
					t.Errorf("expected width 100 and height 100, got %d and %d", opts.Width, opts.Height)
				}
				return []byte("mock-transformed-image"), nil
			},
			mockMimeType: func(data []byte) (string, error) {
				return "image/jpeg", nil
			},
			expectedStatus: http.StatusOK,
			expectedMime:   "image/jpeg",
			expectedHeaders: map[string]string{
				"Cache-Control": "public, max-age=31536000",
			},
		},
		{
			name: "Full parameter test",
			path: "/image/test.jpg",
			queryParams: map[string]string{
				"w": "300",
				"h": "200",
				"fit": "crop",
				"fm": "webp",
				"q": "80",
				"dpr": "2",
				"blur": "15",
				"dl": "1",
			},
			mockFetch: func(remote, path string) ([]byte, error) {
				return []byte("mock-image-data"), nil
			},
			mockTransform: func(data []byte, opts utils.ImageTransformOptions) ([]byte, error) {
				if opts.Width != 300 || opts.Height != 200 || opts.Fit != "crop" ||
				   opts.Format != "webp" || opts.Quality != 80 || opts.Dpr != 2.0 ||
				   opts.Blur != 15 || !opts.ForceDownload {
					t.Error("Transform options not set correctly")
				}
				return []byte("mock-transformed-image"), nil
			},
			mockMimeType: func(data []byte) (string, error) {
				return "image/webp", nil
			},
			expectedStatus: http.StatusOK,
			expectedMime:   "image/webp",
			expectedHeaders: map[string]string{
				"Content-Disposition": "attachment",
				"Cache-Control": "public, max-age=31536000",
			},
		},
		{
			name: "Failed to fetch image",
			path: "/image/nonexistent.jpg",
			mockFetch: func(remote, path string) ([]byte, error) {
				return nil, fmt.Errorf("image not found")
			},
			mockTransform: func(data []byte, opts utils.ImageTransformOptions) ([]byte, error) {
				return nil, nil
			},
			mockMimeType: func(data []byte) (string, error) {
				return "", nil
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Failed to transform image",
			path: "/image/test.jpg",
			mockFetch: func(remote, path string) ([]byte, error) {
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
			mockFetch: func(remote, path string) ([]byte, error) {
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
			mockFetch: func(remote, path string) ([]byte, error) {
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
			expectedHeaders: map[string]string{
				"Cache-Control": "public, max-age=31536000",
			},
		},
		{
			name: "DPR value > 3",
			path: "/image/test.jpg",
			queryParams: map[string]string{
				"w": "100", "h": "100", "dpr": "3.1",
			},
			mockFetch: func(remote, path string) ([]byte, error) {
				return []byte("mock-image-data"), nil
			},
			mockTransform: func(data []byte, opts utils.ImageTransformOptions) ([]byte, error) {
				if opts.Dpr != 3.0 {
					t.Errorf("expected default DPR 3.0, got %f", opts.Dpr)
				}
				return []byte("mock-transformed-image"), nil
			},
			mockMimeType: func(data []byte) (string, error) {
				return "image/jpeg", nil
			},
			expectedStatus: http.StatusOK,
			expectedMime:   "image/jpeg",
			expectedHeaders: map[string]string{
				"Cache-Control": "public, max-age=31536000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRclone := &utils.MockRclone{
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

				for k, v := range tt.expectedHeaders {
					if rr.Header().Get(k) != v {
						t.Errorf("expected %s header %s, got %s", k, v, rr.Header().Get(k))
					}
				}
			}
		})
	}
}

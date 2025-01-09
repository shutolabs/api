package utils

import (
	"os"
	"testing"
)

func TestGetMimeType(t *testing.T) {
	// Create an instance of ImageUtils
	imageUtils := NewImageUtils()

	tests := []struct {
		name         string
		imageData    []byte
		expectedMime string
		expectError  bool
	}{
		{"JPEG Image", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43, 0x00, 0x03, 0x02, 0x02, 0x02, 0x02, 0x02, 0x03, 0x02, 0x02, 0x02, 0x03, 0x03, 0x03, 0x03, 0x04, 0x06, 0x04, 0x04, 0x04, 0x04, 0x04, 0x08, 0x06, 0x06, 0x05, 0x06, 0x09, 0x08, 0x0A, 0x0A, 0x09, 0x08, 0x09, 0x09, 0x0A, 0x0C, 0x0F, 0x0C, 0x0A, 0x0B, 0x0E, 0x0B, 0x09, 0x09, 0x0D, 0x11, 0x0D, 0x0E, 0x0F, 0x10, 0x10, 0x11, 0x10, 0x0A, 0x0C, 0x12, 0x13, 0x12, 0x10, 0x13, 0x0F, 0x10, 0x10, 0x10, 0xFF, 0xC9, 0x00, 0x0B, 0x08, 0x00, 0x01, 0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xFF, 0xCC, 0x00, 0x06, 0x00, 0x10, 0x10, 0x05, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3F, 0x00, 0xD2, 0xCF, 0x20, 0xFF, 0xD9}, "image/jpeg", false},
		{"PNG Image", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x37, 0x6E, 0xF9, 0x24, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x01, 0x63, 0x60, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01, 0x73, 0x75, 0x01, 0x18, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82}, "image/png", false},
		{"WEBP Image", []byte{0x52, 0x49, 0x46, 0x46, 0x1A, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50, 0x56, 0x50, 0x38, 0x4C, 0x0D, 0x00, 0x00, 0x00, 0x2F, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2F}, "image/webp", false},
		{"Unsupported Format", []byte{0x00, 0x00, 0x00}, "application/octet-stream", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mimeType, err := imageUtils.GetMimeType(tt.imageData)
			if (err != nil) != tt.expectError {
				t.Fatalf("GetMimeType() returned an error: %v, expected error: %v", err, tt.expectError)
			}

			if mimeType != tt.expectedMime && !tt.expectError {
				t.Errorf("GetMimeType() = %v, expected %v", mimeType, tt.expectedMime)
			}
		})
	}
}

// New test for TransformImage
func TestTransformImage(t *testing.T) {
	imageUtils := NewImageUtils()

	tests := []struct {
		name         string
		filePath     string
		opts         ImageTransformOptions
		expectError  bool
	}{
		{"Transform JPEG", "testdata/sample.jpeg", ImageTransformOptions{Width: 100, Height: 100, Format: "jpeg", Quality: 80, Fit: "clip", Dpr: 1}, false},
		{"Transform JPEG - 0 Width", "testdata/sample.jpeg", ImageTransformOptions{Width: 0, Height: 100, Format: "jpeg", Quality: 80, Fit: "clip", Dpr: 1}, false},
		{"Transform JPEG - 0 Height", "testdata/sample.jpeg", ImageTransformOptions{Width: 100, Height: 0, Format: "jpeg", Quality: 80, Fit: "clip", Dpr: 1}, false},
		{"Transform JPEG - 0 Height & 0 Width", "testdata/sample.jpeg", ImageTransformOptions{Width: 0, Height: 0, Format: "jpeg", Quality: 80, Fit: "clip", Dpr: 1}, false},
		{"Transform PNG", "testdata/sample.png", ImageTransformOptions{Width: 100, Height: 100, Format: "png", Quality: 80, Fit: "clip", Dpr: 1}, false},
		{"Transform WebP", "testdata/sample.webp", ImageTransformOptions{Width: 100, Height: 100, Format: "webp", Quality: 80, Fit: "clip", Dpr: 1}, false},
		{"Invalid Fit Option", "testdata/sample.jpeg", ImageTransformOptions{Width: 100, Height: 100, Format: "jpeg", Quality: 80, Fit: "invalid"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the image file
			imgData, err := os.ReadFile("../"+tt.filePath)
			if err != nil {
				t.Fatalf("failed to read image file: %v", err)
			}
			
			modifiedImg, err := imageUtils.TransformImage(imgData, tt.opts)
			if (err != nil) != tt.expectError {
				t.Fatalf("TransformImage() returned an error: %v, expected error: %v", err, tt.expectError)
			}

			if !tt.expectError && modifiedImg == nil {
				t.Errorf("TransformImage() returned nil image, expected non-nil")
			}
		})
	}
}

func TestGetImageDimensions(t *testing.T) {
	imageUtils := NewImageUtils()

	tests := []struct {
		name           string
		filePath      string
		expectedWidth  int
		expectedHeight int
		expectError    bool
	}{
		{"JPEG Dimensions", "testdata/sample.jpeg", 3000, 2000, false},
		{"PNG Dimensions", "testdata/sample.png", 3000, 2000, false},
		{"WebP Dimensions", "testdata/sample.webp", 3000, 2000, false},
		{"Invalid Image", "testdata/invalid.jpg", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the image file
			imgData, err := os.ReadFile("../" + tt.filePath)
			if err != nil && !tt.expectError {
				t.Fatalf("failed to read image file: %v", err)
			}

			metadata, err := imageUtils.GetImageMetadata(imgData)
			if (err != nil) != tt.expectError {
				t.Fatalf("GetImageDimensions() returned an error: %v, expected error: %v", err, tt.expectError)
			}

			if !tt.expectError {
				if metadata.Width != tt.expectedWidth {
					t.Errorf("GetImageDimensions() width = %v, expected %v", metadata.Width, tt.expectedWidth)
				}
				if metadata.Height != tt.expectedHeight {
					t.Errorf("GetImageDimensions() height = %v, expected %v", metadata.Height, tt.expectedHeight)
				}
			}
		})
	}
}

package utils

import (
	"fmt"

	"github.com/davidbyttow/govips/v2/vips"
)

// GetMimeType takes an image byte slice and returns its MIME type
func GetMimeType(imgData []byte) (string, error) {
	// Create a new image from the byte buffer
	image, err := vips.NewImageFromBuffer(imgData)
	if err != nil {
		return "", fmt.Errorf("failed to create image from buffer: %v", err)
	}
	defer image.Close()

	// Get the image format
	format := image.Format()

	// Map the format to the corresponding MIME type
	var mimeType string
	switch format {
	case vips.ImageTypeGIF:
		mimeType = "image/gif"
	case vips.ImageTypeJPEG:
		mimeType = "image/jpeg"
	case vips.ImageTypePNG:
		mimeType = "image/png"
	case vips.ImageTypeWEBP:
		mimeType = "image/webp"
	case vips.ImageTypeTIFF:
		mimeType = "image/tiff"
	default:
		mimeType = "application/octet-stream" // Fallback for unknown formats
	}

	return mimeType, nil
}

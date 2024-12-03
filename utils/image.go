package utils

import (
	"errors"
	"fmt"
	"math"

	"github.com/davidbyttow/govips/v2/vips"
)

type ImageTransformOptions struct {
	Width   int
	Height  int
	Crop    bool
	Format  string
	Quality int
	Fit     string
	Dpr     float64
}

// ImageUtils interface for image operations
type ImageUtils interface {
	GetMimeType(data []byte) (string, error)
	TransformImage(imgData []byte, opts ImageTransformOptions) ([]byte, error)
}

// imageUtils is the concrete implementation of ImageUtils
type imageUtils struct{}

func NewImageUtils() ImageUtils {
	return &imageUtils{}
}

func (iu *imageUtils) GetMimeType(data []byte) (string, error) {
	image, err := vips.NewImageFromBuffer(data)
	if err != nil {
		return "", fmt.Errorf("failed to create image from buffer: %v", err)
	}
	defer image.Close()

	// Get the image format
	format := image.Format()

	var mimeType string
	switch format {
	case vips.ImageTypeJPEG:
		mimeType = "image/jpeg"
	case vips.ImageTypePNG:
		mimeType = "image/png"
	case vips.ImageTypeWEBP:
		mimeType = "image/webp"
	}

	return mimeType, nil
}

// New function to transform images
func (iu *imageUtils) TransformImage(imgData []byte, opts ImageTransformOptions) ([]byte, error) {
	image, err := vips.NewImageFromBuffer(imgData)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %v", err)
	}
	defer image.Close()

	width := int(math.Round(float64(opts.Width) * opts.Dpr))
	height := int(math.Round(float64(opts.Height) * opts.Dpr))

	fmt.Println("width: ", width)
	fmt.Println("height: ", height)

	// Apply fit modes with "clip" as the default
	switch opts.Fit {
	case "", "clip":
		if width == 0 && height == 0 {
			width, height = image.Width(), image.Height()
		} else if width == 0 {
			scale := float64(height) / float64(image.Height())
			width = int(float64(image.Width()) * scale)
		} else if height == 0 {
			scale := float64(width) / float64(image.Width())
			height = int(float64(image.Height()) * scale)
		}

		scaleWidth := float64(width) / float64(image.Width())
		scaleHeight := float64(height) / float64(image.Height())
		scale := scaleWidth
		if scaleHeight < scaleWidth {
			scale = scaleHeight
		}

		if err := image.Resize(scale, vips.KernelAuto); err != nil {
			return nil, fmt.Errorf("failed to resize image: %w", err)
		}

	default:
		return nil, errors.New("invalid fit option")
	}

	// Export the transformed image
	var modifiedImg []byte
	switch opts.Format {
	case "jpeg", "jpg":
		modifiedImg, _, err = image.ExportJpeg(&vips.JpegExportParams{Quality: int(opts.Quality)})
	case "png":
		modifiedImg, _, err = image.ExportPng(&vips.PngExportParams{Compression: int(opts.Quality)})
	case "webp":
		modifiedImg, _, err = image.ExportWebp(&vips.WebpExportParams{Quality: int(opts.Quality)})
	default:
		modifiedImg, _, err = image.ExportJpeg(&vips.JpegExportParams{Quality: int(opts.Quality)})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to export image: %v", err)
	}

	return modifiedImg, nil
}

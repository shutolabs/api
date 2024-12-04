package utils

import (
	"errors"
	"fmt"
	"math"

	"github.com/davidbyttow/govips/v2/vips"
)

type ImageTransformOptions struct {
	Width         int
	Height        int
	Fit          string    // clip, crop, fill, max, min, scale
	Format       string    // jpg, png, webp
	Quality      int       // 0-100
	Dpr          float64   // 1.0-3.0
	Blur         int       // 0-100
	ForceDownload bool
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

	// Apply DPR scaling
	width := int(math.Round(float64(opts.Width) * opts.Dpr))
	height := int(math.Round(float64(opts.Height) * opts.Dpr))

	// Apply fit modes according to spec
	switch opts.Fit {
	case "clip", "": // Default
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
	case "crop":
		if err := image.Thumbnail(width, height, vips.InterestingCentre); err != nil {
			return nil, fmt.Errorf("failed to crop image: %w", err)
		}
	case "fill":
		if err := image.Resize(float64(width)/float64(image.Width()), vips.KernelAuto); err != nil {
			return nil, fmt.Errorf("failed to fill image: %w", err)
		}
	default:
		return nil, errors.New("invalid fit option")
	}

	// Apply blur if specified
	if opts.Blur > 0 {
		sigma := float64(opts.Blur) * 0.3 // Convert blur parameter to sigma value
		if err := image.GaussianBlur(sigma); err != nil {
			return nil, fmt.Errorf("failed to apply blur: %w", err)
		}
	}

	// Export with correct format
	var modifiedImg []byte
	switch opts.Format {
	case "jpg", "jpeg":
		modifiedImg, _, err = image.ExportJpeg(&vips.JpegExportParams{Quality: opts.Quality})
	case "png":
		modifiedImg, _, err = image.ExportPng(&vips.PngExportParams{})
	case "webp":
		modifiedImg, _, err = image.ExportWebp(&vips.WebpExportParams{Quality: opts.Quality})
	default:
		// Default to original format
		modifiedImg, _, err = image.ExportJpeg(&vips.JpegExportParams{Quality: opts.Quality})
	}

	return modifiedImg, err
}

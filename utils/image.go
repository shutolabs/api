package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

type ImageTransformOptions struct {
	Width         int
	Height        int
	Fit          string    // clip, crop, fill
	Format       string    // jpg, png, webp
	Quality      int       // 0-100
	Dpr          float64   // 1.0-3.0
	Blur         int       // 0-100
	ForceDownload bool
}

type ImageMetadata struct {
	Width    int
	Height   int
	Keywords []string
}

// ImageUtils interface for image operations
type ImageUtils interface {
	GetMimeType(data []byte) (string, error)
	TransformImage(imgData []byte, opts ImageTransformOptions) ([]byte, error)
	GetImageMetadata(data []byte) (ImageMetadata, error)
}

// imageUtils is the concrete implementation of ImageUtils
type imageUtils struct {
	cmdExecutor CommandExecutor
}

func NewImageUtils() ImageUtils {
	return &imageUtils{
		cmdExecutor: NewCommandExecutor(),
	}
}

func (iu *imageUtils) GetMimeType(data []byte) (string, error) {
	image, err := vips.NewImageFromBuffer(data)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}
	defer image.Close()

	format := image.Format()
	switch format {
	case vips.ImageTypeJPEG:
		return "image/jpeg", nil
	case vips.ImageTypePNG:
		return "image/png", nil
	case vips.ImageTypeWEBP:
		return "image/webp", nil
	default:
		return "", fmt.Errorf("unsupported image format: %v", format)
	}
}

// New function to transform images
func (iu *imageUtils) TransformImage(imgData []byte, opts ImageTransformOptions) ([]byte, error) {
	image, err := vips.NewImageFromBuffer(imgData)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	defer image.Close()

	width := int(math.Round(float64(opts.Width) * opts.Dpr))
	height := int(math.Round(float64(opts.Height) * opts.Dpr))

	switch opts.Fit {
	case "clip", "":
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
		scale := math.Min(scaleWidth, scaleHeight)

		if err := image.Resize(scale, vips.KernelAuto); err != nil {
			return nil, fmt.Errorf("failed to resize image: %w", err)
		}

	case "crop":
		if err := image.Thumbnail(width, height, vips.InterestingCentre); err != nil {
			return nil, fmt.Errorf("failed to crop image: %w", err)
		}

	case "fill":
		scale := float64(width) / float64(image.Width())
		if err := image.Resize(scale, vips.KernelAuto); err != nil {
			return nil, fmt.Errorf("failed to fill image: %w", err)
		}

	default:
		return nil, fmt.Errorf("invalid fit option: %s", opts.Fit)
	}

	if opts.Blur > 0 {
		sigma := float64(opts.Blur) * 0.3
		if err := image.GaussianBlur(sigma); err != nil {
			return nil, fmt.Errorf("failed to apply blur: %w", err)
		}
	}

	var modifiedImg []byte
	var exportErr error

	switch opts.Format {
	case "jpg", "jpeg":
		modifiedImg, _, exportErr = image.ExportJpeg(&vips.JpegExportParams{Quality: opts.Quality})
	case "png":
		modifiedImg, _, exportErr = image.ExportPng(&vips.PngExportParams{})
	case "webp":
		modifiedImg, _, exportErr = image.ExportWebp(&vips.WebpExportParams{Quality: opts.Quality})
	default:
		modifiedImg, _, exportErr = image.ExportJpeg(&vips.JpegExportParams{Quality: opts.Quality})
	}

	if exportErr != nil {
		return nil, fmt.Errorf("failed to export image: %w", exportErr)
	}

	return modifiedImg, nil
}

// Add this custom type for flexible keyword parsing
type flexibleKeywords []string

// UnmarshalJSON implements custom unmarshaling for keywords that can be either a string or an array of strings
func (fk *flexibleKeywords) UnmarshalJSON(data []byte) error {
	var strArray []string
	if err := json.Unmarshal(data, &strArray); err == nil {
		*fk = strArray
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*fk = []string{str}
		return nil
	}

	return errors.New("keywords must be a string or array of strings")
}

func (iu *imageUtils) GetImageMetadata(data []byte) (ImageMetadata, error) {
	image, err := vips.NewImageFromBuffer(data)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("failed to read image: %w", err)
	}
	defer image.Close()

	return ImageMetadata{
		Width:    image.Width(),
		Height:   image.Height(),
		Keywords: nil, // VIPS doesn't support reading IPTC keywords
	}, nil
}

// IsImageFile checks if a file is an image based on its extension
func IsImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

// ParseImageOptionsFromRequest extracts image transformation options from request query parameters
func ParseImageOptionsFromRequest(r *http.Request) ImageTransformOptions {
	width, _ := strconv.Atoi(r.URL.Query().Get("w"))
	height, _ := strconv.Atoi(r.URL.Query().Get("h"))
	fit := r.URL.Query().Get("fit")
	if fit == "" {
		fit = "clip" // Default as per spec
	}
	
	dpr, err := strconv.ParseFloat(r.URL.Query().Get("dpr"), 64)
	if err != nil || dpr == 0 {
		dpr = 1.0 // Default as per spec
	}
	if dpr > 3.0 {
		dpr = 3.0 // Max value as per spec
	}
	
	format := r.URL.Query().Get("fm")
	quality, err := strconv.Atoi(r.URL.Query().Get("q"))
	if err != nil || quality == 0 {
		quality = 75 // Default as per spec
	}
	
	blur, _ := strconv.Atoi(r.URL.Query().Get("blur"))
	forceDownload := r.URL.Query().Get("dl") == "1"

	return ImageTransformOptions{
		Width:         width,
		Height:        height,
		Fit:          fit,
		Format:       format,
		Quality:      quality,
		Dpr:          dpr,
		Blur:         blur,
		ForceDownload: forceDownload,
	}
}

// HasImageTransformParams checks if any image transformation parameters are present in the request
func HasImageTransformParams(r *http.Request) bool {
	params := []string{"w", "h", "fit", "dpr", "fm", "q", "blur"}
	for _, param := range params {
		if r.URL.Query().Get(param) != "" {
			return true
		}
	}
	return false
}

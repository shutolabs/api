package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"shuto-api/security"

	"github.com/joho/godotenv"
)

func main() {
	// Define flags
	var (
		path          = flag.String("path", "", "Path to sign (required)")
		endpoint      = flag.String("endpoint", "image", "Endpoint to use (image or download)")
		width         = flag.Int("w", 0, "Width of the image")
		height        = flag.Int("h", 0, "Height of the image")
		fit           = flag.String("fit", "", "Fit mode (clip, scale, etc.)")
		format        = flag.String("fm", "", "Output format")
		quality       = flag.Int("q", 0, "Quality (1-100)")
		dpr          = flag.Float64("dpr", 0, "Device pixel ratio")
		blur         = flag.Int("blur", 0, "Blur amount")
		download     = flag.Bool("dl", false, "Force download")
		timeless     = flag.Bool("timeless", false, "Generate a timeless URL")
		validityMins = flag.Int("validity", 5, "Validity period in minutes (for time-bound URLs)")
	)

	flag.Parse()

	if *path == "" {
		fmt.Println("Error: path is required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate endpoint
	*endpoint = strings.ToLower(*endpoint)
	if *endpoint != "image" && *endpoint != "download" {
		fmt.Println("Error: endpoint must be either 'image' or 'download'")
		flag.Usage()
		os.Exit(1)
	}

	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	secretKey := os.Getenv("HMAC_SECRET_KEY")
	if secretKey == "" {
		fmt.Println("Error: HMAC_SECRET_KEY not set in environment")
		os.Exit(1)
	}

	// Build query parameters
	params := url.Values{}
	// Add image transformation parameters for both image and download endpoints
	if *width > 0 {
		params.Set("w", fmt.Sprintf("%d", *width))
	}
	if *height > 0 {
		params.Set("h", fmt.Sprintf("%d", *height))
	}
	if *fit != "" {
		params.Set("fit", *fit)
	}
	if *format != "" {
		params.Set("fm", *format)
	}
	if *quality > 0 {
		params.Set("q", fmt.Sprintf("%d", *quality))
	}
	if *dpr > 0 {
		params.Set("dpr", fmt.Sprintf("%.2f", *dpr))
	}
	if *blur > 0 {
		params.Set("blur", fmt.Sprintf("%d", *blur))
	}
	if *download {
		params.Set("dl", "1")
	}

	// Create URL signer
	validityWindow := 0
	if !*timeless {
		validityWindow = *validityMins * 60 // Convert minutes to seconds
	}
	
	signer, err := security.NewURLSigner([]security.SecretKey{{ID: "v1", Secret: []byte(secretKey)}}, validityWindow, "", *endpoint)
	if err != nil {
		fmt.Printf("Error creating URL signer: %v\n", err)
		os.Exit(1)
	}

	// Generate signed URL
	signedURL, err := signer.GenerateSignedURL(*path, params)
	if err != nil {
		fmt.Printf("Error generating signed URL: %v\n", err)
		os.Exit(1)
	}

	// Print the result
	fmt.Printf("\nSigned %s URL:\n", strings.Title(*endpoint))
	fmt.Printf("%s\n", signedURL)

	if !*timeless {
		fmt.Printf("\nThis URL will expire in %d minutes\n", *validityMins)
	} else {
		fmt.Println("\nThis is a permanent URL (will not expire)")
	}
} 
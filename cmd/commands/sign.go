package commands

import (
	"bytes"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"shuto-api/security"

	"github.com/joho/godotenv"
)

// SignCommand represents the sign command
type SignCommand struct {
	fs *flag.FlagSet
	
	// Command flags
	path        *string
	endpoint    *string
	width       *int
	height      *int
	fit         *string
	format      *string
	quality     *int
	dpr         *float64
	blur        *int
	download    *bool
	timeless    *bool
	validityMins *int
}

func NewSignCommand() *SignCommand {
	c := &SignCommand{}
	c.fs = flag.NewFlagSet(c.Name(), flag.ExitOnError)
	
	// Define flags
	c.path = c.fs.String("path", "", "Path to sign (required)")
	c.endpoint = c.fs.String("endpoint", "image", "Endpoint to use (image or download)")
	c.width = c.fs.Int("w", 0, "Width of the image")
	c.height = c.fs.Int("h", 0, "Height of the image")
	c.fit = c.fs.String("fit", "", "Fit mode (clip, scale, etc.)")
	c.format = c.fs.String("fm", "", "Output format")
	c.quality = c.fs.Int("q", 0, "Quality (1-100)")
	c.dpr = c.fs.Float64("dpr", 0, "Device pixel ratio")
	c.blur = c.fs.Int("blur", 0, "Blur amount")
	c.download = c.fs.Bool("dl", false, "Force download")
	c.timeless = c.fs.Bool("timeless", false, "Generate a timeless URL")
	c.validityMins = c.fs.Int("validity", 5, "Validity period in minutes (for time-bound URLs)")
	
	return c
}

func (c *SignCommand) Name() string {
	return "sign"
}

func (c *SignCommand) Description() string {
	return "Generate a signed URL for image or download endpoints"
}

func (c *SignCommand) Usage() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Usage: shuto-cli %s [options]\n\n", c.Name()))
	buf.WriteString("Options:\n")
	c.fs.SetOutput(&buf)
	c.fs.PrintDefaults()
	return buf.String()
}

func (c *SignCommand) Execute(args []string) {
	c.fs.Parse(args)

	if *c.path == "" {
		fmt.Println("Error: path is required")
		fmt.Println(c.Usage())
		os.Exit(1)
	}

	// Validate endpoint
	*c.endpoint = strings.ToLower(*c.endpoint)
	if *c.endpoint != "image" && *c.endpoint != "download" {
		fmt.Println("Error: endpoint must be either 'image' or 'download'")
		fmt.Println(c.Usage())
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
	if *c.width > 0 {
		params.Set("w", fmt.Sprintf("%d", *c.width))
	}
	if *c.height > 0 {
		params.Set("h", fmt.Sprintf("%d", *c.height))
	}
	if *c.fit != "" {
		params.Set("fit", *c.fit)
	}
	if *c.format != "" {
		params.Set("fm", *c.format)
	}
	if *c.quality > 0 {
		params.Set("q", fmt.Sprintf("%d", *c.quality))
	}
	if *c.dpr > 0 {
		params.Set("dpr", fmt.Sprintf("%.2f", *c.dpr))
	}
	if *c.blur > 0 {
		params.Set("blur", fmt.Sprintf("%d", *c.blur))
	}
	if *c.download {
		params.Set("dl", "1")
	}

	// Create URL signer
	validityWindow := 0
	if !*c.timeless {
		validityWindow = *c.validityMins * 60 // Convert minutes to seconds
	}
	
	signer, err := security.NewURLSigner([]security.SecretKey{{ID: "v1", Secret: []byte(secretKey)}}, validityWindow, "", *c.endpoint)
	if err != nil {
		fmt.Printf("Error creating URL signer: %v\n", err)
		os.Exit(1)
	}

	// Generate signed URL
	signedURL, err := signer.GenerateSignedURL(*c.path, params)
	if err != nil {
		fmt.Printf("Error generating signed URL: %v\n", err)
		os.Exit(1)
	}

	// Print the result
	fmt.Printf("\nSigned %s URL:\n", strings.Title(*c.endpoint))
	fmt.Printf("%s\n", signedURL)

	if !*c.timeless {
		fmt.Printf("\nThis URL will expire in %d minutes\n", *c.validityMins)
	} else {
		fmt.Println("\nThis is a permanent URL (will not expire)")
	}
} 
# Shuto API

This repository contains the API service for reducing / enhancing and modifying images.

## Setup

### Prerequisites

- Go 1.21 or higher
- libvips 8.14 or higher (required for image processing)
  - macOS: `brew install vips`
  - Ubuntu: `apt-get install libvips-dev`
- rclone (for remote storage operations)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/shuto-api.git
   cd shuto-api
   ```

2. Install Go dependencies:

   ```bash
   go mod download
   ```

3. Copy the environment file and configure your settings:

   ```bash
   cp .env.example .env
   ```

   Edit `.env` with your configuration values.

4. Configure your domains in `config/domains.yaml`. Example configuration:
   ```yaml
   domains:
     localhost:
       rclone:
         remote: test
         flags: []
     your-domain:
       rclone:
         remote: webdav
         flags:
           - --webdav-url=${RCLONE_CONFIG_SERVER_URL}
           - --webdav-vendor=${RCLONE_CONFIG_SERVER_VENDOR}
           - --webdav-user=${RCLONE_CONFIG_SERVER_USER}
           - --webdav-pass=${RCLONE_CONFIG_SERVER_PASS}
       security:
         mode: hmac_timebound
         secrets:
           - key_id: "v1"
             secret: "${HMAC_SECRET_KEY}"
         validity_window: 300
   ```

### Running the Service

1. Start the server:

   ```bash
   go run main.go
   ```

   The server will start on the configured port (default: 8080)

2. Test the service:
   ```bash
   curl http://localhost:8080/v1/list/
   ```

### Development

- Run tests:

  ```bash
  go test ./...
  ```

- Build the binary:
  ```bash
  go build -o shuto-api
  ```

## Features

### Image Transformation (`/v1/image/`)

- Resize images with width and height parameters
- Multiple fit options for resizing (e.g., crop)
- Format conversion (supports WebP, AVIF, JPEG)
- Quality adjustment
- DPR (Device Pixel Ratio) support
- Blur effects
- Force download option
- Automatic format selection based on browser support
- Caching support with long-term cache headers

### Directory Listing (`/v1/list/`)

- List files and directories
- Returns detailed file information including:
  - File path
  - File size
  - MIME type
  - Directory status
  - Image dimensions (for image files)
  - Image keywords/metadata (if available)
- Metadata caching for improved performance

### File Download (`/v1/download/`)

- Single file downloads
- Bulk directory downloads (as ZIP)
- Support for image transformations during download
- Size limit protection for bulk downloads
- Force download option
- Concurrent processing for bulk downloads

### Security Features

- Domain-based configuration
- Support for signed URLs
- Timebound access control
- Multiple secret key support
- Configurable validity windows

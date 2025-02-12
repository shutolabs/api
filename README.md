# Shuto API

This repository contains the API service for reducing / enhancing and modifying images.

## Running the Service

The following instructions are for users who want to run the service using Docker. For development setup, see the [Development](#development) section.

### Configuration

The API requires two main configuration files:

#### 1. Domain Configuration (domains.yaml)

Basic example of `domains.yaml`:

```yaml
domains:
  localhost:
    rclone:
      remote: local
      flags: []

  example.com:
    rclone:
      remote: webdav
      # Flags can be used to configure the remote instead of rclone.conf
      flags:
        - --webdav-url=${WEBDAV_URL}
        - --webdav-vendor=${WEBDAV_VENDOR}
        - --webdav-user=${WEBDAV_USER}
        - --webdav-pass=${WEBDAV_PASS}
    security:
      mode: hmac_timebound
      secrets:
        - key_id: "v1"
          secret: "${HMAC_SECRET_KEY}"
      validity_window: 300
```

#### 2. Rclone Configuration (rclone.conf)

Basic example of `rclone.conf`:

```ini
[local]
type = local

[webdav]
type = webdav
# Configuration can be done here instead of using flags in domains.yaml
url = https://your-webdav-server.com
vendor = nextcloud
user = your-username
pass = your-password
```

More detailed documentation about configuration options will be available soon.

### Docker Deployment

The Docker container requires configuration files to be mounted as volumes. Make sure you have the following files ready:

- `domains.yaml`: Domain configuration file
- `rclone.conf`: Rclone configuration file
- `.env`: Environment variables file (optional)

1. Using Docker directly:

   ```bash
   docker run -d \
     -p 8080:8080 \
     -v ./domains.yaml:/app/config/domains.yaml:ro \
     -v ./rclone.conf:/root/.config/rclone/rclone.conf:ro \
     -v ./images:/app/images \
     --env-file .env \
     ghcr.io/lgastler/shuto-api:latest
   ```

2. Using Docker Compose:

   See [docker-compose.yml](docker-compose.yml) for an example deployment configuration.

Note: The container will fail to start if either `domains.yaml` or `rclone.conf` is not mounted. This is a safety measure to ensure proper configuration.

## Development

### Prerequisites

- Go 1.21 or higher
- libvips 8.14 or higher (required for image processing)
  - macOS: `brew install vips`
  - Ubuntu: `apt-get install libvips-dev`
- rclone (for remote storage operations)

### Development Tasks

- Run tests:

  ```bash
  go test ./...
  ```

- Build the binary:
  ```bash
  go build -o shuto-api
  ```

## Features

### Image Transformation (`/v2/image/`)

- Resize images with width and height parameters
- Multiple fit options for resizing (e.g., crop)
- Format conversion (supports WebP, AVIF, JPEG)
- Quality adjustment
- DPR (Device Pixel Ratio) support
- Blur effects
- Force download option
- Automatic format selection based on browser support
- Caching support with long-term cache headers

### Directory Listing (`/v2/list/`)

- List files and directories
- Returns detailed file information including:
  - File path
  - File size
  - MIME type
  - Directory status
  - Image dimensions (for image files)
  - Image keywords/metadata (if available)
- Metadata caching for improved performance

### File Download (`/v2/download/`)

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

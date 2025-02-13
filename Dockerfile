# Build stage
FROM golang:1.23-alpine AS builder

# Install required dependencies
RUN apk add --no-cache gcc musl-dev vips-dev=8.15.3-r5

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache vips-dev=8.15.3-r5 rclone~=1.68.2

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Create necessary directories with proper permissions
RUN mkdir -p /app/config && \
    mkdir -p /root/.config/rclone && \
    chmod 755 /root/.config && \
    chmod 755 /root/.config/rclone

# Create a script to check for required configuration files
COPY <<'EOF' /app/docker-entrypoint.sh
#!/bin/sh
set -e

# Check for required configuration files
if [ ! -f "/root/.config/rclone/rclone.conf" ]; then
    echo "Error: rclone.conf not found. Please mount it to /root/.config/rclone/rclone.conf"
    echo "Example: -v $(pwd)/rclone.conf:/root/.config/rclone/rclone.conf:ro"
    ls -la /root/.config/rclone  # Debug: List contents of rclone config directory
    exit 1
fi

if [ ! -f "/app/config/domains.yaml" ]; then
    echo "Error: domains.yaml not found. Please mount it to /app/config/domains.yaml"
    echo "Example: -v $(pwd)/config/domains.yaml:/app/config/domains.yaml:ro"
    exit 1
fi

# Start the application
exec "$@"
EOF

RUN chmod +x /app/docker-entrypoint.sh

# Create volume mount points
VOLUME ["/app/config", "/app/images", "/root/.config/rclone"]

# Expose the port the app runs on
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["./main"]
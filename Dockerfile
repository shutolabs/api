# Build stage
FROM golang:1.23-alpine AS builder

# Install required dependencies
RUN apk add --no-cache gcc musl-dev vips-dev

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
RUN apk add --no-cache vips-dev rclone

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy rclone configuration
COPY ./rclone.conf /root/.config/rclone/rclone.conf

# Copy the domains.yaml file
COPY ./config/domains.yaml ./config/domains.yaml

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./main"]
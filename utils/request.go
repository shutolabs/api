package utils

import (
	"net/http"
	"strings"
)

// GetDomainFromRequest extracts the domain from an HTTP request
func GetDomainFromRequest(r *http.Request) string {
	// First try the X-Forwarded-Host header
	domain := r.Header.Get("X-Forwarded-Host")
	if domain != "" {
		return strings.Split(domain, ":")[0] // Remove port if present
	}

	// Then try the Host header
	domain = r.Host
	if domain != "" {
		return strings.Split(domain, ":")[0] // Remove port if present
	}

	return "default"
} 
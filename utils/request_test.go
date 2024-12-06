package utils

import (
	"net/http"
	"testing"
)

func TestGetDomainFromRequest(t *testing.T) {
	tests := []struct {
		name           string
		forwardedHost  string
		host           string
		expectedDomain string
	}{
		{
			name:           "With X-Forwarded-Host header",
			forwardedHost:  "example.com:8080",
			host:           "localhost:8080",
			expectedDomain: "example.com",
		},
		{
			name:           "With Host header only",
			forwardedHost:  "",
			host:           "test.com:8080",
			expectedDomain: "test.com",
		},
		{
			name:           "With no port in headers",
			forwardedHost:  "example.com",
			host:           "test.com",
			expectedDomain: "example.com",
		},
		{
			name:           "With no headers",
			forwardedHost:  "",
			host:           "",
			expectedDomain: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				Host: tt.host,
				Header: http.Header{},
			}
			if tt.forwardedHost != "" {
				req.Header.Set("X-Forwarded-Host", tt.forwardedHost)
			}

			got := GetDomainFromRequest(req)
			if got != tt.expectedDomain {
				t.Errorf("GetDomainFromRequest() = %v, want %v", got, tt.expectedDomain)
			}
		})
	}
} 
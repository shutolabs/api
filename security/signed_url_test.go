package security

import (
	"net/url"
	"testing"
	"time"
)

func createTestKeys() []SecretKey {
	return []SecretKey{
		{
			ID:     "v1",
			Secret: []byte("test-secret-1"),
		},
		{
			ID:     "v2",
			Secret: []byte("test-secret-2"),
		},
	}
}

func TestNewURLSigner(t *testing.T) {
	tests := []struct {
		name        string
		keys        []SecretKey
		defaultKey  string
		validity    int
		expectError bool
	}{
		{
			name:        "valid configuration with default key",
			keys:        createTestKeys(),
			defaultKey:  "v1",
			validity:    300,
			expectError: false,
		},
		{
			name:        "valid configuration without default key",
			keys:        createTestKeys(),
			defaultKey:  "",
			validity:    300,
			expectError: false,
		},
		{
			name:        "no keys provided",
			keys:        []SecretKey{},
			defaultKey:  "",
			validity:    300,
			expectError: true,
		},
		{
			name:        "invalid default key",
			keys:        createTestKeys(),
			defaultKey:  "v3",
			validity:    300,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := NewURLSigner(tt.keys, tt.validity, tt.defaultKey)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if signer == nil {
					t.Error("expected signer, got nil")
				}
			}
		})
	}
}

func TestURLSigner_GenerateAndValidateTimeboundURL(t *testing.T) {
	keys := createTestKeys()
	signer, err := NewURLSigner(keys, 300, "v1") // 5 minutes validity
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	path := "test/image.jpg"
	params := url.Values{}
	params.Set("width", "100")
	params.Set("height", "100")

	signedURL, err := signer.GenerateSignedURL(path, params)
	if err != nil {
		t.Fatalf("Failed to generate signed URL: %v", err)
	}

	// Parse the URL to get query parameters
	parsedURL, err := url.Parse(signedURL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}

	// Verify key ID in URL
	if kid := parsedURL.Query().Get("kid"); kid != "v1" {
		t.Errorf("Expected key ID 'v1', got '%s'", kid)
	}

	// Validate the signed URL
	err = signer.ValidateSignedURL(path, parsedURL.Query())
	if err != nil {
		t.Errorf("Failed to validate signed URL: %v", err)
	}
}

func TestURLSigner_KeyRotation(t *testing.T) {
	// Create signer with v1 key
	oldSigner, err := NewURLSigner([]SecretKey{{ID: "v1", Secret: []byte("old-secret")}}, 300, "v1")
	if err != nil {
		t.Fatalf("Failed to create old signer: %v", err)
	}

	// Generate URL with old key
	path := "test/image.jpg"
	params := url.Values{}
	oldURL, err := oldSigner.GenerateSignedURL(path, params)
	if err != nil {
		t.Fatalf("Failed to generate URL with old key: %v", err)
	}

	// Create new signer with both keys (simulating key rotation)
	newKeys := []SecretKey{
		{ID: "v1", Secret: []byte("old-secret")},
		{ID: "v2", Secret: []byte("new-secret")},
	}
	newSigner, err := NewURLSigner(newKeys, 300, "v2")
	if err != nil {
		t.Fatalf("Failed to create new signer: %v", err)
	}

	// Verify old URL still works with new signer
	parsedURL, _ := url.Parse(oldURL)
	err = newSigner.ValidateSignedURL(path, parsedURL.Query())
	if err != nil {
		t.Errorf("Failed to validate old URL after key rotation: %v", err)
	}

	// Generate and verify new URL with new key
	newURL, err := newSigner.GenerateSignedURL(path, params)
	if err != nil {
		t.Fatalf("Failed to generate URL with new key: %v", err)
	}

	parsedNewURL, _ := url.Parse(newURL)
	if kid := parsedNewURL.Query().Get("kid"); kid != "v2" {
		t.Errorf("Expected key ID 'v2', got '%s'", kid)
	}
}

func TestURLSigner_ExpiredURL(t *testing.T) {
	signer, err := NewURLSigner(createTestKeys(), 1, "v1") // 1 second validity
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	path := "test/image.jpg"
	params := url.Values{}

	signedURL, err := signer.GenerateSignedURL(path, params)
	if err != nil {
		t.Fatalf("Failed to generate signed URL: %v", err)
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	parsedURL, _ := url.Parse(signedURL)
	err = signer.ValidateSignedURL(path, parsedURL.Query())
	if err != ErrExpiredURL {
		t.Errorf("Expected ErrExpiredURL, got: %v", err)
	}
}

func TestURLSigner_InvalidKey(t *testing.T) {
	signer, err := NewURLSigner(createTestKeys(), 300, "v1")
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	path := "test/image.jpg"
	params := url.Values{}
	params.Set("kid", "v3") // Non-existent key
	params.Set("sig", "invalid")

	err = signer.ValidateSignedURL(path, params)
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got: %v", err)
	}
}

func TestURLSigner_InvalidSignature(t *testing.T) {
	// Create two signers with different secrets but same key ID
	signer1, _ := NewURLSigner([]SecretKey{{ID: "v1", Secret: []byte("secret1")}}, 300, "v1")
	signer2, _ := NewURLSigner([]SecretKey{{ID: "v1", Secret: []byte("secret2")}}, 300, "v1")

	path := "test/image.jpg"
	params := url.Values{}

	// Generate URL with first signer
	signedURL, err := signer1.GenerateSignedURL(path, params)
	if err != nil {
		t.Fatalf("Failed to generate signed URL: %v", err)
	}

	// Validate with second signer (should fail)
	parsedURL, _ := url.Parse(signedURL)
	err = signer2.ValidateSignedURL(path, parsedURL.Query())
	if err != ErrInvalidSignature {
		t.Errorf("Expected ErrInvalidSignature, got: %v", err)
	}
}

func TestURLSigner_TimelessURL(t *testing.T) {
	signer, err := NewURLSigner(createTestKeys(), 0, "v1") // 0 means timeless
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	path := "test/image.jpg"
	params := url.Values{}

	signedURL, err := signer.GenerateSignedURL(path, params)
	if err != nil {
		t.Fatalf("Failed to generate signed URL: %v", err)
	}

	parsedURL, _ := url.Parse(signedURL)
	if ts := parsedURL.Query().Get("ts"); ts != "" {
		t.Error("Timeless URL should not contain timestamp")
	}

	// Validate the URL
	err = signer.ValidateSignedURL(path, parsedURL.Query())
	if err != nil {
		t.Errorf("Failed to validate timeless URL: %v", err)
	}
} 
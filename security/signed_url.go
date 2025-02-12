package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"shuto-api/config"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrExpiredURL      = errors.New("URL has expired")
	ErrInvalidFormat   = errors.New("invalid URL format")
	ErrKeyNotFound     = errors.New("signing key not found")
)

// SecretKey represents a key used for signing
type SecretKey struct {
	ID     string
	Secret []byte
}

// URLSigner handles the generation and validation of signed URLs
type URLSigner struct {
	keys           map[string]SecretKey
	validityWindow int // in seconds, 0 means indefinite
	defaultKeyID   string
	endpoint       string // the endpoint to use (image or download)
}

// NewURLSigner creates a new URLSigner instance
func NewURLSigner(keys []SecretKey, validityWindow int, defaultKeyID string, endpoint string) (*URLSigner, error) {
	if len(keys) == 0 {
		return nil, errors.New("at least one key is required")
	}

	keyMap := make(map[string]SecretKey)
	for _, key := range keys {
		keyMap[key.ID] = key
	}

	// If no default key is specified, use the first key
	if defaultKeyID == "" {
		defaultKeyID = keys[0].ID
	}

	// Verify the default key exists
	if _, exists := keyMap[defaultKeyID]; !exists {
		return nil, errors.New("default key ID not found in provided keys")
	}

	// Default to image endpoint if none specified
	if endpoint == "" {
		endpoint = "image"
	}

	// Validate endpoint
	if endpoint != "image" && endpoint != "download" {
		return nil, errors.New("endpoint must be either 'image' or 'download'")
	}

	return &URLSigner{
		keys:           keyMap,
		validityWindow: validityWindow,
		defaultKeyID:   defaultKeyID,
		endpoint:       endpoint,
	}, nil
}

// GenerateSignedURL creates a signed URL with optional time bound
func (s *URLSigner) GenerateSignedURL(path string, params url.Values) (string, error) {
	if s.validityWindow > 0 {
		return s.generateTimeboundURL(path, params)
	}
	return s.generateTimelessURL(path, params)
}

func (s *URLSigner) generateTimeboundURL(path string, params url.Values) (string, error) {
	timestamp := time.Now().Unix()
	message := fmt.Sprintf("%s|%d|%s", path, timestamp, params.Encode())
	
	key := s.keys[s.defaultKeyID]
	mac := hmac.New(sha256.New, key.Secret)
	mac.Write([]byte(message))
	signature := hex.EncodeToString(mac.Sum(nil))
	
	// Create a copy of params to avoid modifying the original
	signedParams := url.Values{}
	for k, v := range params {
		signedParams[k] = v
	}
	
	signedParams.Set("kid", key.ID)
	signedParams.Set("ts", fmt.Sprintf("%d", timestamp))
	signedParams.Set("sig", signature)
	
	return fmt.Sprintf("/v2/%s/%s?%s", s.endpoint, path, signedParams.Encode()), nil
}

func (s *URLSigner) generateTimelessURL(path string, params url.Values) (string, error) {
	message := fmt.Sprintf("%s|%s", path, params.Encode())
	
	key := s.keys[s.defaultKeyID]
	mac := hmac.New(sha256.New, key.Secret)
	mac.Write([]byte(message))
	signature := hex.EncodeToString(mac.Sum(nil))
	
	// Create a copy of params to avoid modifying the original
	signedParams := url.Values{}
	for k, v := range params {
		signedParams[k] = v
	}
	
	signedParams.Set("kid", key.ID)
	signedParams.Set("sig", signature)
	
	return fmt.Sprintf("/v2/%s/%s?%s", s.endpoint, path, signedParams.Encode()), nil
}

// ValidateSignedURL validates a signed URL
func (s *URLSigner) ValidateSignedURL(path string, params url.Values) error {
	// Extract and remove signature parameters
	signature := params.Get("sig")
	keyID := params.Get("kid")
	timestamp := params.Get("ts")
	
	if signature == "" || keyID == "" {
		return ErrInvalidFormat
	}

	// Look up the key
	key, exists := s.keys[keyID]
	if !exists {
		return ErrKeyNotFound
	}
	
	// Create a copy of params without signature parameters
	validationParams := url.Values{}
	for k, v := range params {
		if k != "sig" && k != "kid" && k != "ts" {
			validationParams[k] = v
		}
	}
	
	if s.validityWindow > 0 {
		if timestamp == "" {
			return ErrInvalidFormat
		}
		return s.validateTimeboundURL(path, validationParams, signature, timestamp, key.Secret)
	}
	
	return s.validateTimelessURL(path, validationParams, signature, key.Secret)
}

func (s *URLSigner) validateTimeboundURL(path string, params url.Values, providedSignature, timestamp string, secret []byte) error {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return ErrInvalidFormat
	}

	// Check expiration
	if time.Now().Unix()-ts > int64(s.validityWindow) {
		return ErrExpiredURL
	}

	message := fmt.Sprintf("%s|%d|%s", path, ts, params.Encode())
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(providedSignature), []byte(expectedSignature)) {
		return ErrInvalidSignature
	}

	return nil
}

func (s *URLSigner) validateTimelessURL(path string, params url.Values, providedSignature string, secret []byte) error {
	message := fmt.Sprintf("%s|%s", path, params.Encode())
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(providedSignature), []byte(expectedSignature)) {
		return ErrInvalidSignature
	}

	return nil
}

// ValidateSignedURLFromConfig validates a signed URL using the provided security configuration
func ValidateSignedURLFromConfig(path string, query url.Values, secrets []config.SecretKey, validityWindow int) error {
	// Convert config secrets to security.SecretKey
	keys := make([]SecretKey, len(secrets))
	for i, secret := range secrets {
		keys[i] = SecretKey{
			ID:     secret.KeyID,
			Secret: []byte(secret.Secret),
		}
	}

	signer, err := NewURLSigner(keys, validityWindow, "", "") // endpoint not needed for validation
	if err != nil {
		return fmt.Errorf("failed to create URL signer: %w", err)
	}

	return signer.ValidateSignedURL(path, query)
} 
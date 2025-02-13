package utils

import (
	"encoding/json"
	"net/http"
	"os"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

func isDevelopment() bool {
	return os.Getenv("APP_ENV") == "development"
}

const (
	ErrCodeInvalidRequest     = "INVALID_REQUEST"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeInternalError     = "INTERNAL_ERROR"
	ErrCodeInvalidDomain     = "INVALID_DOMAIN"
	ErrCodeInvalidPath       = "INVALID_PATH"
	ErrCodeInvalidAPIKey     = "INVALID_API_KEY"
	ErrCodeExpiredToken      = "EXPIRED_TOKEN"
	ErrCodeInvalidSignature  = "INVALID_SIGNATURE"
)

func WriteError(w http.ResponseWriter, status int, code string, message string, details string) {
	resp := ErrorResponse{
		Error: message,
		Code:  code,
	}

	// Only include details in development environment
	if isDevelopment() && details != "" {
		resp.Details = details
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)

	// Log the error with appropriate level based on status code
	if status >= 500 {
		Error("Server error", "code", code, "message", message, "details", details)
	} else {
		Debug("Client error", "code", code, "message", message, "details", details)
	}
}

func WriteInvalidRequestError(w http.ResponseWriter, message string, details string) {
	WriteError(w, http.StatusBadRequest, ErrCodeInvalidRequest, message, details)
}

func WriteUnauthorizedError(w http.ResponseWriter, details string) {
	WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized", details)
}

func WriteForbiddenError(w http.ResponseWriter, details string) {
	WriteError(w, http.StatusForbidden, ErrCodeForbidden, "Forbidden", details)
}

func WriteNotFoundError(w http.ResponseWriter, message string, details string) {
	WriteError(w, http.StatusNotFound, ErrCodeNotFound, message, details)
}

func WriteInternalError(w http.ResponseWriter, message string, details string) {
	WriteError(w, http.StatusInternalServerError, ErrCodeInternalError, message, details)
}

func WriteInvalidDomainError(w http.ResponseWriter, domain string) {
	WriteError(w, http.StatusBadRequest, ErrCodeInvalidDomain, "Invalid domain", domain)
}

func WriteInvalidPathError(w http.ResponseWriter, path string) {
	WriteError(w, http.StatusBadRequest, ErrCodeInvalidPath, "Invalid path", path)
}

func WriteInvalidAPIKeyError(w http.ResponseWriter) {
	WriteError(w, http.StatusUnauthorized, ErrCodeInvalidAPIKey, "Invalid or missing API key", "")
}

func WriteExpiredTokenError(w http.ResponseWriter) {
	WriteError(w, http.StatusGone, ErrCodeExpiredToken, "Token has expired", "")
}

func WriteInvalidSignatureError(w http.ResponseWriter) {
	WriteError(w, http.StatusForbidden, ErrCodeInvalidSignature, "Invalid signature", "")
} 
package helper

import (
	"database/sql"
	"errors"
	"log"
)

// Safe error messages for clients
const (
	ErrMsgInternalServer = "Internal server error"
	ErrMsgUnauthorized   = "Unauthorized"
	ErrMsgNotFound       = "Resource not found"
	ErrMsgBadRequest     = "Invalid request"
	ErrMsgForbidden      = "Access forbidden"
)

// LogAndSanitize logs the actual error internally and returns a safe message for the client
func LogAndSanitize(err error, context string) string {
	// Log actual error with context for debugging
	log.Printf("[ERROR] %s: %v", context, err)

	// Return safe generic message based on error type
	if errors.Is(err, sql.ErrNoRows) {
		return ErrMsgNotFound
	}

	// Default to internal server error
	return ErrMsgInternalServer
}

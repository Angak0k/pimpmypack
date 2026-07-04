package security

import "time"

// RefreshToken represents a refresh token row in the database. It is a
// storage model, never serialized to clients — no json tags on purpose:
// Token holds the plaintext right after creation but the SHA-256 hash when
// read back from the database.
type RefreshToken struct {
	ID         uint
	Token      string
	AccountID  uint
	ExpiresAt  time.Time
	CreatedAt  time.Time
	LastUsedAt *time.Time
	Revoked    bool
}

// RefreshTokenInput represents the input for refresh token endpoint
type RefreshTokenInput struct {
	Token string `json:"refresh_token" binding:"required"`
}

// TokenPairResponse represents access + refresh token pair
type TokenPairResponse struct {
	Token            string `json:"token"` // Backward compatibility - same as AccessToken
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	AccessExpiresIn  int64  `json:"access_expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}

// RefreshResponse represents the response from refresh endpoint
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// LogoutAllResponse is the response of POST /v1/auth/logout-all
type LogoutAllResponse struct {
	Message         string `json:"message"`
	RevokedSessions int64  `json:"revoked_sessions"`
}

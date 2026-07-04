package security

import "time"

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID         uint       `json:"id"`
	Token      string     `json:"token"`
	AccountID  uint       `json:"account_id"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	Revoked    bool       `json:"revoked"`
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

// RefreshResponse represents the response from refresh endpoint.
// RefreshToken carries a freshly rotated token clients should store; the
// presented token currently stays valid (strict rotation lands once clients
// have adopted the new field).
type RefreshResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	RefreshExpiresIn int64  `json:"refresh_expires_in,omitempty"`
}

// LogoutResponse is the response of POST /auth/logout
type LogoutResponse struct {
	Message string `json:"message"`
}

// LogoutAllResponse is the response of POST /v1/auth/logout-all
type LogoutAllResponse struct {
	Message         string `json:"message"`
	RevokedSessions int64  `json:"revoked_sessions"`
}

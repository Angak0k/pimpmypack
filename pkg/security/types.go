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

// RefreshResponse represents the response from refresh endpoint
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

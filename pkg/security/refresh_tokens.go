package security

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
)

// CreateRefreshToken creates a new refresh token for a user
func CreateRefreshToken(ctx context.Context, accountID uint, rememberMe bool) (*RefreshToken, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}
	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)

	// Calculate expiration
	var expiresAt time.Time
	if rememberMe {
		expiresAt = time.Now().Add(time.Hour * 24 * time.Duration(config.RefreshTokenRememberMeDays))
	} else {
		expiresAt = time.Now().Add(time.Hour * 24 * time.Duration(config.RefreshTokenDays))
	}

	// Insert into database
	var token RefreshToken
	err := database.DB().QueryRowContext(ctx,
		`INSERT INTO refresh_token (token, account_id, expires_at, created_at)
         VALUES ($1, $2, $3, $4)
         RETURNING id, token, account_id, expires_at, created_at, last_used_at, revoked`,
		tokenString, accountID, expiresAt, time.Now(),
	).Scan(
		&token.ID,
		&token.Token,
		&token.AccountID,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.LastUsedAt,
		&token.Revoked,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &token, nil
}

// GetRefreshToken retrieves a refresh token by token string
func GetRefreshToken(ctx context.Context, tokenString string) (*RefreshToken, error) {
	var token RefreshToken

	err := database.DB().QueryRowContext(ctx,
		`SELECT id, token, account_id, expires_at, created_at, last_used_at, revoked
         FROM refresh_token
         WHERE token = $1`,
		tokenString,
	).Scan(
		&token.ID,
		&token.Token,
		&token.AccountID,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.LastUsedAt,
		&token.Revoked,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("refresh token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &token, nil
}

// UpdateLastUsed updates the last_used_at timestamp
func UpdateLastUsed(ctx context.Context, tokenID uint) error {
	_, err := database.DB().ExecContext(ctx,
		`UPDATE refresh_token SET last_used_at = $1 WHERE id = $2`,
		time.Now(), tokenID,
	)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}
	return nil
}

// DeleteRefreshToken deletes a refresh token (revocation)
func DeleteRefreshToken(ctx context.Context, tokenString string) error {
	result, err := database.DB().ExecContext(ctx,
		`DELETE FROM refresh_token WHERE token = $1`,
		tokenString,
	)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("refresh token not found")
	}

	return nil
}

// CleanupExpiredTokens deletes expired refresh tokens
func CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := database.DB().ExecContext(ctx,
		`DELETE FROM refresh_token WHERE expires_at < $1`,
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to check rows affected: %w", err)
	}

	return rowsAffected, nil
}

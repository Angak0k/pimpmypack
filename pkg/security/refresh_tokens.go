package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
)

// hashRefreshToken returns the hex-encoded SHA-256 of a refresh token —
// the only form ever persisted, so a database leak cannot be replayed
func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// CreateRefreshToken creates a new refresh token for a user
func CreateRefreshToken(ctx context.Context, accountID uint, rememberMe bool) (*RefreshToken, error) {
	var expiresAt time.Time
	if rememberMe {
		expiresAt = time.Now().Add(time.Hour * 24 * time.Duration(config.RefreshTokenRememberMeDays))
	} else {
		expiresAt = time.Now().Add(time.Hour * 24 * time.Duration(config.RefreshTokenDays))
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}
	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)

	var token RefreshToken
	err := database.DB().QueryRowContext(ctx,
		`INSERT INTO refresh_token (token, account_id, expires_at, created_at)
         VALUES ($1, $2, $3, $4)
         RETURNING id, token, account_id, expires_at, created_at, last_used_at, revoked`,
		hashRefreshToken(tokenString), accountID, expiresAt, time.Now(),
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

	// The database holds only the hash; the caller needs the plaintext to
	// hand to the client — this is the only place it exists server-side
	token.Token = tokenString

	return &token, nil
}

// GetRefreshToken retrieves a refresh token by token string
func GetRefreshToken(ctx context.Context, tokenString string) (*RefreshToken, error) {
	var token RefreshToken

	err := database.DB().QueryRowContext(ctx,
		`SELECT id, token, account_id, expires_at, created_at, last_used_at, revoked
         FROM refresh_token
         WHERE token = $1`,
		hashRefreshToken(tokenString),
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

// RevokeRefreshToken marks a refresh token as revoked. It returns the owning
// account ID and whether a live token matched, so callers can audit the event
// yet respond generically either way (no token enumeration).
func RevokeRefreshToken(ctx context.Context, tokenString string) (uint, bool, error) {
	var accountID uint
	err := database.DB().QueryRowContext(ctx,
		`UPDATE refresh_token SET revoked = TRUE
         WHERE token = $1 AND revoked = FALSE
         RETURNING account_id`,
		hashRefreshToken(tokenString),
	).Scan(&accountID)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return accountID, true, nil
}

// RevokeAllUserTokens revokes every live refresh token of an account and
// returns how many sessions were revoked
func RevokeAllUserTokens(ctx context.Context, accountID uint) (int64, error) {
	result, err := database.DB().ExecContext(ctx,
		`UPDATE refresh_token SET revoked = TRUE
         WHERE account_id = $1 AND revoked = FALSE`,
		accountID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to revoke user refresh tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to check rows affected: %w", err)
	}

	return rowsAffected, nil
}

// CleanupExpiredTokens deletes refresh tokens that can never be valid again:
// expired ones and revoked ones (logout, logout-all, password change)
func CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := database.DB().ExecContext(ctx,
		`DELETE FROM refresh_token WHERE expires_at < $1 OR revoked = TRUE`,
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

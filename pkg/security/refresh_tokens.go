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

// ErrTokenAlreadyRotated is returned when a rotation loses the race against
// a concurrent rotation of the same token
var ErrTokenAlreadyRotated = errors.New("refresh token already rotated")

// hashRefreshToken returns the hex-encoded SHA-256 of a refresh token —
// the only form ever persisted, so a database leak cannot be replayed
func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// newTokenString generates a cryptographically random refresh token
func newTokenString() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

// queryRower is the QueryRowContext surface shared by *sql.DB and *sql.Tx,
// so insertRefreshToken works both standalone and inside a transaction.
type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// insertRefreshToken inserts a token row (storing only its hash) and returns
// the populated model with the plaintext token set for the caller to hand out
func insertRefreshToken(
	ctx context.Context, q queryRower, plaintext string, accountID uint, expiresAt time.Time,
) (*RefreshToken, error) {
	var token RefreshToken
	err := q.QueryRowContext(ctx,
		`INSERT INTO refresh_token (token, account_id, expires_at, created_at)
         VALUES ($1, $2, $3, $4)
         RETURNING id, token, account_id, expires_at, created_at, last_used_at, revoked, rotated_at`,
		hashRefreshToken(plaintext), accountID, expiresAt, time.Now(),
	).Scan(
		&token.ID,
		&token.Token,
		&token.AccountID,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.LastUsedAt,
		&token.Revoked,
		&token.RotatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert refresh token: %w", err)
	}

	// The database holds only the hash; the plaintext exists server-side only
	// here, to be handed to the client
	token.Token = plaintext
	return &token, nil
}

// CreateRefreshToken creates a new refresh token for a user
func CreateRefreshToken(ctx context.Context, accountID uint, rememberMe bool) (*RefreshToken, error) {
	var expiresAt time.Time
	if rememberMe {
		expiresAt = time.Now().Add(time.Hour * 24 * time.Duration(config.RefreshTokenRememberMeDays))
	} else {
		expiresAt = time.Now().Add(time.Hour * 24 * time.Duration(config.RefreshTokenDays))
	}

	tokenString, err := newTokenString()
	if err != nil {
		return nil, err
	}

	return insertRefreshToken(ctx, database.DB(), tokenString, accountID, expiresAt)
}

// GetRefreshToken retrieves a refresh token by token string
func GetRefreshToken(ctx context.Context, tokenString string) (*RefreshToken, error) {
	var token RefreshToken

	err := database.DB().QueryRowContext(ctx,
		`SELECT id, token, account_id, expires_at, created_at, last_used_at, revoked, rotated_at
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
		&token.RotatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("refresh token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &token, nil
}

// HasLiveRefreshToken reports whether the account still has at least one
// non-revoked, unexpired refresh token — used to deny the rotation grace
// once a logout/password-change/admin action has cleared the sessions.
func HasLiveRefreshToken(ctx context.Context, accountID uint) (bool, error) {
	var exists bool
	err := database.DB().QueryRowContext(ctx,
		`SELECT EXISTS(
            SELECT 1 FROM refresh_token
            WHERE account_id = $1 AND revoked = FALSE AND expires_at > $2)`,
		accountID, time.Now(),
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check for live refresh token: %w", err)
	}
	return exists, nil
}

// RotateRefreshToken atomically revokes the presented token and issues its
// successor, preserving the session horizon. Returns ErrTokenAlreadyRotated
// when a concurrent rotation of the same token won the race.
func RotateRefreshToken(ctx context.Context, old *RefreshToken) (*RefreshToken, error) {
	tokenString, err := newTokenString()
	if err != nil {
		return nil, err
	}

	tx, err := database.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin rotation transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op once committed

	now := time.Now()
	result, err := tx.ExecContext(ctx,
		`UPDATE refresh_token SET revoked = TRUE, rotated_at = $1, last_used_at = $1
         WHERE id = $2 AND revoked = FALSE AND expires_at > $1`,
		now, old.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke rotated token: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, ErrTokenAlreadyRotated
	}

	successor, err := insertRefreshToken(ctx, tx, tokenString, old.AccountID, old.ExpiresAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit rotation: %w", err)
	}

	return successor, nil
}

// RevokeRefreshToken marks a refresh token as revoked. It returns the owning
// account ID and whether a live token matched, so callers can audit the event
// yet respond generically either way (no token enumeration).
func RevokeRefreshToken(ctx context.Context, tokenString string) (uint, bool, error) {
	var accountID uint
	err := database.DB().QueryRowContext(ctx,
		`UPDATE refresh_token SET revoked = TRUE
         WHERE token = $1 AND revoked = FALSE AND expires_at > $2
         RETURNING account_id`,
		hashRefreshToken(tokenString), time.Now(),
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
         WHERE account_id = $1 AND revoked = FALSE AND expires_at > $2`,
		accountID, time.Now(),
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

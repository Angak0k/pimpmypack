package security

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Initialize configuration
	if err := config.EnvInit("../../.env"); err != nil {
		panic(err)
	}

	// Initialize database
	if err := database.Initialization(); err != nil {
		panic(err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	os.Exit(code)
}

// Test helper to create test account
func createTestAccount(t *testing.T) uint {
	var accountID uint
	now := time.Now()
	// Use nanoseconds for uniqueness
	username := "testuser_" + now.Format("20060102150405") + "_" + strconv.FormatInt(now.UnixNano(), 10)

	// Create account
	err := database.DB().QueryRow(
		`INSERT INTO account (username, email, firstname, lastname, role, status, created_at, updated_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		username,
		username+"@example.com", // unique email too
		"Test",
		"User",
		"standard",
		"active",
		now,
		now,
	).Scan(&accountID)
	require.NoError(t, err)

	// Create password entry
	hashedPassword, err := HashPassword("testpassword")
	require.NoError(t, err)

	_, err = database.DB().Exec(
		`INSERT INTO password (user_id, password, updated_at)
         VALUES ($1, $2, $3)`,
		accountID,
		hashedPassword,
		now,
	)
	require.NoError(t, err)

	return accountID
}

func TestCreateRefreshToken_Default(t *testing.T) {
	ctx := context.Background()
	accountID := createTestAccount(t)

	token, err := CreateRefreshToken(ctx, accountID, false)

	require.NoError(t, err)
	assert.NotNil(t, token)
	assert.NotEmpty(t, token.Token)
	assert.Equal(t, accountID, token.AccountID)
	assert.False(t, token.Revoked)

	// Check expiration is approximately 24 hours (use 2 hour tolerance for timezone differences)
	expectedExpiry := time.Now().UTC().Add(time.Hour * 24 * time.Duration(config.RefreshTokenDays))
	assert.WithinDuration(t, expectedExpiry, token.ExpiresAt, 2*time.Hour)
}

func TestCreateRefreshToken_RememberMe(t *testing.T) {
	ctx := context.Background()
	accountID := createTestAccount(t)

	token, err := CreateRefreshToken(ctx, accountID, true)

	require.NoError(t, err)
	assert.NotNil(t, token)

	// Check expiration is approximately 30 days (use 2 hour tolerance for timezone differences)
	expectedExpiry := time.Now().UTC().Add(time.Hour * 24 * time.Duration(config.RefreshTokenRememberMeDays))
	assert.WithinDuration(t, expectedExpiry, token.ExpiresAt, 2*time.Hour)
}

func TestGetRefreshToken_Success(t *testing.T) {
	ctx := context.Background()
	accountID := createTestAccount(t)

	created, err := CreateRefreshToken(ctx, accountID, false)
	require.NoError(t, err)

	retrieved, err := GetRefreshToken(ctx, created.Token)

	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Token, retrieved.Token)
	assert.Equal(t, created.AccountID, retrieved.AccountID)
}

func TestGetRefreshToken_NotFound(t *testing.T) {
	ctx := context.Background()

	token, err := GetRefreshToken(ctx, "nonexistent-token")

	require.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateLastUsed(t *testing.T) {
	ctx := context.Background()
	accountID := createTestAccount(t)

	token, err := CreateRefreshToken(ctx, accountID, false)
	require.NoError(t, err)
	assert.Nil(t, token.LastUsedAt)

	time.Sleep(time.Second) // Ensure time difference

	err = UpdateLastUsed(ctx, token.ID)
	require.NoError(t, err)

	updated, err := GetRefreshToken(ctx, token.Token)
	require.NoError(t, err)
	assert.NotNil(t, updated.LastUsedAt)
}

func TestDeleteRefreshToken_Success(t *testing.T) {
	ctx := context.Background()
	accountID := createTestAccount(t)

	token, err := CreateRefreshToken(ctx, accountID, false)
	require.NoError(t, err)

	err = DeleteRefreshToken(ctx, token.Token)
	require.NoError(t, err)

	// Verify it's deleted
	retrieved, err := GetRefreshToken(ctx, token.Token)
	require.Error(t, err)
	assert.Nil(t, retrieved)
}

func TestDeleteRefreshToken_NotFound(t *testing.T) {
	ctx := context.Background()

	err := DeleteRefreshToken(ctx, "nonexistent-token")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCleanupExpiredTokens(t *testing.T) {
	ctx := context.Background()
	accountID := createTestAccount(t)

	// Create an expired token by manually inserting
	_, err := database.DB().Exec(
		`INSERT INTO refresh_token (token, account_id, expires_at, created_at)
         VALUES ($1, $2, $3, $4)`,
		"expired-token-123",
		accountID,
		time.Now().Add(-time.Hour), // expired 1 hour ago
		time.Now().Add(-25*time.Hour),
	)
	require.NoError(t, err)

	// Create a valid token
	validToken, err := CreateRefreshToken(ctx, accountID, false)
	require.NoError(t, err)

	// Run cleanup
	deleted, err := CleanupExpiredTokens(ctx)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, deleted, int64(1))

	// Verify expired token is gone
	_, err = GetRefreshToken(ctx, "expired-token-123")
	require.Error(t, err)

	// Verify valid token still exists
	retrieved, err := GetRefreshToken(ctx, validToken.Token)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
}

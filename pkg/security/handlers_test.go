package security

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/auth/refresh", RefreshTokenHandler)
	r.POST("/auth/logout", LogoutHandler)
	r.POST("/v1/auth/logout-all", JwtAuthProcessor(), LogoutAllHandler)
	return r
}

func postJSON(router *gin.Engine, path string, payload any, accessToken string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestRefreshTokenHandler_Success(t *testing.T) {
	ctx := context.Background()
	accountID := createTestAccount(t)
	router := setupTestRouter()

	// Create a valid refresh token
	refreshToken, err := CreateRefreshToken(ctx, accountID, false)
	require.NoError(t, err)

	w := postJSON(router, "/auth/refresh", RefreshTokenInput{Token: refreshToken.Token}, "")

	assert.Equal(t, http.StatusOK, w.Code)

	var response RefreshResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.Positive(t, response.ExpiresIn)
}

func TestRefreshTokenHandler_InvalidToken(t *testing.T) {
	router := setupTestRouter()

	w := postJSON(router, "/auth/refresh", RefreshTokenInput{Token: "invalid-token"}, "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid refresh token")
}

func TestRefreshTokenHandler_ExpiredToken(t *testing.T) {
	accountID := createTestAccount(t)
	router := setupTestRouter()

	// Create an expired token manually with unique token string
	now := time.Now()
	expiredToken := "expired-token-" + now.Format("20060102150405") + "-" + strconv.FormatInt(now.UnixNano(), 10)
	_, err := database.DB().ExecContext(context.Background(),
		`INSERT INTO refresh_token (token, account_id, expires_at, created_at)
         VALUES ($1, $2, $3, $4)`,
		hashRefreshToken(expiredToken),
		accountID,
		time.Now().Add(-time.Hour),
		time.Now().Add(-25*time.Hour),
	)
	require.NoError(t, err)

	w := postJSON(router, "/auth/refresh", RefreshTokenInput{Token: expiredToken}, "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "expired")
}

func TestRefreshTokenHandler_MissingInput(t *testing.T) {
	router := setupTestRouter()

	w := postJSON(router, "/auth/refresh", struct{}{}, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogoutHandler_Success(t *testing.T) {
	ctx := context.Background()
	accountID := createTestAccount(t)
	router := setupTestRouter()

	refreshToken, err := CreateRefreshToken(ctx, accountID, false)
	require.NoError(t, err)

	w := postJSON(router, "/auth/logout", RefreshTokenInput{Token: refreshToken.Token}, "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Logged out successfully")

	// The revoked token must no longer refresh
	w = postJSON(router, "/auth/refresh", RefreshTokenInput{Token: refreshToken.Token}, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "revoked")
}

func TestLogoutHandler_InvalidTokenStillSucceeds(t *testing.T) {
	router := setupTestRouter()

	// 200 even for unknown tokens: no token enumeration
	w := postJSON(router, "/auth/logout", RefreshTokenInput{Token: "unknown-token"}, "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Logged out successfully")
}

func TestLogoutHandler_MissingInput(t *testing.T) {
	router := setupTestRouter()

	w := postJSON(router, "/auth/logout", struct{}{}, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogoutAllHandler_Success(t *testing.T) {
	ctx := context.Background()
	accountID := createTestAccount(t)
	router := setupTestRouter()

	token1, err := CreateRefreshToken(ctx, accountID, false)
	require.NoError(t, err)
	token2, err := CreateRefreshToken(ctx, accountID, true)
	require.NoError(t, err)

	accessToken, err := GenerateToken(accountID)
	require.NoError(t, err)

	w := postJSON(router, "/v1/auth/logout-all", struct{}{}, accessToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var response LogoutAllResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, "Logged out from all devices", response.Message)
	assert.Equal(t, int64(2), response.RevokedSessions)

	// Both sessions are dead
	for _, tok := range []string{token1.Token, token2.Token} {
		w = postJSON(router, "/auth/refresh", RefreshTokenInput{Token: tok}, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}

func TestLogoutAllHandler_RequiresAuth(t *testing.T) {
	router := setupTestRouter()

	w := postJSON(router, "/v1/auth/logout-all", struct{}{}, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

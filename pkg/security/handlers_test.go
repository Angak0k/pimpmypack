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
	return r
}

func TestRefreshTokenHandler_Success(t *testing.T) {
	ctx := context.Background()
	accountID := createTestAccount(t)
	router := setupTestRouter()

	// Create a valid refresh token
	refreshToken, err := CreateRefreshToken(ctx, accountID, false)
	require.NoError(t, err)

	// Make request
	input := RefreshTokenInput{Token: refreshToken.Token}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response RefreshResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.Positive(t, response.ExpiresIn)
}

func TestRefreshTokenHandler_InvalidToken(t *testing.T) {
	router := setupTestRouter()

	input := RefreshTokenInput{Token: "invalid-token"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid refresh token")
}

func TestRefreshTokenHandler_ExpiredToken(t *testing.T) {
	accountID := createTestAccount(t)
	router := setupTestRouter()

	// Create an expired token manually with unique token string
	now := time.Now()
	expiredToken := "expired-token-" + now.Format("20060102150405") + "-" + strconv.FormatInt(now.UnixNano(), 10)
	_, err := database.DB().Exec(
		`INSERT INTO refresh_token (token, account_id, expires_at, created_at)
         VALUES ($1, $2, $3, $4)`,
		expiredToken,
		accountID,
		time.Now().Add(-time.Hour),
		time.Now().Add(-25*time.Hour),
	)
	require.NoError(t, err)

	// Make request
	input := RefreshTokenInput{Token: expiredToken}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "expired")
}

func TestRefreshTokenHandler_MissingInput(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

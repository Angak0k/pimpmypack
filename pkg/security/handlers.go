package security

import (
	"net/http"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/gin-gonic/gin"
)

// RefreshTokenHandler handles POST /auth/refresh
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new access token
// @Tags authentication
// @Accept json
// @Produce json
// @Param refresh_token body RefreshTokenInput true "Refresh Token"
// @Success 200 {object} RefreshResponse
// @Failure 400 {object} apitypes.ErrorResponse "Invalid input"
// @Failure 401 {object} apitypes.ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} apitypes.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func RefreshTokenHandler(c *gin.Context) {
	var input RefreshTokenInput

	// 1. Bind JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Get refresh token from database
	refreshToken, err := GetRefreshToken(c.Request.Context(), input.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// 3. Validate refresh token
	if refreshToken.Revoked {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has been revoked"})
		return
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has expired"})
		return
	}

	// 4. Generate new access token
	accessToken, err := GenerateToken(refreshToken.AccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	// 5. Update last_used_at (ignore errors - non-blocking)
	_ = UpdateLastUsed(c.Request.Context(), refreshToken.ID)

	// 6. Respond with new access token
	c.JSON(http.StatusOK, RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(config.AccessTokenMinutes * 60),
	})
}

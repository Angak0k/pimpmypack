package security

import (
	"net/http"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/gin-gonic/gin"
)

// RefreshTokenHandler handles POST /auth/refresh
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new access token
// @Tags Authentication
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
		helper.LogAndSanitize(err, "refresh token: bind JSON failed")
		AuditRefreshFailed(c, "invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	// 2. Get refresh token from database
	refreshToken, err := GetRefreshToken(c.Request.Context(), input.Token)
	if err != nil {
		AuditRefreshFailed(c, "invalid refresh token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// 3. Validate refresh token
	if refreshToken.Revoked {
		AuditRefreshFailed(c, "refresh token has been revoked")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has been revoked"})
		return
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		AuditRefreshFailed(c, "refresh token has expired")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has expired"})
		return
	}

	// 4. Generate new access token
	accessToken, err := GenerateToken(refreshToken.AccountID)
	if err != nil {
		AuditRefreshFailed(c, "failed to generate access token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	// 5. Update last_used_at (ignore errors - non-blocking)
	_ = UpdateLastUsed(c.Request.Context(), refreshToken.ID)

	// 6. Issue a rotated refresh token preserving the session horizon.
	// Non-fatal: refresh still succeeds with the presented token if this
	// fails, since strict rotation is not enforced yet.
	response := RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(config.AccessTokenMinutes * 60),
	}
	rotated, err := createRefreshTokenExpiringAt(c.Request.Context(), refreshToken.AccountID, refreshToken.ExpiresAt)
	if err != nil {
		helper.LogAndSanitize(err, "refresh token: rotate refresh token failed")
	} else {
		response.RefreshToken = rotated.Token
		response.RefreshExpiresIn = int64(time.Until(rotated.ExpiresAt).Seconds())
	}

	// 7. Audit successful refresh and respond
	AuditRefreshSuccess(c, refreshToken.AccountID)
	c.JSON(http.StatusOK, response)
}

// LogoutHandler handles POST /auth/logout
// @Summary Logout (revoke a refresh token)
// @Description Revoke the presented refresh token, ending that session. Always 200: token existence cannot be probed.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refresh_token body RefreshTokenInput true "Refresh Token"
// @Success 200 {object} LogoutResponse
// @Failure 400 {object} apitypes.ErrorResponse "Invalid input"
// @Failure 500 {object} apitypes.ErrorResponse "Internal server error"
// @Router /auth/logout [post]
func LogoutHandler(c *gin.Context) {
	var input RefreshTokenInput

	if err := c.ShouldBindJSON(&input); err != nil {
		helper.LogAndSanitize(err, "logout: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	accountID, found, err := RevokeRefreshToken(c.Request.Context(), input.Token)
	if err != nil {
		helper.LogAndSanitize(err, "logout: revoke refresh token failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}
	if found {
		AuditLogout(c, accountID)
	}

	// 200 even when no live token matched: no token enumeration
	c.JSON(http.StatusOK, LogoutResponse{Message: "Logged out successfully"})
}

// LogoutAllHandler handles POST /v1/auth/logout-all
// @Summary Logout from all devices
// @Description Revoke all refresh tokens of the authenticated user
// @Security Bearer
// @Tags Authentication
// @Produce json
// @Success 200 {object} LogoutAllResponse
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 500 {object} apitypes.ErrorResponse "Internal server error"
// @Router /v1/auth/logout-all [post]
func LogoutAllHandler(c *gin.Context) {
	userID, err := ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "logout all: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	revoked, err := RevokeAllUserTokens(c.Request.Context(), userID)
	if err != nil {
		helper.LogAndSanitize(err, "logout all: revoke user tokens failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	AuditLogoutAll(c, userID, revoked)

	c.JSON(http.StatusOK, LogoutAllResponse{
		Message:         "Logged out from all devices",
		RevokedSessions: revoked,
	})
}

package security

import (
	"errors"
	"net/http"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/apitypes"
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
	if time.Now().After(refreshToken.ExpiresAt) {
		AuditRefreshFailed(c, "refresh token has expired")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has expired"})
		return
	}

	// 4. A revoked token may still be within the rotation grace window
	if refreshToken.Revoked {
		handleRevokedRefreshToken(c, refreshToken)
		return
	}

	// 5. Generate the access token BEFORE rotating: if signing fails we must
	// not have already burned (revoked) the presented refresh token.
	accessToken, err := GenerateToken(refreshToken.AccountID)
	if err != nil {
		AuditRefreshFailed(c, "failed to generate access token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	// 6. Strict rotation: revoke the presented token and issue its successor.
	successor, err := RotateRefreshToken(c.Request.Context(), refreshToken)
	switch {
	case err == nil:
		AuditRefreshSuccess(c, refreshToken.AccountID)
		c.JSON(http.StatusOK, RefreshResponse{
			AccessToken:      accessToken,
			ExpiresIn:        int64(config.AccessTokenMinutes * 60),
			RefreshToken:     successor.Token,
			RefreshExpiresIn: max(0, int64(time.Until(successor.ExpiresAt).Seconds())),
		})
	case errors.Is(err, ErrTokenAlreadyRotated):
		// The row changed between the read and the rotation UPDATE (a
		// concurrent rotation, a logout/password/admin revocation, or a
		// just-crossed expiry). Re-read and apply the current policy rather
		// than blindly granting access.
		reevaluateRefreshAfterRace(c, input.Token)
	default:
		AuditRefreshFailed(c, "failed to rotate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
	}
}

// reevaluateRefreshAfterRace re-reads a token that could not be rotated
// (the row was revoked/rotated/expired concurrently) and re-applies the full
// validation, so a benign rotation race still yields access while a
// concurrent revocation or expiry correctly returns 401.
func reevaluateRefreshAfterRace(c *gin.Context, tokenString string) {
	latest, err := GetRefreshToken(c.Request.Context(), tokenString)
	if err != nil {
		AuditRefreshFailed(c, "invalid refresh token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	if time.Now().After(latest.ExpiresAt) {
		AuditRefreshFailed(c, "refresh token has expired")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has expired"})
		return
	}
	handleRevokedRefreshToken(c, latest)
}

// handleRevokedRefreshToken answers a refresh attempt with a revoked token.
// Rotation deliberately does NOT auto-revoke sibling sessions on replay: a
// superseded token replayed by a benign client (multi-tab race, a lost
// response on a flaky network, a backgrounded tab) is common, and an
// account-wide logout would be a worse outcome than a re-login.
//   - rotated within the grace window AND the account still has a live
//     session: benign race → issue an access token (no re-rotation)
//   - otherwise: plain 401. A replay past the grace window is logged as a
//     reuse signal for out-of-band alerting, but triggers no revocation.
func handleRevokedRefreshToken(c *gin.Context, refreshToken *RefreshToken) {
	graceWindow := time.Duration(config.RefreshRotationGraceSeconds) * time.Second
	if refreshToken.RotatedAt != nil && time.Since(*refreshToken.RotatedAt) <= graceWindow {
		// Guard the grace path against a concurrent logout / password change /
		// admin action: if no live session remains, the account was revoked
		// and the grace must not resurrect access.
		live, err := HasLiveRefreshToken(c.Request.Context(), refreshToken.AccountID)
		if err != nil {
			helper.LogAndSanitize(err, "refresh grace: live-token check failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}
		if live {
			respondWithAccessToken(c, refreshToken.AccountID)
			return
		}
	}

	if refreshToken.RotatedAt != nil {
		// Superseded token replayed beyond grace: log for alerting, no action
		AuditRefreshReuseDetected(c, refreshToken.AccountID)
	} else {
		AuditRefreshFailed(c, "refresh token has been revoked")
	}
	c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
}

// respondWithAccessToken issues a new access token for an already-valid
// session (rotation-race loser or benign grace replay) without rotating
func respondWithAccessToken(c *gin.Context, accountID uint) {
	accessToken, err := GenerateToken(accountID)
	if err != nil {
		AuditRefreshFailed(c, "failed to generate access token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}
	AuditRefreshSuccess(c, accountID)
	c.JSON(http.StatusOK, RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(config.AccessTokenMinutes * 60),
	})
}

// LogoutHandler handles POST /auth/logout
// @Summary Logout (revoke a refresh token)
// @Description Revoke the presented refresh token, ending that session. Always 200: token existence cannot be probed.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refresh_token body RefreshTokenInput true "Refresh Token"
// @Success 200 {object} apitypes.OkResponse
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
	c.JSON(http.StatusOK, apitypes.OkResponse{Response: "Logged out successfully"})
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

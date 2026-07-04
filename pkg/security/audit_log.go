package security

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	// Authentication events
	EventLoginSuccess      AuditEventType = "login_success"
	EventLoginFailed       AuditEventType = "login_failed"
	EventRefreshSuccess    AuditEventType = "refresh_success"
	EventRefreshFailed     AuditEventType = "refresh_failed"
	EventLogout            AuditEventType = "logout_success"
	EventLogoutAll         AuditEventType = "logout_all_success"
	EventRefreshReuse      AuditEventType = "refresh_token_reuse_detected"
	EventRateLimitExceeded AuditEventType = "rate_limit_exceeded"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	Timestamp  time.Time      `json:"timestamp"`
	EventType  AuditEventType `json:"event_type"`
	UserID     *uint          `json:"user_id,omitempty"`
	Username   string         `json:"username,omitempty"`
	IP         string         `json:"ip"`
	UserAgent  string         `json:"user_agent,omitempty"`
	Message    string         `json:"message"`
	RememberMe bool           `json:"remember_me,omitempty"`
}

// logAuditEvent logs an audit event as structured JSON
func logAuditEvent(event AuditEvent) {
	event.Timestamp = time.Now().UTC()

	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal audit event: %v", err)
		return
	}

	log.Printf("[AUDIT] %s", string(jsonData))
}

// AuditLoginSuccess logs a successful login attempt
func AuditLoginSuccess(c *gin.Context, userID uint, rememberMe bool) {
	logAuditEvent(AuditEvent{
		EventType:  EventLoginSuccess,
		UserID:     &userID,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Message:    "User logged in successfully",
		RememberMe: rememberMe,
	})
}

// AuditLoginFailed logs a failed login attempt
func AuditLoginFailed(c *gin.Context, username, reason string) {
	logAuditEvent(AuditEvent{
		EventType: EventLoginFailed,
		Username:  username,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Message:   reason,
	})
}

// AuditRefreshSuccess logs a successful token refresh
func AuditRefreshSuccess(c *gin.Context, userID uint) {
	logAuditEvent(AuditEvent{
		EventType: EventRefreshSuccess,
		UserID:    &userID,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Message:   "Access token refreshed successfully",
	})
}

// AuditRefreshFailed logs a failed token refresh attempt
func AuditRefreshFailed(c *gin.Context, reason string) {
	logAuditEvent(AuditEvent{
		EventType: EventRefreshFailed,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Message:   reason,
	})
}

// AuditLogout logs a logout event
func AuditLogout(c *gin.Context, userID uint) {
	logAuditEvent(AuditEvent{
		EventType: EventLogout,
		UserID:    &userID,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Message:   "User logged out",
	})
}

// AuditLogoutAll logs a logout-from-all-devices event
func AuditLogoutAll(c *gin.Context, userID uint, revokedSessions int64) {
	logAuditEvent(AuditEvent{
		EventType: EventLogoutAll,
		UserID:    &userID,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Message:   fmt.Sprintf("User logged out from all devices (%d sessions revoked)", revokedSessions),
	})
}

// AuditRefreshReuseDetected logs the replay of a superseded refresh token
// beyond the rotation grace window — a possible token-theft signal surfaced
// for out-of-band alerting (no automatic revocation is performed)
func AuditRefreshReuseDetected(c *gin.Context, userID uint) {
	logAuditEvent(AuditEvent{
		EventType: EventRefreshReuse,
		UserID:    &userID,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Message:   "Superseded refresh token replayed beyond grace window",
	})
}

// AuditRateLimitExceeded logs a rate limit exceeded event
func AuditRateLimitExceeded(c *gin.Context, endpoint string) {
	logAuditEvent(AuditEvent{
		EventType: EventRateLimitExceeded,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Message:   "Rate limit exceeded for " + endpoint,
	})
}

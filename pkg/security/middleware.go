package security

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/gin-gonic/gin"
)

// MaxBodyBytes limits the size of the request body. Requests declaring a
// larger Content-Length are rejected with 413; bodies without a declared
// length are capped while reading via http.MaxBytesReader (the read error
// surfaces in the handler, typically as a 400 on binding).
func MaxBodyBytes(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > limit {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}

// Headers sets defense-in-depth response headers on every response.
// HSTS is intentionally left to the reverse proxy / load balancer.
func Headers() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Next()
	}
}

// JwtAuthProcessor validates JWT tokens (existing function, moved here)
func JwtAuthProcessor() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := TokenValid(c)
		if err != nil {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

// JwtAuthAdminProcessor validates JWT and checks admin role (existing function, moved here).
// Alg-confusion and key-source attacks are handled by TokenValid (via jwtKeyFunc + WithValidMethods).
func JwtAuthAdminProcessor() gin.HandlerFunc {
	return func(c *gin.Context) {
		// check if token is valid
		err := TokenValid(c)
		if err != nil {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		// get user_id from token
		userID, err := ExtractTokenID(c)
		if err != nil {
			c.String(http.StatusUnauthorized, "Invalid Token")
			c.Abort()
			return
		}

		// check if user is admin
		var role string
		row := database.DB().QueryRowContext(context.Background(), "SELECT role FROM account WHERE id = $1;", userID)
		err = row.Scan(&role)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.String(http.StatusUnauthorized, "Unauthorized")
				c.Abort()
				return
			}
			c.String(http.StatusInternalServerError, "Something goes wrong")
			c.Abort()
			return
		}
		if role != "admin" {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		c.Next()
	}
}

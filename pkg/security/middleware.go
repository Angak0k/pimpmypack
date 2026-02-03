package security

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/gin-gonic/gin"
)

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

// JwtAuthAdminProcessor validates JWT and checks admin role (existing function, moved here)
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

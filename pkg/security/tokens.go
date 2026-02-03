package security

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken generates a JWT access token (existing function, moved here)
func GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"authorized": true,
		"user_id":    userID,
		"exp":        time.Now().Add(time.Minute * time.Duration(config.AccessTokenMinutes)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(config.APISecret))
}

// TokenValid validates a JWT token (existing function, moved here)
func TokenValid(c *gin.Context) error {
	tokenString := ExtractToken(c)
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.APISecret), nil
	})
	if err != nil {
		return err
	}
	return nil
}

// ExtractToken extracts token from header or query (existing function, moved here)
func ExtractToken(c *gin.Context) string {
	token := c.Query("token")
	if token != "" {
		return token
	}
	bearerToken := c.Request.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

// ExtractTokenID extracts user ID from token (existing function, moved here)
func ExtractTokenID(c *gin.Context) (uint, error) {
	tokenString := ExtractToken(c)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		uid, err := strconv.ParseUint(fmt.Sprintf("%.0f", claims["user_id"]), 10, 32)
		if err != nil {
			return 0, err
		}
		return uint(uid), nil
	}
	return 0, nil
}

// GenerateTokenPair generates both access and refresh tokens (NEW)
func GenerateTokenPair(ctx context.Context, accountID uint, rememberMe bool) (*TokenPairResponse, error) {
	// Generate access token
	accessToken, err := GenerateToken(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := CreateRefreshToken(ctx, accountID, rememberMe)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Calculate expiration times
	accessExpiresIn := int64(config.AccessTokenMinutes * 60)
	var refreshExpiresIn int64
	if rememberMe {
		refreshExpiresIn = int64(config.RefreshTokenRememberMeDays * 24 * 3600)
	} else {
		refreshExpiresIn = int64(config.RefreshTokenDays * 24 * 3600)
	}

	return &TokenPairResponse{
		Token:            accessToken, // Backward compatibility
		AccessToken:      accessToken,
		RefreshToken:     refreshToken.Token,
		AccessExpiresIn:  accessExpiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}

package security

import (
	"golang.org/x/crypto/bcrypt"
)

// Security Definitions:
// securityDefinitions:
//   Bearer:
//     type: apiKey
//     name: Authorization
//     in: header

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), err
}

// VerifyPassword verifies a password against a bcrypt hash
func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

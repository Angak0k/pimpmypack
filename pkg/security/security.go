package security

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// hashCost is the bcrypt work factor. Test binaries use the minimum cost,
// since bcrypt at the default cost dominates test runtime under -race.
func hashCost() int {
	if testing.Testing() {
		return bcrypt.MinCost
	}
	return bcrypt.DefaultCost
}

// Security Definitions:
// securityDefinitions:
//   Bearer:
//     type: apiKey
//     name: Authorization
//     in: header

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), hashCost())
	if err != nil {
		return "", err
	}
	return string(hashedPassword), err
}

// VerifyPassword verifies a password against a bcrypt hash
func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

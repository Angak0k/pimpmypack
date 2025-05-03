package security

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const testAPISecret = "test_secret"
const testTokenLifespan = 24 // 24 hours for tests

func setupTestEnv(t *testing.T) {
	// Set up test environment
	os.Setenv("API_SECRET", testAPISecret)
	config.APISecret = testAPISecret
	config.TokenLifespan = testTokenLifespan
}

func teardownTestEnv(t *testing.T) {
	os.Unsetenv("API_SECRET")
	config.APISecret = ""
	config.TokenLifespan = 1 // Reset to default
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, got)
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	firstHashedPassword, err := HashPassword("password")
	require.NoError(t, err)

	secondHashedPassword, err := HashPassword("password1")
	require.NoError(t, err)

	tests := []struct {
		name           string
		password       string
		hashedPassword string
		wantErr        bool
		wantErrType    error
	}{
		{
			name:           "valid password",
			password:       "password",
			hashedPassword: firstHashedPassword,
			wantErr:        false,
		},
		{
			name:           "invalid password",
			password:       "password",
			hashedPassword: secondHashedPassword,
			wantErr:        true,
			wantErrType:    bcrypt.ErrMismatchedHashAndPassword,
		},
		{
			name:           "empty password",
			password:       "",
			hashedPassword: firstHashedPassword,
			wantErr:        true,
			wantErrType:    bcrypt.ErrMismatchedHashAndPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword(tt.password, tt.hashedPassword)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGenerateToken(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	tests := []struct {
		name    string
		userID  uint
		wantErr bool
	}{
		{
			name:    "valid user ID",
			userID:  123,
			wantErr: false,
		},
		{
			name:    "zero user ID",
			userID:  0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token structure
			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte(testAPISecret), nil
			})
			assert.NoError(t, err)
			assert.True(t, parsedToken.Valid)

			claims, ok := parsedToken.Claims.(jwt.MapClaims)
			assert.True(t, ok)
			assert.Equal(t, float64(tt.userID), claims["user_id"])
			assert.Equal(t, true, claims["authorized"])
		})
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func(*http.Request)
		expectedToken string
		expectedError bool
	}{
		{
			name: "token in query parameter",
			setupRequest: func(req *http.Request) {
				q := req.URL.Query()
				q.Add("token", "test_token")
				req.URL.RawQuery = q.Encode()
			},
			expectedToken: "test_token",
			expectedError: false,
		},
		{
			name: "token in authorization header",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer test_token")
			},
			expectedToken: "test_token",
			expectedError: false,
		},
		{
			name: "no token",
			setupRequest: func(req *http.Request) {
				// No setup needed
			},
			expectedToken: "",
			expectedError: false,
		},
		{
			name: "malformed authorization header",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "test_token")
			},
			expectedToken: "",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", "/", nil)
			tt.setupRequest(req)

			// Create a test context
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = req

			token := ExtractToken(c)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestExtractTokenID(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	tests := []struct {
		name          string
		setupToken    func() string
		expectedID    uint
		expectedError bool
	}{
		{
			name: "valid token",
			setupToken: func() string {
				token, _ := GenerateToken(123)
				return token
			},
			expectedID:    123,
			expectedError: false,
		},
		{
			name: "invalid token",
			setupToken: func() string {
				return "invalid_token"
			},
			expectedID:    0,
			expectedError: true,
		},
		{
			name: "expired token",
			setupToken: func() string {
				claims := jwt.MapClaims{
					"authorized": true,
					"user_id":    123,
					"exp":        time.Now().Add(-1 * time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(testAPISecret))
				return tokenString
			},
			expectedID:    0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request with the token
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+tt.setupToken())

			// Create a test context
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = req

			userID, err := ExtractTokenID(c)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, uint(0), userID)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedID, userID)
		})
	}
}

func TestJwtAuthProcessor(t *testing.T) {
	setupTestEnv(t)
	defer teardownTestEnv(t)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
	}{
		{
			name: "valid token",
			setupRequest: func(req *http.Request) {
				token, _ := GenerateToken(123)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid token",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalid_token")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "no token",
			setupRequest: func(req *http.Request) {
				// No setup needed
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", "/", nil)
			tt.setupRequest(req)

			// Create a test recorder
			w := httptest.NewRecorder()

			// Create a test context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Create a test handler
			handler := JwtAuthProcessor()

			// Create a test next handler
			nextCalled := false
			next := func(c *gin.Context) {
				nextCalled = true
				c.Status(http.StatusOK)
			}

			// Run the middleware
			handler(c)
			if c.IsAborted() {
				assert.Equal(t, tt.expectedStatus, w.Code)
				assert.False(t, nextCalled)
			} else {
				next(c)
				assert.Equal(t, http.StatusOK, w.Code)
				assert.True(t, nextCalled)
			}
		})
	}
}

package security

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newBodyLimitedRouter(limit int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/test", MaxBodyBytes(limit), func(c *gin.Context) {
		if _, err := io.ReadAll(c.Request.Body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	return router
}

func TestMaxBodyBytes_AllowsUnderLimit(t *testing.T) {
	router := newBodyLimitedRouter(100)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(strings.Repeat("a", 50))))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMaxBodyBytes_RejectsDeclaredOversizedBody(t *testing.T) {
	router := newBodyLimitedRouter(100)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(strings.Repeat("a", 101))))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Headers())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestMaxBodyBytes_CapsChunkedBody(t *testing.T) {
	router := newBodyLimitedRouter(100)

	// No declared Content-Length: the cap must apply while reading
	req := httptest.NewRequest(http.MethodPost, "/test", io.NopCloser(strings.NewReader(strings.Repeat("a", 101))))
	req.ContentLength = -1
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

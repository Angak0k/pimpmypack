package packs

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseLighterPackURL_InvalidURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/parselighterpackurl", ParseLighterPackURL)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com/not-lighterpack"})
	req, _ := http.NewRequest(http.MethodPost, "/parselighterpackurl", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestParseLighterPackURL_NoAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/parselighterpackurl", ParseLighterPackURL)

	body, _ := json.Marshal(map[string]string{"url": "ftp://bad"})
	req, _ := http.NewRequest(http.MethodPost, "/parselighterpackurl", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Errorf("endpoint must not require auth, got 401")
	}
}

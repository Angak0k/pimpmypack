package packs

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Angak0k/pimpmypack/pkg/security"
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

func TestImportPack_Success(t *testing.T) {
	token, err := security.GenerateToken(users[0].ID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/importpack", ImportPack)

	payload := ParseExternalPackResponse{
		PackName:        "Imported Pack",
		PackDescription: "from anonymous flow",
		Items: ExternalPack{
			{
				ItemName: "Tent",
				Category: "Shelter",
				Qty:      1,
				Weight:   1200,
				Unit:     "g",
				Currency: "EUR",
			},
		},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/importpack", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	verifyImportResponse(t, w.Body.Bytes())
}

func TestImportPack_EmptyItems(t *testing.T) {
	token, err := security.GenerateToken(users[0].ID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/importpack", ImportPack)

	payload := ParseExternalPackResponse{
		PackName:        "Empty Pack",
		PackDescription: "no items",
		Items:           ExternalPack{},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/importpack", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d (body: %s)", w.Code, w.Body.String())
	}
}

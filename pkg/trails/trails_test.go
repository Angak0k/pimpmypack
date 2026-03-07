package trails

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
)

func TestMain(m *testing.M) {
	err := config.EnvInit("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file or environment variable : %v", err)
	}
	println("Environment variables loaded")

	err = database.Initialization()
	if err != nil {
		log.Fatalf("Error connecting database : %v", err)
	}
	println("Database connected")

	err = database.Migrate()
	if err != nil {
		log.Fatalf("Error migrating database : %v", err)
	}
	println("Database migrated")

	println("Loading dataset...")
	err = loadingTrailDataset()
	if err != nil {
		log.Fatalf("Error loading dataset : %v", err)
	}
	println("Dataset loaded...")

	ret := m.Run()

	println("Cleaning up test data...")
	err = cleanupTrailDataset()
	if err != nil {
		log.Printf("Warning: Error cleaning up dataset : %v", err)
	}

	os.Exit(ret)
}

func TestGetTrails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.GET("/api/admin/trails", GetTrails)

	token, err := security.GenerateToken(testUsers[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/trails", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var trails Trails
	err = json.Unmarshal(w.Body.Bytes(), &trails)
	if err != nil {
		t.Fatal(err)
	}

	if len(trails) < 3 {
		t.Errorf("Expected at least 3 trails, got %d", len(trails))
	}
}

func TestGetTrailByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.GET("/api/admin/trails/:id", GetTrailByID)

	token, err := security.GenerateToken(testUsers[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		url := fmt.Sprintf("/api/admin/trails/%d", testTrails[0].ID)
		req := httptest.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d: %s",
				http.StatusOK, w.Code, w.Body.String())
		}

		var trail Trail
		err = json.Unmarshal(w.Body.Bytes(), &trail)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(testTrails[0].Name, trail.Name); diff != "" {
			t.Errorf("Trail name mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			"/api/admin/trails/999999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d",
				http.StatusNotFound, w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			"/api/admin/trails/abc", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d",
				http.StatusBadRequest, w.Code)
		}
	})
}

func TestPostTrail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.POST("/api/admin/trails", PostTrail)

	token, err := security.GenerateToken(testUsers[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		input := TrailCreateRequest{
			Name:      "Test New Trail",
			Country:   "Switzerland",
			Continent: "Europe",
		}
		body, _ := json.Marshal(input)

		req := httptest.NewRequest(http.MethodPost,
			"/api/admin/trails", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d: %s",
				http.StatusCreated, w.Code, w.Body.String())
		}

		var trail Trail
		err = json.Unmarshal(w.Body.Bytes(), &trail)
		if err != nil {
			t.Fatal(err)
		}

		if trail.ID == 0 {
			t.Error("Expected trail ID to be set")
		}

		// Clean up
		_, _ = database.DB().ExecContext(
			req.Context(), "DELETE FROM trail WHERE id = $1", trail.ID)
	})

	t.Run("duplicate name", func(t *testing.T) {
		input := TrailCreateRequest{
			Name:      testTrails[0].Name,
			Country:   "France",
			Continent: "Europe",
		}
		body, _ := json.Marshal(input)

		req := httptest.NewRequest(http.MethodPost,
			"/api/admin/trails", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected status %d, got %d: %s",
				http.StatusConflict, w.Code, w.Body.String())
		}
	})
}

func TestPutTrailByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.PUT("/api/admin/trails/:id", PutTrailByID)

	token, err := security.GenerateToken(testUsers[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		input := TrailUpdateRequest{
			Name:      testTrails[0].Name,
			Country:   "France",
			Continent: "Europe",
		}
		body, _ := json.Marshal(input)

		url := fmt.Sprintf("/api/admin/trails/%d", testTrails[0].ID)
		req := httptest.NewRequest(http.MethodPut, url,
			bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d: %s",
				http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		input := TrailUpdateRequest{
			Name:      "Nonexistent",
			Country:   "France",
			Continent: "Europe",
		}
		body, _ := json.Marshal(input)

		req := httptest.NewRequest(http.MethodPut,
			"/api/admin/trails/999999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d",
				http.StatusNotFound, w.Code)
		}
	})
}

func TestDeleteTrailByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.DELETE("/api/admin/trails/:id", DeleteTrailByID)

	token, err := security.GenerateToken(testUsers[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		ctx := t.Context()
		var trailID uint
		err := database.DB().QueryRowContext(ctx,
			`INSERT INTO trail (name, country, continent,
			created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id;`,
			"Trail To Delete", "France", "Europe").Scan(&trailID)
		if err != nil {
			t.Fatal(err)
		}

		url := fmt.Sprintf("/api/admin/trails/%d", trailID)
		req := httptest.NewRequest(http.MethodDelete, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d: %s",
				http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete,
			"/api/admin/trails/999999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d",
				http.StatusNotFound, w.Code)
		}
	})
}

func TestPostTrailsBulk(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.POST("/api/admin/trails/bulk", PostTrailsBulk)

	token, err := security.GenerateToken(testUsers[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		input := TrailBulkCreateRequest{
			Trails: []TrailCreateRequest{
				{Name: "Bulk Trail 1", Country: "Germany",
					Continent: "Europe"},
				{Name: "Bulk Trail 2", Country: "Austria",
					Continent: "Europe"},
			},
		}
		body, _ := json.Marshal(input)

		req := httptest.NewRequest(http.MethodPost,
			"/api/admin/trails/bulk", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d: %s",
				http.StatusCreated, w.Code, w.Body.String())
		}

		var created Trails
		err = json.Unmarshal(w.Body.Bytes(), &created)
		if err != nil {
			t.Fatal(err)
		}

		if len(created) != 2 {
			t.Errorf("Expected 2 trails, got %d", len(created))
		}

		// Clean up
		for _, trail := range created {
			_, _ = database.DB().ExecContext(
				req.Context(),
				"DELETE FROM trail WHERE id = $1", trail.ID)
		}
	})
}

func TestDeleteTrailsBulk(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.DELETE("/api/admin/trails/bulk", DeleteTrailsBulk)

	token, err := security.GenerateToken(testUsers[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		ctx := t.Context()
		var id1, id2 uint
		err := database.DB().QueryRowContext(ctx,
			`INSERT INTO trail (name, country, continent,
			created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id;`,
			"Bulk Delete 1", "France", "Europe").Scan(&id1)
		if err != nil {
			t.Fatal(err)
		}
		err = database.DB().QueryRowContext(ctx,
			`INSERT INTO trail (name, country, continent,
			created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id;`,
			"Bulk Delete 2", "France", "Europe").Scan(&id2)
		if err != nil {
			t.Fatal(err)
		}

		input := TrailBulkDeleteRequest{IDs: []uint{id1, id2}}
		body, _ := json.Marshal(input)

		req := httptest.NewRequest(http.MethodDelete,
			"/api/admin/trails/bulk", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d: %s",
				http.StatusOK, w.Code, w.Body.String())
		}
	})
}

func TestFindTrailByName(t *testing.T) {
	ctx := t.Context()

	t.Run("found", func(t *testing.T) {
		trail, err := FindTrailByName(ctx, testTrails[0].Name)
		if err != nil {
			t.Fatal(err)
		}
		if trail.Name != testTrails[0].Name {
			t.Errorf("Expected %q, got %q",
				testTrails[0].Name, trail.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := FindTrailByName(ctx, "Nonexistent Trail XYZ")
		if !errors.Is(err, ErrTrailNotFound) {
			t.Errorf("Expected ErrTrailNotFound, got %v", err)
		}
	})
}

func TestIsValidTrailName(t *testing.T) {
	ctx := t.Context()

	t.Run("valid", func(t *testing.T) {
		name := testTrails[0].Name
		valid, err := IsValidTrailName(ctx, &name)
		if err != nil {
			t.Fatal(err)
		}
		if !valid {
			t.Error("Expected trail to be valid")
		}
	})

	t.Run("nil", func(t *testing.T) {
		valid, err := IsValidTrailName(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !valid {
			t.Error("Expected nil trail to be valid")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		name := "Invalid Trail XYZ"
		valid, err := IsValidTrailName(ctx, &name)
		if err != nil {
			t.Fatal(err)
		}
		if valid {
			t.Error("Expected trail to be invalid")
		}
	})
}

func TestReturnTrailNames(t *testing.T) {
	ctx := t.Context()

	names, err := ReturnTrailNames(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 3 {
		t.Errorf("Expected at least 3 trail names, got %d", len(names))
	}
}

func TestReturnTrailsGrouped(t *testing.T) {
	ctx := t.Context()

	grouped, err := ReturnTrailsGrouped(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(grouped) == 0 {
		t.Error("Expected non-empty grouped result")
	}
	if _, ok := grouped["Europe"]; !ok {
		t.Error("Expected Europe continent in grouped result")
	}
}

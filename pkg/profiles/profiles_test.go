package profiles

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/gruntwork-io/terratest/modules/random"
)

// test data IDs populated during setup
var (
	publicUserID    uint
	privateUserID   uint
	publicUsername  string
	privateUsername string
	sharedPackID    uint
	privatePackID   uint
)

func TestMain(m *testing.M) {
	err := config.EnvInit("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	err = database.Initialization()
	if err != nil {
		log.Fatalf("Error connecting database: %v", err)
	}

	err = database.Migrate()
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	println("Loading profiles test dataset...")
	err = loadingProfileDataset()
	if err != nil {
		log.Fatalf("Error loading dataset: %v", err)
	}

	ret := m.Run()

	println("Cleaning up profiles test data...")
	err = cleanupProfileDataset()
	if err != nil {
		log.Printf("Warning: Error cleaning up dataset: %v", err)
	}

	os.Exit(ret)
}

func loadingProfileDataset() error {
	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	publicUsername = "pub-" + random.UniqueId()
	privateUsername = "priv-" + random.UniqueId()

	// Create public user (is_profile_public = true)
	err := database.DB().QueryRowContext(ctx,
		`INSERT INTO account (username, email, firstname, lastname, role, status,
			is_profile_public, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id;`,
		publicUsername, publicUsername+"@test.com", "Public", "User",
		"standard", "active", true, now, now,
	).Scan(&publicUserID)
	if err != nil {
		return fmt.Errorf("failed to insert public user: %w", err)
	}

	hashedPassword, err := security.HashPassword("password")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	var pwID int
	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO password (user_id, password, updated_at) VALUES ($1,$2,$3) RETURNING id;`,
		publicUserID, hashedPassword, now,
	).Scan(&pwID)
	if err != nil {
		return fmt.Errorf("failed to insert password for public user: %w", err)
	}

	// Create private user (is_profile_public = false, default)
	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO account (username, email, firstname, lastname, role, status,
			created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id;`,
		privateUsername, privateUsername+"@test.com", "Private", "User",
		"standard", "active", now, now,
	).Scan(&privateUserID)
	if err != nil {
		return fmt.Errorf("failed to insert private user: %w", err)
	}

	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO password (user_id, password, updated_at) VALUES ($1,$2,$3) RETURNING id;`,
		privateUserID, hashedPassword, now,
	).Scan(&pwID)
	if err != nil {
		return fmt.Errorf("failed to insert password for private user: %w", err)
	}

	// Create inventory item for public user
	var itemID uint
	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO inventory (user_id, item_name, category, description,
			weight, url, price, currency, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id;`,
		publicUserID, "Backpack", "Gear", "Test backpack", 1000, "https://example.com", 100, "USD", now, now,
	).Scan(&itemID)
	if err != nil {
		return fmt.Errorf("failed to insert inventory item: %w", err)
	}

	// Create shared pack for public user
	sharingCode := "profile-test-" + random.UniqueId()
	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO pack (user_id, pack_name, pack_description, sharing_code, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id;`,
		publicUserID, "Shared Pack", "A shared pack", sharingCode, now, now,
	).Scan(&sharedPackID)
	if err != nil {
		return fmt.Errorf("failed to insert shared pack: %w", err)
	}

	// Add item to shared pack
	var pcID uint
	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO pack_content (pack_id, item_id, quantity, worn, consumable, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id;`,
		sharedPackID, itemID, 1, false, false, now, now,
	).Scan(&pcID)
	if err != nil {
		return fmt.Errorf("failed to insert pack content: %w", err)
	}

	// Create private pack (no sharing_code) for public user
	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO pack (user_id, pack_name, pack_description, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5) RETURNING id;`,
		publicUserID, "Private Pack", "A private pack", now, now,
	).Scan(&privatePackID)
	if err != nil {
		return fmt.Errorf("failed to insert private pack: %w", err)
	}

	return nil
}

func cleanupProfileDataset() error {
	ctx := context.Background()

	for _, id := range []uint{publicUserID, privateUserID} {
		if id != 0 {
			_, err := database.DB().ExecContext(ctx, "DELETE FROM account WHERE id = $1", id)
			if err != nil {
				return fmt.Errorf("failed to delete user %d: %w", id, err)
			}
		}
	}
	return nil
}

func TestGetPublicProfile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/user/:username", GetPublicProfile)

	req, err := http.NewRequest(http.MethodGet, "/user/"+publicUsername, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var profile PublicProfile
	if err := json.Unmarshal(w.Body.Bytes(), &profile); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if profile.Username != publicUsername {
		t.Errorf("Expected username %s, got %s", publicUsername, profile.Username)
	}

	if profile.Firstname != "Public" {
		t.Errorf("Expected firstname 'Public', got %s", profile.Firstname)
	}

	// Should have exactly 1 shared pack (private pack excluded)
	if len(profile.SharedPacks) != 1 {
		t.Fatalf("Expected 1 shared pack, got %d", len(profile.SharedPacks))
	}

	if profile.SharedPacks[0].PackName != "Shared Pack" {
		t.Errorf("Expected pack name 'Shared Pack', got %s", profile.SharedPacks[0].PackName)
	}

	if profile.SharedPacks[0].PackItemsCount != 1 {
		t.Errorf("Expected 1 item in pack, got %d", profile.SharedPacks[0].PackItemsCount)
	}

	if profile.SharedPacks[0].PackWeight != 1000 {
		t.Errorf("Expected pack weight 1000, got %d", profile.SharedPacks[0].PackWeight)
	}
}

func TestGetPublicProfile_NonExistentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/user/:username", GetPublicProfile)

	req, err := http.NewRequest(http.MethodGet, "/user/nonexistent-user-xyz", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetPublicProfile_PrivateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/user/:username", GetPublicProfile)

	req, err := http.NewRequest(http.MethodGet, "/user/"+privateUsername, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 404 (anti-enumeration: same as non-existent user)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d for private user, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetPublicProfile_NoSharedPacks(t *testing.T) {
	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// Create a public user with no shared packs
	username := "nopacks-" + random.UniqueId()
	var userID uint
	err := database.DB().QueryRowContext(ctx,
		`INSERT INTO account (username, email, firstname, lastname, role, status,
			is_profile_public, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id;`,
		username, username+"@test.com", "NoPack", "User",
		"standard", "active", true, now, now,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	hashedPassword, err := security.HashPassword("password")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	var pwID int
	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO password (user_id, password, updated_at) VALUES ($1,$2,$3) RETURNING id;`,
		userID, hashedPassword, now,
	).Scan(&pwID)
	if err != nil {
		t.Fatalf("Failed to insert password: %v", err)
	}

	defer func() {
		_, _ = database.DB().ExecContext(ctx, "DELETE FROM account WHERE id = $1", userID)
	}()

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/user/:username", GetPublicProfile)

	req, err := http.NewRequest(http.MethodGet, "/user/"+username, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var profile PublicProfile
	if err := json.Unmarshal(w.Body.Bytes(), &profile); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should return empty array, not null
	if profile.SharedPacks == nil {
		t.Error("Expected empty array for shared_packs, got nil")
	}
	if len(profile.SharedPacks) != 0 {
		t.Errorf("Expected 0 shared packs, got %d", len(profile.SharedPacks))
	}
}

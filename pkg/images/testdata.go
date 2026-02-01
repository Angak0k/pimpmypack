package images

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/accounts"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/packs"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gruntwork-io/terratest/modules/random"
)

// Test user for image storage tests
var testUser = accounts.User{
	Username:     "image-test-user-" + random.UniqueId(),
	Email:        "image-test-" + random.UniqueId() + "@example.com",
	Firstname:    "Image",
	Lastname:     "Tester",
	Role:         "standard",
	Status:       "active",
	Password:     "password",
	LastPassword: "password",
}

// Test packs for image storage tests
var testPacks = []packs.Pack{
	{
		ID:              999,
		UserID:          0, // Will be set dynamically
		PackName:        "Image Test Pack 1",
		PackDescription: "Pack for testing image save operations",
	},
	{
		ID:              1000,
		UserID:          0, // Will be set dynamically
		PackName:        "Image Test Pack 2",
		PackDescription: "Pack for testing image update operations",
	},
	{
		ID:              1001,
		UserID:          0, // Will be set dynamically
		PackName:        "Image Test Pack 3",
		PackDescription: "Pack for testing image get operations",
	},
	{
		ID:              1002,
		UserID:          0, // Will be set dynamically
		PackName:        "Image Test Pack 4",
		PackDescription: "Pack for testing image delete operations",
	},
	{
		ID:              1003,
		UserID:          0, // Will be set dynamically
		PackName:        "Image Test Pack 5",
		PackDescription: "Pack for testing image exists operations",
	},
}

// createOrFindTestUser creates or finds the test user in the database
func createOrFindTestUser(ctx context.Context, tx *sql.Tx) error {
	var existingID int
	err := tx.QueryRowContext(ctx,
		"SELECT id FROM account WHERE username = $1", testUser.Username).Scan(&existingID)

	switch {
	case err == nil:
		// User exists, use existing ID
		if existingID < 0 {
			return fmt.Errorf("invalid user ID: negative value %d for user %s", existingID, testUser.Username)
		}
		testUser.ID = uint(existingID)
		return nil

	case errors.Is(err, sql.ErrNoRows):
		// User doesn't exist, create them
		return createTestUser(ctx, tx)

	default:
		return fmt.Errorf("failed to check existing user: %w", err)
	}
}

// createTestUser creates a new test user with password
func createTestUser(ctx context.Context, tx *sql.Tx) error {
	println("-> Creating test user for image tests...")

	err := tx.QueryRowContext(ctx,
		`INSERT INTO account (username, email, firstname, lastname, role, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id;`,
		testUser.Username,
		testUser.Email,
		testUser.Firstname,
		testUser.Lastname,
		testUser.Role,
		testUser.Status,
		time.Now().Truncate(time.Second),
		time.Now().Truncate(time.Second)).Scan(&testUser.ID)
	if err != nil {
		return fmt.Errorf("failed to insert test user: %w", err)
	}

	// Create password for test user
	hashedPassword, err := security.HashPassword(testUser.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	hashedLastPassword, err := security.HashPassword(testUser.LastPassword)
	if err != nil {
		return fmt.Errorf("failed to hash last password: %w", err)
	}

	var passwordID uint
	err = tx.QueryRowContext(ctx,
		`INSERT INTO password (user_id, password, last_password, updated_at) VALUES ($1,$2,$3,$4)
		RETURNING id;`,
		testUser.ID,
		hashedPassword,
		hashedLastPassword,
		time.Now().Truncate(time.Second)).Scan(&passwordID)
	if err != nil {
		return fmt.Errorf("failed to insert password for test user: %w", err)
	}

	return nil
}

// loadImageTestData loads test packs for image storage tests
func loadImageTestData() error {
	ctx := context.Background()

	// Start a transaction
	tx, err := database.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	println("-> Loading image test data...")

	// Create or find test user
	err = createOrFindTestUser(ctx, tx)
	if err != nil {
		return err
	}

	println("-> Loading image test packs...")

	// Set user ID for all test packs
	for i := range testPacks {
		testPacks[i].UserID = testUser.ID
	}

	// Insert test packs
	for i := range testPacks {
		err = insertTestPack(ctx, tx, &testPacks[i])
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	println("-> Image test data loaded...")
	return nil
}

// insertTestPack inserts a test pack if it doesn't already exist
func insertTestPack(ctx context.Context, tx *sql.Tx, pack *packs.Pack) error {
	// Check if pack already exists
	var existingPackID int
	err := tx.QueryRowContext(ctx,
		"SELECT id FROM pack WHERE id = $1", pack.ID).Scan(&existingPackID)
	if err == nil {
		// Pack exists, skip
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check existing pack %d: %w", pack.ID, err)
	}

	// Pack doesn't exist, create it
	_, err = tx.ExecContext(ctx,
		`INSERT INTO pack (id, user_id, pack_name, pack_description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		pack.ID,
		pack.UserID,
		pack.PackName,
		pack.PackDescription,
		time.Now().Truncate(time.Second),
		time.Now().Truncate(time.Second))
	if err != nil {
		return fmt.Errorf("failed to insert pack %d: %w", pack.ID, err)
	}

	return nil
}

// cleanupImageTestData removes test packs, user, and their images
func cleanupImageTestData() error {
	ctx := context.Background()

	println("-> Cleaning up image test data...")

	// Delete test packs (images will cascade delete due to ON DELETE CASCADE)
	for _, pack := range testPacks {
		_, err := database.DB().ExecContext(ctx, "DELETE FROM pack WHERE id = $1", pack.ID)
		if err != nil {
			return fmt.Errorf("failed to delete pack %d: %w", pack.ID, err)
		}
	}

	// Delete test user (this will cascade delete passwords)
	if testUser.ID != 0 {
		_, err := database.DB().ExecContext(ctx, "DELETE FROM account WHERE id = $1", testUser.ID)
		if err != nil {
			return fmt.Errorf("failed to delete test user %d: %w", testUser.ID, err)
		}
	}

	println("-> Image test data cleaned up...")
	return nil
}

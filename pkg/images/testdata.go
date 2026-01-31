package images

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/dataset"
)

// Test packs for image storage tests
var testPacks = []dataset.Pack{
	{
		ID:              999,
		UserID:          1,
		PackName:        "Image Test Pack 1",
		PackDescription: "Pack for testing image save operations",
	},
	{
		ID:              1000,
		UserID:          1,
		PackName:        "Image Test Pack 2",
		PackDescription: "Pack for testing image update operations",
	},
	{
		ID:              1001,
		UserID:          1,
		PackName:        "Image Test Pack 3",
		PackDescription: "Pack for testing image get operations",
	},
	{
		ID:              1002,
		UserID:          1,
		PackName:        "Image Test Pack 4",
		PackDescription: "Pack for testing image delete operations",
	},
	{
		ID:              1003,
		UserID:          1,
		PackName:        "Image Test Pack 5",
		PackDescription: "Pack for testing image exists operations",
	},
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

	println("-> Loading image test packs...")

	// Insert test packs
	for i := range testPacks {
		// Check if pack already exists
		var existingID int
		err := tx.QueryRowContext(ctx,
			"SELECT id FROM pack WHERE id = $1", testPacks[i].ID).Scan(&existingID)
		if err == nil {
			// Pack exists, skip
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check existing pack %d: %w", testPacks[i].ID, err)
		}

		// Pack doesn't exist, create it
		_, err = tx.ExecContext(ctx,
			`INSERT INTO pack (id, user_id, pack_name, pack_description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			testPacks[i].ID,
			testPacks[i].UserID,
			testPacks[i].PackName,
			testPacks[i].PackDescription,
			time.Now().Truncate(time.Second),
			time.Now().Truncate(time.Second))
		if err != nil {
			return fmt.Errorf("failed to insert pack %d: %w", testPacks[i].ID, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	println("-> Image test packs loaded...")
	return nil
}

// cleanupImageTestData removes test packs and their images
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

	println("-> Image test data cleaned up...")
	return nil
}

package inventories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/accounts"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gruntwork-io/terratest/modules/random"
)

var users = []accounts.User{
	{
		Username:     "user-" + random.UniqueId(),
		Email:        "user-" + random.UniqueId() + "@exemple.com",
		Firstname:    "Joseph",
		Lastname:     "Doe",
		Role:         "standard",
		Status:       "active",
		Password:     "password",
		LastPassword: "password",
	},
	{
		Username:     "user-" + random.UniqueId(),
		Email:        "user-" + random.UniqueId() + "@exemple.com",
		Firstname:    "Syvie",
		Lastname:     "Doe",
		Role:         "standard",
		Status:       "active",
		Password:     "password",
		LastPassword: "password",
	},
}

var testPackIDs []int

var inventories = Inventories{
	{
		UserID:   1,
		ItemName: "Backpack",
		Category: "Outdoor Gear",
		Weight:   950,
		URL:      "https://example.com/backpack",
		Price:    50,
		Currency: "USD",
	},
	{
		UserID:   1,
		ItemName: "Tent",
		Category: "Shelter",
		Weight:   1200,
		URL:      "https://example.com/tent",
		Price:    150,
		Currency: "USD",
	},
	{
		UserID:   1,
		ItemName: "Sleeping Bag",
		Category: "Sleeping",
		Weight:   800,
		URL:      "https://example.com/sleeping-bag",
		Price:    120,
		Currency: "EUR",
	},
	{
		UserID:   2,
		ItemName: "Sleeping Bag",
		Category: "Sleeping",
		Weight:   800,
		URL:      "https://example.com/sleeping-bag",
		Price:    120,
		Currency: "EUR",
	},
}

// mergeTestItems holds extra inventory items created specifically for merge tests.
// These are populated by createMergeTestData and cleaned up by cleanupMergeTestData.
var mergeTestItems []Inventory
var mergeTestPackIDs []int

func loadingInventoryDataset() error {
	// Start a transaction
	tx, err := database.DB().BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Load accounts dataset
	for i := range users {
		var id uint
		err := tx.QueryRowContext(context.Background(),
			`INSERT INTO account (username, email, firstname, lastname, role, status, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id;`,
			users[i].Username,
			users[i].Email,
			users[i].Firstname,
			users[i].Lastname,
			users[i].Role,
			users[i].Status,
			time.Now().Truncate(time.Second),
			time.Now().Truncate(time.Second)).Scan(&users[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}

		hashedPassword, err := security.HashPassword(users[i].Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		hashedLastPassword, err := security.HashPassword(users[i].LastPassword)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		err = tx.QueryRowContext(context.Background(),
			`INSERT INTO password (user_id, password, last_password, updated_at)
			VALUES ($1,$2,$3,$4)
			RETURNING id;`,
			users[i].ID,
			hashedPassword,
			hashedLastPassword,
			time.Now().Truncate(time.Second)).Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to insert password: %w", err)
		}
	}

	// Transform inventories dataset by using the real user_id
	for i := range inventories {
		switch inventories[i].UserID {
		case 1:
			inventories[i].UserID = accounts.FindUserIDByUsername(users, users[0].Username)
		case 2:
			inventories[i].UserID = accounts.FindUserIDByUsername(users, users[1].Username)
		}
	}

	// Insert inventories dataset
	for i := range inventories {
		err := tx.QueryRowContext(context.Background(),
			`INSERT INTO inventory
			(user_id, item_name, category, description, weight, url, price, currency,
				created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			RETURNING id;`,
			inventories[i].UserID,
			inventories[i].ItemName,
			inventories[i].Category,
			inventories[i].Description,
			inventories[i].Weight,
			inventories[i].URL,
			inventories[i].Price,
			inventories[i].Currency,
			time.Now().Truncate(time.Second),
			time.Now().Truncate(time.Second)).Scan(&inventories[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert inventory: %w", err)
		}
	}

	// Insert test packs and pack_content for pack_count testing
	if err := insertTestPackData(tx); err != nil {
		return err
	}

	// Set expected PackCount values to match inserted pack_content data
	inventories[0].PackCount = 2 // Backpack → 2 packs
	inventories[1].PackCount = 1 // Tent → 1 pack
	// inventories[2] and inventories[3] remain 0 (not linked to any pack)

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// insertTestPackData creates test packs and links them to inventory items for pack_count testing.
// Backpack → 2 packs, Tent → 1 pack, Sleeping Bag → 0 packs.
func insertTestPackData(tx *sql.Tx) error {
	now := time.Now().Truncate(time.Second)
	ctx := context.Background()

	var testPackID, testPackID2 int

	err := tx.QueryRowContext(ctx,
		`INSERT INTO pack (user_id, pack_name, pack_description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;`,
		inventories[0].UserID, "Test Pack", "Pack for testing pack_count", now, now).Scan(&testPackID)
	if err != nil {
		return fmt.Errorf("failed to insert test pack: %w", err)
	}

	err = tx.QueryRowContext(ctx,
		`INSERT INTO pack (user_id, pack_name, pack_description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;`,
		inventories[0].UserID, "Test Pack 2", "Second pack for testing pack_count", now, now).Scan(&testPackID2)
	if err != nil {
		return fmt.Errorf("failed to insert test pack 2: %w", err)
	}

	// Link Backpack (inventories[0]) to both packs → pack_count=2
	for _, packID := range []int{testPackID, testPackID2} {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO pack_content (pack_id, item_id, quantity, worn, consumable, created_at, updated_at)
			VALUES ($1, $2, 1, false, false, $3, $4);`,
			packID, inventories[0].ID, now, now)
		if err != nil {
			return fmt.Errorf("failed to insert pack_content: %w", err)
		}
	}

	// Link Tent (inventories[1]) to first pack only → pack_count=1
	_, err = tx.ExecContext(ctx,
		`INSERT INTO pack_content (pack_id, item_id, quantity, worn, consumable, created_at, updated_at)
		VALUES ($1, $2, 1, false, false, $3, $4);`,
		testPackID, inventories[1].ID, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert pack_content for tent: %w", err)
	}

	testPackIDs = []int{testPackID, testPackID2}

	return nil
}

// cleanupInventoryDataset removes all test data created by loadingInventoryDataset
func cleanupInventoryDataset() error {
	ctx := context.Background()

	println("-> Cleaning up inventory test data...")

	// Delete test packs first (pack_content will cascade)
	for _, packID := range testPackIDs {
		_, err := database.DB().ExecContext(ctx, "DELETE FROM pack WHERE id = $1", packID)
		if err != nil {
			return fmt.Errorf("failed to delete test pack %d: %w", packID, err)
		}
	}

	// Delete inventories
	for _, inv := range inventories {
		if inv.ID != 0 {
			_, err := database.DB().ExecContext(ctx, "DELETE FROM inventory WHERE id = $1", inv.ID)
			if err != nil {
				return fmt.Errorf("failed to delete inventory %d: %w", inv.ID, err)
			}
		}
	}

	// Delete users (passwords will cascade delete)
	for _, user := range users {
		if user.ID != 0 {
			_, err := database.DB().ExecContext(ctx, "DELETE FROM account WHERE id = $1", user.ID)
			if err != nil {
				return fmt.Errorf("failed to delete user %d: %w", user.ID, err)
			}
		}
	}

	println("-> Inventory test data cleaned up...")
	return nil
}

// createMergeTestData creates inventory items and pack_content rows for merge testing.
// It creates a source and target item, a shared pack, and a non-shared pack.
// Returns (sourceItem, targetItem, sharedPackID, nonSharedPackID).
func createMergeTestData(ctx context.Context) error {
	now := time.Now().Truncate(time.Second)

	// Create source item
	source := Inventory{
		UserID:      users[0].ID,
		ItemName:    "Merge Source",
		Category:    "Test",
		Description: "Source item for merge test",
		Weight:      100,
		URL:         "https://example.com/source",
		Price:       10,
		Currency:    "USD",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := database.DB().QueryRowContext(ctx,
		`INSERT INTO inventory (user_id, item_name, category, description,
		weight, url, price, currency, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id;`,
		source.UserID, source.ItemName, source.Category, source.Description,
		source.Weight, source.URL, source.Price, source.Currency, now, now).Scan(&source.ID)
	if err != nil {
		return fmt.Errorf("failed to insert merge source item: %w", err)
	}

	// Create target item
	target := Inventory{
		UserID:      users[0].ID,
		ItemName:    "Merge Target",
		Category:    "Test",
		Description: "Target item for merge test",
		Weight:      200,
		URL:         "https://example.com/target",
		Price:       20,
		Currency:    "USD",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO inventory (user_id, item_name, category, description,
		weight, url, price, currency, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id;`,
		target.UserID, target.ItemName, target.Category, target.Description,
		target.Weight, target.URL, target.Price, target.Currency, now, now).Scan(&target.ID)
	if err != nil {
		return fmt.Errorf("failed to insert merge target item: %w", err)
	}

	mergeTestItems = []Inventory{source, target}

	// Create a shared pack (both items in it)
	var sharedPackID, nonSharedPackID int
	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO pack (user_id, pack_name, pack_description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id;`,
		users[0].ID, "Merge Shared Pack", "Shared pack for merge test", now, now).Scan(&sharedPackID)
	if err != nil {
		return fmt.Errorf("failed to insert merge shared pack: %w", err)
	}

	// Add both source and target to the shared pack with different quantities
	_, err = database.DB().ExecContext(ctx,
		`INSERT INTO pack_content (pack_id, item_id, quantity, worn, consumable, created_at, updated_at)
		VALUES ($1, $2, 2, false, false, $3, $4);`,
		sharedPackID, source.ID, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert source into shared pack: %w", err)
	}
	_, err = database.DB().ExecContext(ctx,
		`INSERT INTO pack_content (pack_id, item_id, quantity, worn, consumable, created_at, updated_at)
		VALUES ($1, $2, 3, false, false, $3, $4);`,
		sharedPackID, target.ID, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert target into shared pack: %w", err)
	}

	// Create a non-shared pack (only source in it)
	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO pack (user_id, pack_name, pack_description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id;`,
		users[0].ID, "Merge Non-Shared Pack", "Non-shared pack for merge test", now, now).Scan(&nonSharedPackID)
	if err != nil {
		return fmt.Errorf("failed to insert merge non-shared pack: %w", err)
	}
	_, err = database.DB().ExecContext(ctx,
		`INSERT INTO pack_content (pack_id, item_id, quantity, worn, consumable, created_at, updated_at)
		VALUES ($1, $2, 1, false, false, $3, $4);`,
		nonSharedPackID, source.ID, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert source into non-shared pack: %w", err)
	}

	mergeTestPackIDs = []int{sharedPackID, nonSharedPackID}

	return nil
}

// cleanupMergeTestData removes all merge test data
func cleanupMergeTestData(ctx context.Context) {
	for _, packID := range mergeTestPackIDs {
		_, _ = database.DB().ExecContext(ctx, "DELETE FROM pack WHERE id = $1", packID)
	}
	for _, item := range mergeTestItems {
		if item.ID != 0 {
			_, _ = database.DB().ExecContext(ctx, "DELETE FROM inventory_images WHERE item_id = $1", item.ID)
			_, _ = database.DB().ExecContext(ctx, "DELETE FROM inventory WHERE id = $1", item.ID)
		}
	}
	mergeTestItems = nil
	mergeTestPackIDs = nil
}

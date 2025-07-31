package packs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/dataset"
	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gruntwork-io/terratest/modules/random"
)

var users = []dataset.User{
	{
		Username:     "user-" + random.UniqueId(),
		Email:        "user-" + random.UniqueId() + "@exemple.com",
		Firstname:    "John",
		Lastname:     "Doe",
		Role:         "standard",
		Status:       "active",
		Password:     "password",
		LastPassword: "password",
	},
	{
		Username:     "user-" + random.UniqueId(),
		Email:        "user-" + random.UniqueId() + "@exemple.com",
		Firstname:    "Jane",
		Lastname:     "Doe",
		Role:         "standard",
		Status:       "active",
		Password:     "password",
		LastPassword: "password",
	},
}

var inventoriesUserPack1 = dataset.Inventories{
	{
		UserID:      1,
		ItemName:    "Backpack",
		Category:    "Outdoor Gear",
		Description: "Spacious backpack for hiking",
		Weight:      950,
		URL:         "https://example.com/backpack",
		Price:       50,
		Currency:    "USD",
	},
	{
		UserID:      1,
		ItemName:    "Tent",
		Category:    "Shelter",
		Description: "Spacious tent for hiking",
		Weight:      1200,
		URL:         "https://example.com/tent",
		Price:       150,
		Currency:    "USD",
	},
	{
		UserID:      1,
		ItemName:    "Sleeping Bag",
		Category:    "Sleeping",
		Description: "Spacious sleeping bag for hiking",
		Weight:      800,
		URL:         "https://example.com/sleeping-bag",
		Price:       120,
		Currency:    "EUR",
	},
}

var packs = dataset.Packs{
	{
		UserID:          1,
		PackName:        "First Pack",
		PackDescription: "Description for the first pack",
		SharingCode:     "123456",
	},
	{
		UserID:          1,
		PackName:        "Second Pack",
		PackDescription: "Description for the second pack",
		SharingCode:     "654321",
	},
	{
		UserID:          2,
		PackName:        "Third Pack",
		PackDescription: "Description for the third pack",
		SharingCode:     "789456",
	},
	{
		UserID:          1,
		PackName:        "Special Pack",
		PackDescription: "Description for the special pack",
		SharingCode:     "321654",
	},
}

var packItems = dataset.PackContents{
	{
		PackID:     1,
		ItemID:     1,
		Quantity:   2,
		Worn:       true,
		Consumable: false,
	},
	{
		PackID:     1,
		ItemID:     2,
		Quantity:   3,
		Worn:       false,
		Consumable: true,
	},
	{
		PackID:     2,
		ItemID:     2,
		Quantity:   1,
		Worn:       true,
		Consumable: false,
	},
	{
		PackID:     2,
		ItemID:     1,
		Quantity:   4,
		Worn:       true,
		Consumable: true,
	},
	{
		PackID:     3,
		ItemID:     1,
		Quantity:   2,
		Worn:       false,
		Consumable: false,
	},
	{
		PackID:     4,
		ItemID:     1,
		Quantity:   2,
		Worn:       false,
		Consumable: false,
	},
}

var packWithItems = dataset.PackContentWithItems{
	{
		PackContentID:   1,
		PackID:          4,
		ItemName:        "Backpack",
		Category:        "Outdoor Gear",
		ItemDescription: "Spacious backpack for hiking",
		Weight:          950,
		ItemURL:         "https://example.com/backpack",
		Price:           50,
		Currency:        "USD",
		Quantity:        2,
		Worn:            false,
		Consumable:      false,
	},
}

func loadingPackDataset() error {
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

	// Always load our test accounts
	if err := loadAccounts(tx); err != nil {
		return fmt.Errorf("failed to load accounts: %w", err)
	}

	// Now load the rest of the data
	transformInventories()
	if err := loadInventories(tx); err != nil {
		return fmt.Errorf("failed to load inventories: %w", err)
	}
	transformPackContents()
	if err := loadPackContents(tx); err != nil {
		return fmt.Errorf("failed to load pack contents: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func loadAccounts(tx *sql.Tx) error {
	println("-> Loading accounts and passwords ...")
	for i := range users {
		// First, check if the user already exists
		var existingID int
		err := tx.QueryRowContext(context.Background(), "SELECT id FROM account WHERE username = $1", users[i].Username).Scan(&existingID)
		if err == nil {
			// User exists, update their ID
			if existingID < 0 {
				return fmt.Errorf("invalid user ID: negative value %d for user %s", existingID, users[i].Username)
			}
			users[i].ID = uint(existingID)
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check existing user %s: %w", users[i].Username, err)
		}

		// User doesn't exist, create them
		//nolint:execinquery
		err = tx.QueryRowContext(context.Background(),
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
			return fmt.Errorf("failed to insert user %s: %w", users[i].Username, err)
		}

		hashedPassword, err := security.HashPassword(users[i].Password)
		if err != nil {
			return fmt.Errorf("failed to hash password for user %s: %w", users[i].Username, err)
		}

		hashedLastPassword, err := security.HashPassword(users[i].LastPassword)
		if err != nil {
			return fmt.Errorf("failed to hash last password for user %s: %w", users[i].Username, err)
		}

		//nolint:execinquery
		err = tx.QueryRowContext(context.Background(),
			`INSERT INTO password (user_id, password, last_password, updated_at) VALUES ($1,$2,$3,$4) 
			RETURNING id;`,
			users[i].ID,
			hashedPassword,
			hashedLastPassword,
			time.Now().Truncate(time.Second)).Scan(&users[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert password for user %s (ID: %d): %w", users[i].Username, users[i].ID, err)
		}
	}
	println("-> Accounts Loaded...")
	return nil
}

func transformInventories() {
	// Transform inventories dataset by using the real user_id
	for i := range inventoriesUserPack1 {
		switch inventoriesUserPack1[i].UserID {
		case 1:
			inventoriesUserPack1[i].UserID = helper.FindUserIDByUsername(users, users[0].Username)
		case 2:
			inventoriesUserPack1[i].UserID = helper.FindUserIDByUsername(users, users[1].Username)
		}
	}

	// Transform packs dataset
	for i := range packs {
		switch packs[i].UserID {
		case 1:
			packs[i].UserID = helper.FindUserIDByUsername(users, users[0].Username)
		case 2:
			packs[i].UserID = helper.FindUserIDByUsername(users, users[1].Username)
		}
	}
}

func loadInventories(tx *sql.Tx) error {
	println("-> Loading Inventories...")
	for i := range inventoriesUserPack1 {
		// Check if user exists before inserting inventory
		var userExists bool
		err := tx.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM account WHERE id = $1)",
			inventoriesUserPack1[i].UserID).Scan(&userExists)
		if err != nil {
			return fmt.Errorf("failed to check if user %d exists: %w", inventoriesUserPack1[i].UserID, err)
		}
		if !userExists {
			return fmt.Errorf("foreign key violation: user_id %d does not exist in account table",
				inventoriesUserPack1[i].UserID)
		}

		//nolint:execinquery
		err = tx.QueryRowContext(context.Background(),
			`INSERT INTO inventory 
			(user_id, item_name, category, description, weight, url, price, currency, 
				created_at, updated_at) 
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) 
			RETURNING id;`,
			inventoriesUserPack1[i].UserID,
			inventoriesUserPack1[i].ItemName,
			inventoriesUserPack1[i].Category,
			inventoriesUserPack1[i].Description,
			inventoriesUserPack1[i].Weight,
			inventoriesUserPack1[i].URL,
			inventoriesUserPack1[i].Price,
			inventoriesUserPack1[i].Currency,
			time.Now().Truncate(time.Second),
			time.Now().Truncate(time.Second)).Scan(&inventoriesUserPack1[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert inventory item %s for user %d: %w",
				inventoriesUserPack1[i].ItemName,
				inventoriesUserPack1[i].UserID,
				err)
		}
	}
	println("-> Inventories Loaded...")
	println("-> Loading Packs...")

	// Insert packs dataset
	for i := range packs {
		// Check if user exists before inserting pack
		var userExists bool
		err := tx.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM account WHERE id = $1)", packs[i].UserID).Scan(&userExists)
		if err != nil {
			return fmt.Errorf("failed to check if user %d exists: %w", packs[i].UserID, err)
		}
		if !userExists {
			return fmt.Errorf("foreign key violation: user_id %d does not exist in account table", packs[i].UserID)
		}

		//nolint:execinquery
		err = tx.QueryRowContext(context.Background(),
			`INSERT INTO pack (user_id, pack_name, pack_description, sharing_code, created_at, updated_at) 
			VALUES ($1,$2,$3,$4,$5,$6) 
			RETURNING id;`,
			packs[i].UserID,
			packs[i].PackName,
			packs[i].PackDescription,
			packs[i].SharingCode,
			time.Now().Truncate(time.Second),
			time.Now().Truncate(time.Second)).Scan(&packs[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert pack %s for user %d: %w",
				packs[i].PackName,
				packs[i].UserID,
				err)
		}
	}
	println("-> Packs Loaded...")
	return nil
}

func transformPackContents() {
	// Transform packs_contents dataset

	for i := range packItems {
		switch packItems[i].PackID {
		case 1:
			packItems[i].PackID = helper.FindPackIDByPackName(packs, "First Pack")
		case 2:
			packItems[i].PackID = helper.FindPackIDByPackName(packs, "Second Pack")
		case 3:
			packItems[i].PackID = helper.FindPackIDByPackName(packs, "Third Pack")
		case 4:
			packItems[i].PackID = helper.FindPackIDByPackName(packs, "Special Pack")
		}
		switch packItems[i].ItemID {
		case 1:
			packItems[i].ItemID = helper.FindItemIDByItemName(inventoriesUserPack1, "Backpack")
		case 2:
			packItems[i].ItemID = helper.FindItemIDByItemName(inventoriesUserPack1, "Tent")
		case 3:
			packItems[i].ItemID = helper.FindItemIDByItemName(inventoriesUserPack1, "Sleeping Bag")
		}
	}

	for i := range packWithItems {
		switch packWithItems[i].PackID {
		case 1:
			packWithItems[i].PackID = helper.FindPackIDByPackName(packs, "First Pack")
		case 2:
			packWithItems[i].PackID = helper.FindPackIDByPackName(packs, "Second Pack")
		case 3:
			packWithItems[i].PackID = helper.FindPackIDByPackName(packs, "Third Pack")
		case 4:
			packWithItems[i].PackID = helper.FindPackIDByPackName(packs, "Special Pack")
		}
	}
}

func loadPackContents(tx *sql.Tx) error {
	println("-> Loading Pack Contents...")
	// Then load pack contents
	for i := range packItems {
		// Check if pack exists
		var packExists bool
		err := tx.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM pack WHERE id = $1)", packItems[i].PackID).Scan(&packExists)
		if err != nil {
			return fmt.Errorf("failed to check if pack %d exists: %w", packItems[i].PackID, err)
		}
		if !packExists {
			return fmt.Errorf("foreign key violation: pack_id %d does not exist in pack table", packItems[i].PackID)
		}

		// Check if item exists
		var itemExists bool
		err = tx.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM inventory WHERE id = $1)", packItems[i].ItemID).Scan(&itemExists)
		if err != nil {
			return fmt.Errorf("failed to check if item %d exists: %w", packItems[i].ItemID, err)
		}
		if !itemExists {
			return fmt.Errorf("foreign key violation: item_id %d does not exist in inventory table", packItems[i].ItemID)
		}

		//nolint:execinquery
		err = tx.QueryRowContext(context.Background(),
			`INSERT INTO pack_content (pack_id, item_id, quantity, worn, consumable, created_at, updated_at) 
			VALUES ($1,$2,$3,$4,$5,$6,$7) 
			RETURNING id;`,
			packItems[i].PackID,
			packItems[i].ItemID,
			packItems[i].Quantity,
			packItems[i].Worn,
			packItems[i].Consumable,
			time.Now().Truncate(time.Second),
			time.Now().Truncate(time.Second)).Scan(&packItems[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert pack content for pack %d and item %d: %w",
				packItems[i].PackID,
				packItems[i].ItemID,
				err)
		}
	}
	println("-> Pack Contents Loaded...")
	return nil
}

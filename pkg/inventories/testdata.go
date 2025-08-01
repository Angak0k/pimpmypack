package inventories

import (
	"context"
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

var inventories = dataset.Inventories{
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
		//nolint:execinquery
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

		//nolint:execinquery
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
			inventories[i].UserID = helper.FindUserIDByUsername(users, users[0].Username)
		case 2:
			inventories[i].UserID = helper.FindUserIDByUsername(users, users[1].Username)
		}
	}

	// Insert inventories dataset
	for i := range inventories {
		//nolint:execinquery
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

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

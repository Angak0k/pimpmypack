package accounts

import (
	"context"
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/dataset"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gruntwork-io/terratest/modules/random"
)

var users = []dataset.User{
	{
		Username:            "user-" + random.UniqueId(),
		Email:               "user-" + random.UniqueId() + "@exemple.com",
		Firstname:           "John",
		Lastname:            "Doe",
		Role:                "admin",
		Status:              "active",
		Password:            "password",
		LastPassword:        "password",
		PreferredCurrency:   "EUR",
		PreferredUnitSystem: "METRIC",
	},
	{
		Username:            "user-" + random.UniqueId(),
		Email:               "user-" + random.UniqueId() + "@exemple.com",
		Firstname:           "Jane",
		Lastname:            "Smith",
		Role:                "standard",
		Status:              "pending",
		Password:            "password",
		LastPassword:        "",
		PreferredCurrency:   "EUR",
		PreferredUnitSystem: "METRIC",
	},
	{
		Username:            "user-" + random.UniqueId(),
		Email:               "user-" + random.UniqueId() + "@exemple.com",
		Firstname:           "Alice",
		Lastname:            "Johnson",
		Role:                "standard",
		Status:              "inactive",
		Password:            "password",
		LastPassword:        "old_password",
		PreferredCurrency:   "USD",
		PreferredUnitSystem: "IMPERIAL",
	},
}

func loadingAccountDataset() error {
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
			`INSERT INTO account (username, email, firstname, lastname, role, status, preferred_currency, 
				preferred_unit_system, created_at, updated_at) 
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) 
			RETURNING id;`,
			users[i].Username,
			users[i].Email,
			users[i].Firstname,
			users[i].Lastname,
			users[i].Role,
			users[i].Status,
			users[i].PreferredCurrency,
			users[i].PreferredUnitSystem,
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
			VALUES ($1,$2,$3,$4) RETURNING id;`,
			users[i].ID,
			hashedPassword,
			hashedLastPassword,
			time.Now().Truncate(time.Second)).Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to insert password: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

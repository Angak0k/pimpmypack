package accounts

import (
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/dataset"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gruntwork-io/terratest/modules/random"
)

var users = []dataset.User{
	{
		Username:     "user-" + random.UniqueId(),
		Email:        "user-" + random.UniqueId() + "@exemple.com",
		Firstname:    "John",
		Lastname:     "Doe",
		Role:         "admin",
		Status:       "active",
		Password:     "password",
		LastPassword: "password",
	},
	{
		Username:     "user-" + random.UniqueId(),
		Email:        "user-" + random.UniqueId() + "@exemple.com",
		Firstname:    "Jane",
		Lastname:     "Smith",
		Role:         "standard",
		Status:       "pending",
		Password:     "password",
		LastPassword: "",
	},
	{
		Username:     "user-" + random.UniqueId(),
		Email:        "user-" + random.UniqueId() + "@exemple.com",
		Firstname:    "Alice",
		Lastname:     "Johnson",
		Role:         "standard",
		Status:       "inactive",
		Password:     "password",
		LastPassword: "old_password",
	},
}

func loadingAccountDataset() error {
	// Load accounts dataset
	for i := range users {
		var id uint

		//nolint:execinquery
		err := database.DB().QueryRow(
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
			return err
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
		err = database.DB().QueryRow(
			`INSERT INTO password (user_id, password, last_password, updated_at) 
			VALUES ($1,$2,$3,$4) RETURNING id;`,
			users[i].ID,
			hashedPassword,
			hashedLastPassword,
			time.Now().Truncate(time.Second)).Scan(&id)
		if err != nil {
			return err
		}
	}
	return nil
}

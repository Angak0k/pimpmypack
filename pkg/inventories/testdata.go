package inventories

import (
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
		Username:     fmt.Sprintf("user-%s", random.UniqueId()),
		Email:        fmt.Sprintf("user-%s@exemple.com", random.UniqueId()),
		Firstname:    "Joseph",
		Lastname:     "Doe",
		Role:         "standard",
		Status:       "active",
		Password:     "password",
		LastPassword: "password",
	},
	{
		Username:     fmt.Sprintf("user-%s", random.UniqueId()),
		Email:        fmt.Sprintf("user-%s@exemple.com", random.UniqueId()),
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
		User_id:     1,
		Item_name:   "Backpack",
		Category:    "Outdoor Gear",
		Weight:      950,
		Weight_unit: "METRIC",
		Url:         "https://example.com/backpack",
		Price:       50,
		Currency:    "USD",
	},
	{
		User_id:     1,
		Item_name:   "Tent",
		Category:    "Shelter",
		Weight:      1200,
		Weight_unit: "METRIC",
		Url:         "https://example.com/tent",
		Price:       150,
		Currency:    "USD",
	},
	{
		User_id:     1,
		Item_name:   "Sleeping Bag",
		Category:    "Sleeping",
		Weight:      800,
		Weight_unit: "METRIC",
		Url:         "https://example.com/sleeping-bag",
		Price:       120,
		Currency:    "EUR",
	},
	{
		User_id:     2,
		Item_name:   "Sleeping Bag",
		Category:    "Sleeping",
		Weight:      800,
		Weight_unit: "METRIC",
		Url:         "https://example.com/sleeping-bag",
		Price:       120,
		Currency:    "EUR",
	},
}

func loadingInventoryDataset() error {

	// Load accounts datasetx
	for i := range users {
		var id uint
		err := database.Db().QueryRow("INSERT INTO account (username, email, firstname, lastname, role, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id;",
			users[i].Username, users[i].Email, users[i].Firstname, users[i].Lastname, users[i].Role, users[i].Status, time.Now().Truncate(time.Second), time.Now().Truncate(time.Second)).Scan(&users[i].ID)
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

		err = database.Db().QueryRow("INSERT INTO password (user_id, password, last_password, updated_at) VALUES ($1,$2,$3,$4) RETURNING id;", users[i].ID, hashedPassword, hashedLastPassword, time.Now().Truncate(time.Second)).Scan(&id)
		if err != nil {
			return err
		}
	}

	// Transform inventories dataset by using the real user_id
	for i := range inventories {
		switch inventories[i].User_id {
		case 1:
			inventories[i].User_id = helper.FinUserIDByUsername(users, users[0].Username)
		case 2:
			inventories[i].User_id = helper.FinUserIDByUsername(users, users[0].Username)
		}
	}

	// Insert inventories dataset
	for i := range inventories {
		err := database.Db().QueryRow("INSERT INTO inventory (user_id, item_name, category, description, weight, weight_unit, url, price, currency, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id;",
			inventories[i].User_id, inventories[i].Item_name, inventories[i].Category, inventories[i].Description, inventories[i].Weight, inventories[i].Weight_unit, inventories[i].Url, inventories[i].Price, inventories[i].Currency, time.Now().Truncate(time.Second), time.Now().Truncate(time.Second)).Scan(&inventories[i].ID)
		if err != nil {
			return err
		}
	}
	return nil
}

package packs

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
		Firstname:    "John",
		Lastname:     "Doe",
		Role:         "standard",
		Status:       "active",
		Password:     "password",
		LastPassword: "password",
	},
	{
		Username:     fmt.Sprintf("user-%s", random.UniqueId()),
		Email:        fmt.Sprintf("user-%s@exemple.com", random.UniqueId()),
		Firstname:    "John",
		Lastname:     "Doe",
		Role:         "standard",
		Status:       "active",
		Password:     "password",
		LastPassword: "password",
	},
}

var inventories_user_pack1 = dataset.Inventories{
	{
		User_id:     1,
		Item_name:   "Backpack",
		Category:    "Outdoor Gear",
		Description: "Spacious backpack for hiking",
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
		Description: "Spacious tent for hiking",
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
		Description: "Spacious sleeping bag for hiking",
		Weight:      800,
		Weight_unit: "METRIC",
		Url:         "https://example.com/sleeping-bag",
		Price:       120,
		Currency:    "EUR",
	},
}

var packs = dataset.Packs{
	{
		User_id:          1,
		Pack_name:        "First Pack",
		Pack_description: "Description for the first pack",
	},
	{
		User_id:          1,
		Pack_name:        "Second Pack",
		Pack_description: "Description for the second pack",
	},
	{
		User_id:          2,
		Pack_name:        "Third Pack",
		Pack_description: "Description for the third pack",
	},
	{
		User_id:          1,
		Pack_name:        "Special Pack",
		Pack_description: "Description for the second pack",
	},
}

var packItems = dataset.PackContents{
	{
		Pack_id:    1,
		Item_id:    1,
		Quantity:   2,
		Worn:       true,
		Consumable: false,
	},
	{
		Pack_id:    1,
		Item_id:    2,
		Quantity:   3,
		Worn:       false,
		Consumable: true,
	},
	{
		Pack_id:    2,
		Item_id:    2,
		Quantity:   1,
		Worn:       true,
		Consumable: false,
	},
	{
		Pack_id:    2,
		Item_id:    1,
		Quantity:   4,
		Worn:       true,
		Consumable: true,
	},
	{
		Pack_id:    3,
		Item_id:    1,
		Quantity:   2,
		Worn:       false,
		Consumable: false,
	},
	{
		Pack_id:    4,
		Item_id:    1,
		Quantity:   2,
		Worn:       false,
		Consumable: false,
	},
}

var packWithItems = dataset.PackContentWithItems{
	{
		Pack_content_id:  1,
		Pack_id:          4,
		Item_name:        "Backpack",
		Category:         "Outdoor Gear",
		Item_description: "Spacious backpack for hiking",
		Weight:           950,
		Weight_unit:      "METRIC",
		Item_url:         "https://example.com/backpack",
		Price:            50,
		Currency:         "USD",
		Quantity:         2,
		Worn:             false,
		Consumable:       false,
	},
}

func loadingPackDataset() error {

	// Load accounts dataset
	println("-> Loading accounts and passwords ...")
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

	println("-> Accounts Loaded...")

	// Transform inventories dataset by using the real user_id
	for i := range inventories_user_pack1 {
		switch inventories_user_pack1[i].User_id {
		case 1:
			inventories_user_pack1[i].User_id = helper.FindUserIDByUsername(users, users[0].Username)
		case 2:
			inventories_user_pack1[i].User_id = helper.FindUserIDByUsername(users, users[0].Username)
		}
	}

	// Transform packs dataset
	for i := range packs {
		switch packs[i].User_id {
		case 1:
			packs[i].User_id = helper.FindUserIDByUsername(users, users[0].Username)
		case 2:
			packs[i].User_id = helper.FindUserIDByUsername(users, users[0].Username)
		}
	}

	// Load inventories dataset
	println("-> Loading Inventories...")

	// Insert inventories dataset
	for i := range inventories_user_pack1 {
		err := database.Db().QueryRow("INSERT INTO inventory (user_id, item_name, category, description, weight, weight_unit, url, price, currency, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id;",
			inventories_user_pack1[i].User_id, inventories_user_pack1[i].Item_name, inventories_user_pack1[i].Category, inventories_user_pack1[i].Description, inventories_user_pack1[i].Weight, inventories_user_pack1[i].Weight_unit, inventories_user_pack1[i].Url, inventories_user_pack1[i].Price, inventories_user_pack1[i].Currency, time.Now().Truncate(time.Second), time.Now().Truncate(time.Second)).Scan(&inventories_user_pack1[i].ID)
		if err != nil {
			return err
		}
	}
	println("-> Inventories Loaded...")

	// Load packs dataset
	println("-> Loading Packs...")

	// Insert packs dataset
	for i := range packs {
		err := database.Db().QueryRow("INSERT INTO pack (user_id, pack_name, pack_description, created_at, updated_at) VALUES ($1,$2,$3,$4,$5) RETURNING id;",
			packs[i].User_id, packs[i].Pack_name, packs[i].Pack_description, time.Now().Truncate(time.Second), time.Now().Truncate(time.Second)).Scan(&packs[i].ID)
		if err != nil {
			return err
		}
	}
	println("-> Packs Loaded...")

	// Transform packs_contents dataset

	for i := range packItems {
		switch packItems[i].Pack_id {
		case 1:
			packItems[i].Pack_id = helper.FindPackIDByPackName(packs, "First Pack")
		case 2:
			packItems[i].Pack_id = helper.FindPackIDByPackName(packs, "Second Pack")
		case 3:
			packItems[i].Pack_id = helper.FindPackIDByPackName(packs, "Third Pack")
		case 4:
			packItems[i].Pack_id = helper.FindPackIDByPackName(packs, "Special Pack")
		}
		switch packItems[i].Item_id {
		case 1:
			packItems[i].Item_id = helper.FindItemIDByItemName(inventories_user_pack1, "Backpack")
		case 2:
			packItems[i].Item_id = helper.FindItemIDByItemName(inventories_user_pack1, "Tent")
		case 3:
			packItems[i].Item_id = helper.FindItemIDByItemName(inventories_user_pack1, "Sleeping Bag")
		}
	}

	for i := range packWithItems {
		switch packWithItems[i].Pack_id {
		case 1:
			packWithItems[i].Pack_id = helper.FindPackIDByPackName(packs, "First Pack")
		case 2:
			packWithItems[i].Pack_id = helper.FindPackIDByPackName(packs, "Second Pack")
		case 3:
			packWithItems[i].Pack_id = helper.FindPackIDByPackName(packs, "Third Pack")
		case 4:
			packWithItems[i].Pack_id = helper.FindPackIDByPackName(packs, "Special Pack")
		}
	}

	// Load pack_contents dataset
	println("-> Loading Pack Contents...")

	// Insert pack_contents dataset
	for i := range packItems {
		err := database.Db().QueryRow("INSERT INTO pack_content (pack_id, item_id, quantity, worn, consumable, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id;",
			packItems[i].Pack_id, packItems[i].Item_id, packItems[i].Quantity, packItems[i].Worn, packItems[i].Consumable, time.Now().Truncate(time.Second), time.Now().Truncate(time.Second)).Scan(&packItems[i].ID)
		if err != nil {
			return err
		}
	}
	println("-> Pack Contents Loaded...")
	return nil
}

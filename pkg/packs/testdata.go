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
		WeightUnit:  "METRIC",
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
		WeightUnit:  "METRIC",
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
		WeightUnit:  "METRIC",
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
		PackDescription: "Description for the second pack",
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
		WeightUnit:      "METRIC",
		ItemURL:         "https://example.com/backpack",
		Price:           50,
		Currency:        "USD",
		Quantity:        2,
		Worn:            false,
		Consumable:      false,
	},
}

func loadingPackDataset() error {
	if err := loadAccounts(); err != nil {
		return err
	}
	if err := transformInventories(); err != nil {
		return err
	}
	if err := loadInventories(); err != nil {
		return err
	}
	if err := transformPackContents(); err != nil {
		return err
	}
	return loadPackContents()
}

func loadAccounts() error {
	println("-> Loading accounts and passwords ...")
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
			`INSERT INTO password (user_id, password, last_password, updated_at) VALUES ($1,$2,$3,$4) 
			RETURNING id;`,
			users[i].ID,
			hashedPassword,
			hashedLastPassword,
			time.Now().Truncate(time.Second)).Scan(&id)
		if err != nil {
			return err
		}
	}
	println("-> Accounts Loaded...")
	return nil
}

func transformInventories() error {
	// Transform inventories dataset by using the real user_id
	for i := range inventoriesUserPack1 {
		switch inventoriesUserPack1[i].UserID {
		case 1:
			inventoriesUserPack1[i].UserID = helper.FindUserIDByUsername(users, users[0].Username)
		case 2:
			inventoriesUserPack1[i].UserID = helper.FindUserIDByUsername(users, users[0].Username)
		}
	}

	// Transform packs dataset
	for i := range packs {
		switch packs[i].UserID {
		case 1:
			packs[i].UserID = helper.FindUserIDByUsername(users, users[0].Username)
		case 2:
			packs[i].UserID = helper.FindUserIDByUsername(users, users[0].Username)
		}
	}
	return nil
}

func loadInventories() error {
	println("-> Loading Inventories...")
	for i := range inventoriesUserPack1 {
		//nolint:execinquery
		err := database.DB().QueryRow(
			`INSERT INTO inventory 
			(user_id, item_name, category, description, weight, weight_unit, url, price, currency, 
				created_at, updated_at) 
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) 
			RETURNING id;`,
			inventoriesUserPack1[i].UserID,
			inventoriesUserPack1[i].ItemName,
			inventoriesUserPack1[i].Category,
			inventoriesUserPack1[i].Description,
			inventoriesUserPack1[i].Weight,
			inventoriesUserPack1[i].WeightUnit,
			inventoriesUserPack1[i].URL,
			inventoriesUserPack1[i].Price,
			inventoriesUserPack1[i].Currency,
			time.Now().Truncate(time.Second),
			time.Now().Truncate(time.Second)).Scan(&inventoriesUserPack1[i].ID)
		if err != nil {
			return err
		}
	}
	println("-> Inventories Loaded...")
	println("-> Loading Packs...")

	// Insert packs dataset
	for i := range packs {
		//nolint:execinquery
		err := database.DB().QueryRow(
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
			return err
		}
	}
	println("-> Packs Loaded...")
	return nil
}

func transformPackContents() error {
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
	return nil
}

func loadPackContents() error {
	println("-> Loading Pack Contents...")
	for i := range packItems {
		//nolint:execinquery
		err := database.DB().QueryRow(
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
			return err
		}
	}
	println("-> Pack Contents Loaded...")
	return nil
}

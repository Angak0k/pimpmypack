package accounts

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/gruntwork-io/terratest/modules/random"
)

var users = []User{
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

// cleanupAccountDataset removes all test data created by loadingAccountDataset
func cleanupAccountDataset() error {
	ctx := context.Background()

	println("-> Cleaning up account test data...")

	// Delete users (passwords will cascade delete)
	for _, user := range users {
		if user.ID != 0 {
			_, err := database.DB().ExecContext(ctx, "DELETE FROM account WHERE id = $1", user.ID)
			if err != nil {
				return fmt.Errorf("failed to delete user %d: %w", user.ID, err)
			}
		}
	}

	println("-> Account test data cleaned up...")
	return nil
}

// setupLoginTestData creates a test user and returns router, user, and marshaled login JSON
func setupLoginTestData(t *testing.T) (*gin.Engine, User, []byte) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/login", Login)

	newUser := User{
		Username:  "user-" + random.UniqueId(),
		Password:  "password2",
		Email:     "newuser2@pmp.com",
		Firstname: "Jules",
		Lastname:  "Doe",
		Role:      "standard",
		Status:    "active",
		CreatedAt: time.Now().Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}

	ctx := context.Background()
	var id int
	err := database.DB().QueryRowContext(ctx,
		`INSERT INTO account (username, email, firstname, lastname, role, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id;`,
		newUser.Username, newUser.Email, newUser.Firstname, newUser.Lastname,
		newUser.Role, newUser.Status, newUser.CreatedAt, newUser.UpdatedAt).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	hashedPassword, err := security.HashPassword(newUser.Password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO password (user_id, password, updated_at) VALUES ($1,$2,$3) RETURNING id;`,
		id, hashedPassword, newUser.UpdatedAt).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert password: %v", err)
	}

	userLogin := LoginInput{
		Username: newUser.Username,
		Password: newUser.Password,
	}
	jsonData, err := json.Marshal(userLogin)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}

	return router, newUser, jsonData
}

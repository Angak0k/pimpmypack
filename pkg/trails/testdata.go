package trails

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/accounts"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gruntwork-io/terratest/modules/random"
)

var testUsers = []accounts.User{
	{
		Username:     "trailuser-" + random.UniqueId(),
		Email:        "trailuser-" + random.UniqueId() + "@example.com",
		Firstname:    "Trail",
		Lastname:     "Tester",
		Role:         "admin",
		Status:       "active",
		Password:     "password",
		LastPassword: "password",
	},
}

var testTrails = Trails{
	{
		Name:      "Test Trail Alpha",
		Country:   "France",
		Continent: "Europe",
	},
	{
		Name:      "Test Trail Beta",
		Country:   "United States",
		Continent: "North America",
	},
	{
		Name:      "Test Trail Gamma",
		Country:   "Japan",
		Continent: "Asia",
	},
}

func loadingTrailDataset() error {
	tx, err := database.DB().BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err := loadTrailTestAccounts(tx); err != nil {
		return fmt.Errorf("failed to load accounts: %w", err)
	}

	if err := loadTrailTestData(tx); err != nil {
		return fmt.Errorf("failed to load trails: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func loadTrailTestAccounts(tx *sql.Tx) error {
	println("-> Loading trail test accounts...")
	for i := range testUsers {
		var existingID int
		err := tx.QueryRowContext(context.Background(),
			"SELECT id FROM account WHERE username = $1", testUsers[i].Username).Scan(&existingID)
		if err == nil {
			if existingID < 0 {
				return fmt.Errorf("invalid user ID: negative value %d", existingID)
			}
			testUsers[i].ID = uint(existingID)
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		err = tx.QueryRowContext(context.Background(),
			`INSERT INTO account (username, email, firstname, lastname, role, status, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id;`,
			testUsers[i].Username, testUsers[i].Email,
			testUsers[i].Firstname, testUsers[i].Lastname,
			testUsers[i].Role, testUsers[i].Status,
			time.Now().Truncate(time.Second),
			time.Now().Truncate(time.Second)).Scan(&testUsers[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}

		hashedPassword, err := security.HashPassword(testUsers[i].Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		hashedLastPassword, err := security.HashPassword(testUsers[i].LastPassword)
		if err != nil {
			return fmt.Errorf("failed to hash last password: %w", err)
		}

		var passwordID uint
		err = tx.QueryRowContext(context.Background(),
			`INSERT INTO password (user_id, password, last_password, updated_at) VALUES ($1,$2,$3,$4)
			RETURNING id;`,
			testUsers[i].ID, hashedPassword, hashedLastPassword,
			time.Now().Truncate(time.Second)).Scan(&passwordID)
		if err != nil {
			return fmt.Errorf("failed to insert password: %w", err)
		}
	}
	println("-> Trail test accounts loaded...")
	return nil
}

func loadTrailTestData(tx *sql.Tx) error {
	println("-> Loading trail test data...")
	now := time.Now().Truncate(time.Second)

	for i := range testTrails {
		err := tx.QueryRowContext(context.Background(),
			`INSERT INTO trail (name, country, continent, distance_km, url, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id;`,
			testTrails[i].Name, testTrails[i].Country, testTrails[i].Continent,
			testTrails[i].DistanceKm, testTrails[i].URL,
			now, now).Scan(&testTrails[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert trail %q: %w", testTrails[i].Name, err)
		}
	}

	println("-> Trail test data loaded...")
	return nil
}

func cleanupTrailDataset() error {
	ctx := context.Background()
	println("-> Cleaning up trail test data...")

	for _, t := range testTrails {
		if t.ID != 0 {
			_, err := database.DB().ExecContext(ctx, "DELETE FROM trail WHERE id = $1", t.ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("failed to delete trail %d: %w", t.ID, err)
			}
		}
	}

	for _, user := range testUsers {
		if user.ID != 0 {
			_, err := database.DB().ExecContext(ctx, "DELETE FROM account WHERE id = $1", user.ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("failed to delete user %d: %w", user.ID, err)
			}
		}
	}

	println("-> Trail test data cleaned up...")
	return nil
}

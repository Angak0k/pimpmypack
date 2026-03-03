package seed

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/security"
)

// itemKey identifies an inventory item by its composite key.
type itemKey struct {
	itemName    string
	category    string
	description string
}

// Seed populates the database with realistic development data.
// It is idempotent: if the seed user already exists, it skips.
func Seed(ctx context.Context) error {
	exists, err := userExists(ctx, defaultUser.username)
	if err != nil {
		return fmt.Errorf("failed to check for existing user: %w", err)
	}
	if exists {
		log.Printf("[SEED] User '%s' already exists, skipping seed.",
			defaultUser.username)
		return nil
	}

	log.Println("[SEED] Starting database seed...")

	tx, err := database.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().Truncate(time.Second)

	userID, err := createAccount(ctx, tx, now)
	if err != nil {
		return err
	}

	if err = createPassword(ctx, tx, userID, now); err != nil {
		return err
	}

	itemIDs, err := createInventoryItems(ctx, tx, userID, now)
	if err != nil {
		return err
	}

	packIDs, err := createPacks(ctx, tx, userID, now)
	if err != nil {
		return err
	}

	contentCount, err := createPackContents(ctx, tx, packIDs, itemIDs, now)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit seed transaction: %w", err)
	}

	log.Printf("[SEED] Database seed completed successfully! "+
		"Created %d inventory items, %d packs, %d pack contents.",
		len(itemIDs), len(seedPackDefs), contentCount)

	return nil
}

func userExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := database.DB().QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM account WHERE username = $1)`,
		username,
	).Scan(&exists)
	return exists, err
}

func createAccount(
	ctx context.Context, tx *sql.Tx, now time.Time,
) (uint, error) {
	var userID uint
	err := tx.QueryRowContext(ctx,
		`INSERT INTO account
		(username, email, firstname, lastname, role, status,
		 preferred_currency, preferred_unit_system,
		 created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id`,
		defaultUser.username, defaultUser.email,
		defaultUser.firstname, defaultUser.lastname,
		defaultUser.role, defaultUser.status,
		"EUR", "METRIC", now, now,
	).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create account: %w", err)
	}
	return userID, nil
}

func createPassword(
	ctx context.Context, tx *sql.Tx, userID uint, now time.Time,
) error {
	hashedPassword, err := security.HashPassword(defaultUser.password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO password (user_id, password, updated_at)
		VALUES ($1, $2, $3)`,
		userID, hashedPassword, now,
	)
	if err != nil {
		return fmt.Errorf("failed to create password: %w", err)
	}
	return nil
}

func createInventoryItems(
	ctx context.Context, tx *sql.Tx, userID uint, now time.Time,
) (map[itemKey]uint, error) {
	itemIDs := make(map[itemKey]uint, len(seedInventoryItems))

	for _, item := range seedInventoryItems {
		key := itemKey{
			itemName:    item.itemName,
			category:    item.category,
			description: item.description,
		}
		if _, ok := itemIDs[key]; ok {
			continue
		}

		var id uint
		err := tx.QueryRowContext(ctx,
			`INSERT INTO inventory
			(user_id, item_name, category, description,
			 weight, url, price, currency,
			 created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			RETURNING id`,
			userID, item.itemName, item.category,
			item.description, item.weight, item.url,
			item.price, item.currency, now, now,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to create inventory item %q: %w",
				item.itemName, err)
		}
		itemIDs[key] = id
	}

	return itemIDs, nil
}

func createPacks(
	ctx context.Context, tx *sql.Tx, userID uint, now time.Time,
) ([]uint, error) {
	packIDs := make([]uint, len(seedPackDefs))

	for i, p := range seedPackDefs {
		var id uint
		err := tx.QueryRowContext(ctx,
			`INSERT INTO pack
			(user_id, pack_name, pack_description,
			 season, trail, adventure,
			 created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id`,
			userID, p.packName, p.packDescription,
			p.season, p.trail, p.adventure, now, now,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to create pack %q: %w", p.packName, err)
		}
		packIDs[i] = id

		if p.isFavorite {
			_, err = tx.ExecContext(ctx,
				`UPDATE pack SET is_favorite = true WHERE id = $1`,
				id,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to set favorite on pack %q: %w",
					p.packName, err)
			}
		}
	}

	return packIDs, nil
}

func createPackContents(
	ctx context.Context,
	tx *sql.Tx,
	packIDs []uint,
	itemIDs map[itemKey]uint,
	now time.Time,
) (int, error) {
	count := 0
	for _, pc := range seedPackContentDefs {
		key := itemKey{
			itemName:    pc.itemName,
			category:    pc.category,
			description: pc.description,
		}

		itemID, ok := itemIDs[key]
		if !ok {
			return 0, fmt.Errorf(
				"pack content references unknown item: %q/%q/%q",
				pc.itemName, pc.category, pc.description)
		}

		var dupExists bool
		err := tx.QueryRowContext(ctx,
			`SELECT EXISTS(
				SELECT 1 FROM pack_content
				WHERE pack_id = $1 AND item_id = $2)`,
			packIDs[pc.packIndex], itemID,
		).Scan(&dupExists)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf(
				"failed to check pack content duplicate: %w", err)
		}
		if dupExists {
			continue
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO pack_content
			(pack_id, item_id, quantity, worn, consumable,
			 created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			packIDs[pc.packIndex], itemID,
			pc.quantity, pc.worn, pc.consumable, now, now,
		)
		if err != nil {
			return 0, fmt.Errorf(
				"failed to create pack content for %q: %w",
				pc.itemName, err)
		}
		count++
	}
	return count, nil
}

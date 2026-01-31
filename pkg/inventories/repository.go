package inventories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Angak0k/pimpmypack/pkg/database"
)

// ErrNoItemFound is returned when no item is found for a given request.
var ErrNoItemFound = errors.New("no item found")

// returnInventories retrieves all inventory items from the database
func returnInventories(ctx context.Context) (*Inventories, error) {
	var inventories Inventories

	rows, err := database.DB().QueryContext(ctx,
		`SELECT id,
			user_id,
			item_name,
			category,
			description,
			weight,
			url,
			price,
			currency,
			created_at,
			updated_at
		FROM inventory;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoItemFound
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var inventory Inventory
		err := rows.Scan(
			&inventory.ID,
			&inventory.UserID,
			&inventory.ItemName,
			&inventory.Category,
			&inventory.Description,
			&inventory.Weight,
			&inventory.URL,
			&inventory.Price,
			&inventory.Currency,
			&inventory.CreatedAt,
			&inventory.UpdatedAt)
		if err != nil {
			return nil, err
		}
		inventories = append(inventories, inventory)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &inventories, nil
}

// returnInventoriesByUserID retrieves all inventory items for a specific user
func returnInventoriesByUserID(ctx context.Context, userID uint) (*Inventories, error) {
	var inventories Inventories

	rows, err := database.DB().QueryContext(ctx,
		`SELECT id,
			user_id,
			item_name,
			category,
			description,
			weight,
			url,
			price,
			currency,
			created_at,
			updated_at
		FROM inventory WHERE user_id = $1 ORDER BY category;`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var inventory Inventory
		err := rows.Scan(
			&inventory.ID,
			&inventory.UserID,
			&inventory.ItemName,
			&inventory.Category,
			&inventory.Description,
			&inventory.Weight,
			&inventory.URL,
			&inventory.Price,
			&inventory.Currency,
			&inventory.CreatedAt,
			&inventory.UpdatedAt)
		if err != nil {
			return nil, err
		}
		inventories = append(inventories, inventory)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &inventories, nil
}

// findInventoryByID retrieves a single inventory item by ID
func findInventoryByID(ctx context.Context, id uint) (*Inventory, error) {
	var inventory Inventory

	row := database.DB().QueryRowContext(ctx,
		`SELECT id,
			user_id,
			item_name,
			category,
			description,
			weight,
			url,
			price,
			currency,
			created_at,
			updated_at
		FROM inventory WHERE id = $1;`,
		id)
	err := row.Scan(
		&inventory.ID,
		&inventory.UserID,
		&inventory.ItemName,
		&inventory.Category,
		&inventory.Description,
		&inventory.Weight,
		&inventory.URL,
		&inventory.Price,
		&inventory.Currency,
		&inventory.CreatedAt,
		&inventory.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoItemFound
		}
		return nil, err
	}

	return &inventory, nil
}

// findInventoryItemByAttributes finds an existing inventory item for a user
// by exact match on item_name, category, and description
func findInventoryItemByAttributes(
	ctx context.Context,
	userID uint,
	itemName, category, description string,
) (*Inventory, error) {
	var inventory Inventory

	query := `
		SELECT id,
			user_id,
			item_name,
			category,
			description,
			weight,
			url,
			price,
			currency,
			created_at,
			updated_at
		FROM inventory
		WHERE user_id = $1 AND item_name = $2 AND category = $3 AND description = $4
		LIMIT 1`

	row := database.DB().QueryRowContext(ctx, query, userID, itemName, category, description)
	err := row.Scan(
		&inventory.ID,
		&inventory.UserID,
		&inventory.ItemName,
		&inventory.Category,
		&inventory.Description,
		&inventory.Weight,
		&inventory.URL,
		&inventory.Price,
		&inventory.Currency,
		&inventory.CreatedAt,
		&inventory.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoItemFound
		}
		return nil, err
	}

	return &inventory, nil
}

// insertInventory inserts a new inventory item into the database
func insertInventory(ctx context.Context, i *Inventory) error {
	if i == nil {
		return errors.New("payload is empty")
	}

	//nolint:execinquery
	err := database.DB().QueryRowContext(ctx,
		`INSERT INTO inventory
		(user_id, item_name, category, description, weight, url, price, currency, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id;`,
		i.UserID,
		i.ItemName,
		i.Category,
		i.Description,
		i.Weight,
		i.URL,
		i.Price,
		i.Currency,
		i.CreatedAt,
		i.UpdatedAt).Scan(&i.ID)

	if err != nil {
		return err
	}

	return nil
}

// updateInventoryByID updates an existing inventory item in the database
func updateInventoryByID(ctx context.Context, id uint, i *Inventory) error {
	if i == nil {
		return errors.New("payload is empty")
	}

	statement, err := database.DB().PrepareContext(ctx,
		`UPDATE inventory
		SET user_id=$1,
			item_name=$2,
			category=$3,
			description=$4,
			weight=$5,
			url=$6,
			price=$7,
			currency=$8,
			updated_at=$9
		WHERE id=$10;`)
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.ExecContext(ctx,
		i.UserID,
		i.ItemName,
		i.Category,
		i.Description,
		i.Weight,
		i.URL,
		i.Price,
		i.Currency,
		i.UpdatedAt,
		id)
	if err != nil {
		return err
	}

	return nil
}

// deleteInventoryByID removes an inventory item from the database
func deleteInventoryByID(ctx context.Context, id uint) error {
	statement, err := database.DB().PrepareContext(ctx, "DELETE FROM inventory WHERE id=$1;")
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.ExecContext(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

// checkInventoryOwnership verifies if an inventory item belongs to a specific user
func checkInventoryOwnership(ctx context.Context, id uint, userID uint) (bool, error) {
	var rows int

	row := database.DB().QueryRowContext(ctx,
		"SELECT COUNT(id) FROM inventory WHERE id = $1 AND user_id = $2;", id, userID)
	err := row.Scan(&rows)
	if err != nil {
		return false, err
	}

	if rows == 0 {
		return false, nil
	}

	return true, nil
}

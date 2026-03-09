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
		`SELECT i.id,
			i.user_id,
			i.item_name,
			i.category,
			i.description,
			i.weight,
			i.url,
			i.price,
			i.currency,
			CASE WHEN ii.item_id IS NOT NULL THEN true ELSE false END as has_image,
			(SELECT COUNT(DISTINCT pc.pack_id) FROM pack_content pc WHERE pc.item_id = i.id) as pack_count,
			i.created_at,
			i.updated_at
		FROM inventory i
		LEFT JOIN inventory_images ii ON i.id = ii.item_id;`)
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
			&inventory.HasImage,
			&inventory.PackCount,
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
		`SELECT i.id,
			i.user_id,
			i.item_name,
			i.category,
			i.description,
			i.weight,
			i.url,
			i.price,
			i.currency,
			CASE WHEN ii.item_id IS NOT NULL THEN true ELSE false END as has_image,
			(SELECT COUNT(DISTINCT pc.pack_id) FROM pack_content pc WHERE pc.item_id = i.id) as pack_count,
			i.created_at,
			i.updated_at
		FROM inventory i
		LEFT JOIN inventory_images ii ON i.id = ii.item_id
		WHERE i.user_id = $1 ORDER BY i.category;`,
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
			&inventory.HasImage,
			&inventory.PackCount,
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
		`SELECT i.id,
			i.user_id,
			i.item_name,
			i.category,
			i.description,
			i.weight,
			i.url,
			i.price,
			i.currency,
			CASE WHEN ii.item_id IS NOT NULL THEN true ELSE false END as has_image,
			(SELECT COUNT(DISTINCT pc.pack_id) FROM pack_content pc WHERE pc.item_id = i.id) as pack_count,
			i.created_at,
			i.updated_at
		FROM inventory i
		LEFT JOIN inventory_images ii ON i.id = ii.item_id
		WHERE i.id = $1;`,
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
		&inventory.HasImage,
		&inventory.PackCount,
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
		SELECT i.id,
			i.user_id,
			i.item_name,
			i.category,
			i.description,
			i.weight,
			i.url,
			i.price,
			i.currency,
			CASE WHEN ii.item_id IS NOT NULL THEN true ELSE false END as has_image,
			(SELECT COUNT(DISTINCT pc.pack_id) FROM pack_content pc WHERE pc.item_id = i.id) as pack_count,
			i.created_at,
			i.updated_at
		FROM inventory i
		LEFT JOIN inventory_images ii ON i.id = ii.item_id
		WHERE i.user_id = $1 AND i.item_name = $2 AND i.category = $3 AND i.description = $4
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
		&inventory.HasImage,
		&inventory.PackCount,
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

// mergeInventoryItems merges the source inventory item into the target inventory item within a transaction.
// It updates the target item properties, consolidates pack_content references, handles images,
// and deletes the source item.
func mergeInventoryItems(ctx context.Context, req *MergeInventoryRequest) (*Inventory, error) {
	tx, err := database.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	sourceID := uint(req.SourceItemID)
	targetID := uint(req.TargetItemID)

	// Step 1: Update target item with merged property values
	_, err = tx.ExecContext(ctx,
		`UPDATE inventory
		SET item_name = $1, category = $2, description = $3, weight = $4,
			url = $5, price = $6, currency = $7, updated_at = NOW()
		WHERE id = $8;`,
		req.ItemName, req.Category, req.Description, req.Weight,
		req.URL, req.Price, req.Currency, targetID)
	if err != nil {
		return nil, err
	}

	// Step 2: Sum quantities for packs containing both items
	_, err = tx.ExecContext(ctx,
		`UPDATE pack_content AS t
		SET quantity = t.quantity + s.quantity, updated_at = NOW()
		FROM pack_content AS s
		WHERE s.item_id = $1 AND t.item_id = $2 AND s.pack_id = t.pack_id;`,
		sourceID, targetID)
	if err != nil {
		return nil, err
	}

	// Step 3: Delete overlapping source pack_content rows
	_, err = tx.ExecContext(ctx,
		`DELETE FROM pack_content WHERE item_id = $1
		AND pack_id IN (SELECT pack_id FROM pack_content WHERE item_id = $2);`,
		sourceID, targetID)
	if err != nil {
		return nil, err
	}

	// Step 4: Reassign remaining source pack_content rows to target
	_, err = tx.ExecContext(ctx,
		`UPDATE pack_content SET item_id = $1 WHERE item_id = $2;`,
		targetID, sourceID)
	if err != nil {
		return nil, err
	}

	// Step 5: Handle image based on image_source
	switch req.ImageSource {
	case "source":
		// Delete target image, then update source image's item_id to target
		_, err = tx.ExecContext(ctx, `DELETE FROM inventory_images WHERE item_id = $1;`, targetID)
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(ctx, `UPDATE inventory_images SET item_id = $1 WHERE item_id = $2;`, targetID, sourceID)
		if err != nil {
			return nil, err
		}
	case "target":
		// Delete source image (if any) so it doesn't block source item deletion
		_, err = tx.ExecContext(ctx, `DELETE FROM inventory_images WHERE item_id = $1;`, sourceID)
		if err != nil {
			return nil, err
		}
	case "none":
		// Delete both images
		_, err = tx.ExecContext(ctx, `DELETE FROM inventory_images WHERE item_id IN ($1, $2);`, sourceID, targetID)
		if err != nil {
			return nil, err
		}
	}

	// Step 6: Delete source inventory item (CASCADE cleans up remaining refs)
	_, err = tx.ExecContext(ctx, `DELETE FROM inventory WHERE id = $1;`, sourceID)
	if err != nil {
		return nil, err
	}

	// Step 7: Return updated target item (query it back after all updates)
	var inventory Inventory
	err = tx.QueryRowContext(ctx,
		`SELECT i.id,
			i.user_id,
			i.item_name,
			i.category,
			i.description,
			i.weight,
			i.url,
			i.price,
			i.currency,
			CASE WHEN ii.item_id IS NOT NULL THEN true ELSE false END as has_image,
			(SELECT COUNT(DISTINCT pc.pack_id) FROM pack_content pc WHERE pc.item_id = i.id) as pack_count,
			i.created_at,
			i.updated_at
		FROM inventory i
		LEFT JOIN inventory_images ii ON i.id = ii.item_id
		WHERE i.id = $1;`,
		targetID).Scan(
		&inventory.ID,
		&inventory.UserID,
		&inventory.ItemName,
		&inventory.Category,
		&inventory.Description,
		&inventory.Weight,
		&inventory.URL,
		&inventory.Price,
		&inventory.Currency,
		&inventory.HasImage,
		&inventory.PackCount,
		&inventory.CreatedAt,
		&inventory.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &inventory, nil
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

package packs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/inventories"
)

// Pack queries

func returnPacks(ctx context.Context) (Packs, error) {
	var packs Packs

	rows, err := database.DB().QueryContext(ctx,
		`SELECT p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code, p.created_at, p.updated_at,
		COALESCE(SUM(pc.quantity), 0) as items_count,
		COALESCE(SUM(i.weight * pc.quantity), 0) as total_weight,
		CASE WHEN pi.pack_id IS NOT NULL THEN true ELSE false END as has_image
		FROM pack p
		LEFT JOIN pack_content pc ON p.id = pc.pack_id
		LEFT JOIN inventory i ON pc.item_id = i.id
		LEFT JOIN pack_images pi ON p.id = pi.pack_id
		GROUP BY p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code, p.created_at, p.updated_at, pi.pack_id
		ORDER BY p.id;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pack Pack
		err := rows.Scan(
			&pack.ID,
			&pack.UserID,
			&pack.PackName,
			&pack.PackDescription,
			&pack.SharingCode,
			&pack.CreatedAt,
			&pack.UpdatedAt,
			&pack.PackItemsCount,
			&pack.PackWeight,
			&pack.HasImage,
		)
		if err != nil {
			return nil, err
		}
		packs = append(packs, pack)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return packs, nil
}

// FindPackByID finds a pack by its ID
func FindPackByID(ctx context.Context, id uint) (*Pack, error) {
	var pack Pack

	row := database.DB().QueryRowContext(ctx,
		`SELECT p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code, p.created_at, p.updated_at,
		COALESCE(SUM(pc.quantity), 0) as items_count,
		COALESCE(SUM(i.weight * pc.quantity), 0) as total_weight,
		CASE WHEN pi.pack_id IS NOT NULL THEN true ELSE false END as has_image
		FROM pack p
		LEFT JOIN pack_content pc ON p.id = pc.pack_id
		LEFT JOIN inventory i ON pc.item_id = i.id
		LEFT JOIN pack_images pi ON p.id = pi.pack_id
		WHERE p.id = $1
		GROUP BY p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code, p.created_at, p.updated_at, pi.pack_id;`,
		id)
	err := row.Scan(
		&pack.ID,
		&pack.UserID,
		&pack.PackName,
		&pack.PackDescription,
		&pack.SharingCode,
		&pack.CreatedAt,
		&pack.UpdatedAt,
		&pack.PackItemsCount,
		&pack.PackWeight,
		&pack.HasImage)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPackNotFound
		}
		return nil, err
	}

	return &pack, nil
}

func findPacksByUserID(ctx context.Context, id uint) (*Packs, error) {
	var packs Packs

	rows, err := database.DB().QueryContext(ctx, `
		SELECT p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code, p.created_at, p.updated_at,
		COALESCE(SUM(pc.quantity), 0) as items_count,
		COALESCE(SUM(i.weight * pc.quantity), 0) as total_weight,
		CASE WHEN pi.pack_id IS NOT NULL THEN true ELSE false END as has_image
		FROM pack p
		LEFT JOIN pack_content pc ON p.id = pc.pack_id
		LEFT JOIN inventory i ON pc.item_id = i.id
		LEFT JOIN pack_images pi ON p.id = pi.pack_id
		WHERE p.user_id = $1
		GROUP BY p.id, p.user_id, p.pack_name, p.pack_description, p.sharing_code,
		         p.created_at, p.updated_at, pi.pack_id;`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pack Pack
		err := rows.Scan(
			&pack.ID,
			&pack.UserID,
			&pack.PackName,
			&pack.PackDescription,
			&pack.SharingCode,
			&pack.CreatedAt,
			&pack.UpdatedAt,
			&pack.PackItemsCount,
			&pack.PackWeight,
			&pack.HasImage)
		if err != nil {
			return nil, err
		}
		packs = append(packs, pack)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(packs) == 0 {
		return nil, ErrPackContentNotFound
	}

	return &packs, nil
}

func findPackIDBySharingCode(ctx context.Context, sharingCode string) (uint, error) {
	var packID uint
	row := database.DB().QueryRowContext(ctx, "SELECT id FROM pack WHERE sharing_code = $1;", sharingCode)
	err := row.Scan(&packID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return packID, nil
}

// Helper function to check if a pack exists
func checkPackExists(ctx context.Context, id uint) (bool, error) {
	var count int
	err := database.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM pack WHERE id = $1", id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Pack writes

func insertPack(ctx context.Context, p *Pack) error {
	if p == nil {
		return errors.New("payload is empty")
	}
	p.CreatedAt = time.Now().Truncate(time.Second)
	p.UpdatedAt = time.Now().Truncate(time.Second)
	// SharingCode is now NULL by default (pack is private)

	//nolint:execinquery
	err := database.DB().QueryRowContext(ctx,
		`INSERT INTO pack (user_id, pack_name, pack_description, sharing_code, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id;`,
		p.UserID,
		p.PackName,
		p.PackDescription,
		p.SharingCode,
		p.CreatedAt,
		p.UpdatedAt).Scan(&p.ID)
	if err != nil {
		return err
	}

	return nil
}

func updatePackByID(ctx context.Context, id uint, p *Pack) error {
	if p == nil {
		return errors.New("payload is empty")
	}
	p.UpdatedAt = time.Now().Truncate(time.Second)

	statement, err := database.DB().PrepareContext(ctx,
		`UPDATE pack SET user_id=$1, pack_name=$2, pack_description=$3, updated_at=$4
		WHERE id=$5;`)
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.ExecContext(ctx, p.UserID, p.PackName, p.PackDescription, p.UpdatedAt, id)
	if err != nil {
		return err
	}

	return nil
}

func deletePackByID(ctx context.Context, id uint) error {
	statement, err := database.DB().PrepareContext(ctx, "DELETE FROM pack WHERE id=$1;")
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

// Pack sharing

// CheckPackOwnership verifies if a user owns a specific pack
func CheckPackOwnership(ctx context.Context, id uint, userID uint) (bool, error) {
	var rows int

	row := database.DB().QueryRowContext(ctx,
		"SELECT COUNT(id) FROM pack WHERE id = $1 AND user_id = $2;", id, userID)
	err := row.Scan(&rows)
	if err != nil {
		return false, err
	}

	if rows == 0 {
		return false, nil
	}

	return true, nil
}

// sharePackByID generates and sets a sharing code for a pack (idempotent)
func sharePackByID(ctx context.Context, packID uint, userID uint) (string, error) {
	// First check ownership
	owns, err := CheckPackOwnership(ctx, packID, userID)
	if err != nil {
		return "", err
	}
	if !owns {
		return "", errors.New("pack does not belong to user")
	}

	// Check if pack already has a sharing code (idempotent behavior)
	var existingCode *string
	err = database.DB().QueryRowContext(ctx,
		"SELECT sharing_code FROM pack WHERE id = $1;", packID).Scan(&existingCode)
	if err != nil {
		return "", err
	}

	// If already shared, return existing code
	if existingCode != nil && *existingCode != "" {
		return *existingCode, nil
	}

	// Generate new sharing code
	sharingCode, err := helper.GenerateRandomCode(30)
	if err != nil {
		return "", errors.New("failed to generate sharing code")
	}

	// Update pack with new sharing code
	statement, err := database.DB().PrepareContext(ctx,
		"UPDATE pack SET sharing_code = $1, updated_at = $2 WHERE id = $3;")
	if err != nil {
		return "", err
	}
	defer statement.Close()

	_, err = statement.ExecContext(ctx, sharingCode, time.Now().Truncate(time.Second), packID)
	if err != nil {
		return "", err
	}

	return sharingCode, nil
}

// unsharePackByID removes the sharing code from a pack (idempotent)
func unsharePackByID(ctx context.Context, packID uint, userID uint) error {
	// First check ownership
	owns, err := CheckPackOwnership(ctx, packID, userID)
	if err != nil {
		return err
	}
	if !owns {
		return errors.New("pack does not belong to user")
	}

	// Set sharing_code to NULL (idempotent - no error if already NULL)
	statement, err := database.DB().PrepareContext(ctx,
		"UPDATE pack SET sharing_code = NULL, updated_at = $1 WHERE id = $2;")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.ExecContext(ctx, time.Now().Truncate(time.Second), packID)
	if err != nil {
		return err
	}

	return nil
}

// returnPackInfoBySharingCode retrieves pack metadata using sharing code.
// Returns nil if pack not found or sharing_code is NULL.
func returnPackInfoBySharingCode(ctx context.Context, sharingCode string) (*SharedPackInfo, error) {
	query := `
		SELECT p.id, p.pack_name, p.pack_description, p.created_at,
		CASE WHEN pi.pack_id IS NOT NULL THEN true ELSE false END as has_image
		FROM pack p
		LEFT JOIN pack_images pi ON p.id = pi.pack_id
		WHERE p.sharing_code = $1 AND p.sharing_code IS NOT NULL
	`

	var packInfo SharedPackInfo
	err := database.DB().QueryRowContext(ctx, query, sharingCode).Scan(
		&packInfo.ID,
		&packInfo.PackName,
		&packInfo.PackDescription,
		&packInfo.CreatedAt,
		&packInfo.HasImage,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPackNotFound
		}
		return nil, fmt.Errorf("failed to fetch pack info by sharing code: %w", err)
	}

	return &packInfo, nil
}

// returnSharedPack retrieves both pack metadata and contents for a shared pack.
// This function is used by the public shared pack endpoint.
func returnSharedPack(ctx context.Context, sharingCode string) (*SharedPackResponse, error) {
	// Get pack metadata
	packInfo, err := returnPackInfoBySharingCode(ctx, sharingCode)
	if err != nil {
		return nil, err
	}

	// Get pack contents
	packContents, err := returnPackContentsByPackID(ctx, packInfo.ID)
	if err != nil {
		if errors.Is(err, ErrPackContentNotFound) {
			// Pack exists but has no contents - return empty array
			return &SharedPackResponse{
				Pack:     *packInfo,
				Contents: PackContentWithItems{},
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch pack contents: %w", err)
	}

	// Handle nil contents as empty array
	if packContents == nil {
		packContents = &PackContentWithItems{}
	}

	return &SharedPackResponse{
		Pack:     *packInfo,
		Contents: *packContents,
	}, nil
}

// Pack content queries

func returnPackContents(ctx context.Context) (*PackContents, error) {
	var packContents PackContents

	rows, err := database.DB().QueryContext(ctx,
		`SELECT id, pack_id, item_id, quantity, worn, consumable, created_at, updated_at
		FROM pack_content;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var packContent PackContent
		err := rows.Scan(
			&packContent.ID,
			&packContent.PackID,
			&packContent.ItemID,
			&packContent.Quantity,
			&packContent.Worn,
			&packContent.Consumable,
			&packContent.CreatedAt,
			&packContent.UpdatedAt)
		if err != nil {
			return nil, err
		}
		packContents = append(packContents, packContent)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &packContents, nil
}

func findPackContentByID(ctx context.Context, id uint) (*PackContent, error) {
	var packcontent PackContent

	row := database.DB().QueryRowContext(ctx,
		`SELECT id, pack_id, item_id, quantity, worn, consumable, created_at, updated_at
		FROM pack_content
		WHERE id = $1;`,
		id)
	err := row.Scan(
		&packcontent.ID,
		&packcontent.PackID,
		&packcontent.ItemID,
		&packcontent.Quantity,
		&packcontent.Worn,
		&packcontent.Consumable,
		&packcontent.CreatedAt,
		&packcontent.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPackContentNotFound
		}
		return nil, err
	}

	return &packcontent, nil
}

func returnPackContentsByPackID(ctx context.Context, id uint) (*PackContentWithItems, error) {
	// First, check if the pack exists in the database
	packExists, err := checkPackExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !packExists {
		return nil, ErrPackNotFound
	}

	// Pack exists, continue with fetching its contents
	var packWithItems PackContentWithItems

	rows, err := database.DB().QueryContext(ctx,
		`SELECT pc.id AS pack_content_id,
			pc.pack_id as pack_id,
			i.id AS inventory_id,
			i.item_name,
			i.category,
			i.description AS item_description,
			i.weight,
			i.url AS item_url,
			i.price,
			i.currency,
			pc.quantity,
			pc.worn,
			pc.consumable
			FROM pack_content pc JOIN inventory i ON pc.item_id = i.id
			WHERE pc.pack_id = $1;`,
		id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item PackContentWithItem
		err := rows.Scan(
			&item.PackContentID,
			&item.PackID,
			&item.InventoryID,
			&item.ItemName,
			&item.Category,
			&item.ItemDescription,
			&item.Weight,
			&item.ItemURL,
			&item.Price,
			&item.Currency,
			&item.Quantity,
			&item.Worn,
			&item.Consumable)
		if err != nil {
			return nil, err
		}
		packWithItems = append(packWithItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &packWithItems, nil
}

// Pack content writes

func insertPackContent(ctx context.Context, pc *PackContent) error {
	if pc == nil {
		return errors.New("payload is empty")
	}
	pc.CreatedAt = time.Now().Truncate(time.Second)
	pc.UpdatedAt = time.Now().Truncate(time.Second)

	//nolint:execinquery
	err := database.DB().QueryRowContext(ctx, `
		INSERT INTO pack_content
		(pack_id, item_id, quantity, worn, consumable, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id;`,
		pc.PackID,
		pc.ItemID,
		pc.Quantity,
		pc.Worn,
		pc.Consumable,
		pc.CreatedAt,
		pc.UpdatedAt).Scan(&pc.ID)

	if err != nil {
		return err
	}
	return nil
}

func updatePackContentByID(ctx context.Context, id uint, pc *PackContent) error {
	if pc == nil {
		return errors.New("payload is empty")
	}
	pc.UpdatedAt = time.Now().Truncate(time.Second)

	statement, err := database.DB().PrepareContext(ctx, `
		UPDATE pack_content
		SET pack_id=$1, item_id=$2, quantity=$3, worn=$4, consumable=$5, updated_at=$6
		WHERE id=$7;`)
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.ExecContext(ctx,
		pc.PackID, pc.ItemID, pc.Quantity, pc.Worn, pc.Consumable, pc.UpdatedAt, id)
	if err != nil {
		return err
	}
	return nil
}

func deletePackContentByID(ctx context.Context, id uint) error {
	statement, err := database.DB().PrepareContext(ctx, "DELETE FROM pack_content WHERE id=$1;")
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

// Import/CSV

// Take a record from csv.Newreder and return a LighterPackItem
func readLineFromCSV(ctx context.Context, record []string) (LighterPackItem, error) {
	var lighterPackItem LighterPackItem

	lighterPackItem.ItemName = record[0]
	lighterPackItem.Category = record[1]
	lighterPackItem.Desc = record[2]
	lighterPackItem.Unit = record[5]
	lighterPackItem.URL = record[6]

	qty, err := strconv.Atoi(record[3])
	if err != nil {
		return lighterPackItem, errors.New("invalid CSV format - failed to convert qty to number")
	}

	lighterPackItem.Qty = qty

	weight, err := strconv.Atoi(record[4])
	if err != nil {
		return lighterPackItem, errors.New("invalid CSV format - failed to convert weight to number")
	}

	lighterPackItem.Weight = weight

	price, err := strconv.Atoi(record[7])
	if err != nil {
		return lighterPackItem, errors.New("invalid CSV format - failed to convert price to number")
	}

	lighterPackItem.Price = price

	if record[8] == "Worn" {
		lighterPackItem.Worn = true
	}

	if record[9] == "Consumable" {
		lighterPackItem.Consumable = true
	}

	return lighterPackItem, nil
}

func insertLighterPack(ctx context.Context, lp *LighterPack, userID uint) (uint, error) {
	if lp == nil || len(*lp) == 0 {
		return 0, errors.New("payload is empty")
	}

	// Create new pack
	var newPack Pack
	newPack.UserID = userID
	newPack.PackName = "LighterPack Import"
	newPack.PackDescription = "LighterPack Import"
	err := insertPack(ctx, &newPack)
	if err != nil {
		return 0, err
	}

	// Insert content in new pack with insertPackContent
	for _, item := range *lp {
		var itemID uint

		// Check if item already exists in inventory
		existingItem, err := inventories.FindInventoryItemByAttributes(
			ctx,
			userID,
			item.ItemName,
			item.Category,
			item.Desc,
		)
		if err != nil && !errors.Is(err, inventories.ErrNoItemFound) {
			return 0, fmt.Errorf("failed to check for existing item: %w", err)
		}

		if existingItem != nil {
			// Item exists, reuse it
			itemID = existingItem.ID
		} else {
			// Item doesn't exist, create it
			var i inventories.Inventory
			i.UserID = userID
			i.ItemName = item.ItemName
			i.Category = item.Category
			i.Description = item.Desc
			i.Weight = item.Weight
			i.URL = item.URL
			i.Price = item.Price
			i.Currency = "USD"
			err := inventories.InsertInventory(ctx, &i)
			if err != nil {
				return 0, err
			}
			itemID = i.ID
		}

		// Create PackContent with the item (existing or new)
		var pc PackContent
		pc.PackID = newPack.ID
		pc.ItemID = itemID
		pc.Quantity = item.Qty
		pc.Worn = item.Worn
		pc.Consumable = item.Consumable
		err = insertPackContent(ctx, &pc)
		if err != nil {
			return 0, err
		}
	}
	return newPack.ID, nil
}

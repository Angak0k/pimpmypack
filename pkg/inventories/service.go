package inventories

import (
	"context"
	"time"
)

// FindInventoryItemByAttributes finds an existing inventory item for a user
// by exact match on item_name, category, and description.
// Returns the item if found, nil with ErrNoItemFound if not found, or error if query fails.
func FindInventoryItemByAttributes(
	ctx context.Context,
	userID uint,
	itemName, category, description string,
) (*Inventory, error) {
	return findInventoryItemByAttributes(ctx, userID, itemName, category, description)
}

// InsertInventory creates a new inventory item with timestamps
func InsertInventory(ctx context.Context, i *Inventory) error {
	i.CreatedAt = time.Now().Truncate(time.Second)
	i.UpdatedAt = time.Now().Truncate(time.Second)
	return insertInventory(ctx, i)
}

// FindItemIDByItemName finds an inventory item ID by its name
// Returns 0 if not found
func FindItemIDByItemName(inventories Inventories, itemname string) uint {
	for _, item := range inventories {
		if item.ItemName == itemname {
			return item.ID
		}
	}
	return 0
}

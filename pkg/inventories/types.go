package inventories

import "time"

// DefaultCurrency is the default currency used when none is specified
// This matches the database default value in the Currency enum
const DefaultCurrency = "EUR"

// Inventory represents an item in a user's inventory
type Inventory struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	ItemName    string    `json:"item_name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Weight      int       `json:"weight"`
	URL         string    `json:"url"`
	Price       int       `json:"price"`
	Currency    string    `json:"currency"`
	HasImage    bool      `json:"has_image"`
	PackCount   int       `json:"pack_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Inventories represents a collection of inventory items
type Inventories []Inventory

// InventoryCreateRequest represents the input for creating an inventory item (user endpoint)
type InventoryCreateRequest struct {
	ItemName    string `json:"item_name" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Description string `json:"description"`
	Weight      int    `json:"weight"`
	URL         string `json:"url"`
	Price       int    `json:"price"`
	Currency    string `json:"currency"`
}

// InventoryCreateAdminRequest represents the input for creating an inventory item (admin endpoint)
type InventoryCreateAdminRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	ItemName    string `json:"item_name" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Description string `json:"description"`
	Weight      int    `json:"weight"`
	URL         string `json:"url"`
	Price       int    `json:"price"`
	Currency    string `json:"currency"`
}

// InventoryUpdateRequest represents the input for updating an inventory item
type InventoryUpdateRequest struct {
	ItemName    string `json:"item_name" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Description string `json:"description"`
	Weight      int    `json:"weight"`
	URL         string `json:"url"`
	Price       int    `json:"price"`
	Currency    string `json:"currency"`
}

// MergeInventoryRequest represents the input for merging two inventory items
type MergeInventoryRequest struct {
	SourceItemID uint   `json:"source_item_id" binding:"required"`
	TargetItemID uint   `json:"target_item_id" binding:"required"`
	ItemName     string `json:"item_name" binding:"required"`
	Category     string `json:"category" binding:"required"`
	Description  string `json:"description"`
	Weight       int    `json:"weight"`
	URL          string `json:"url"`
	Price        int    `json:"price"`
	Currency     string `json:"currency"`
	ImageSource  string `json:"image_source" binding:"required,oneof=source target none"`
}

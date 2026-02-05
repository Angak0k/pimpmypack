package inventories

import "time"

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

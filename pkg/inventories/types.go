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

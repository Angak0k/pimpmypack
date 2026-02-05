package packs

import (
	"errors"
	"time"
)

// Domain errors
var (
	// ErrPackNotFound is returned when a pack is not found
	ErrPackNotFound = errors.New("pack not found")

	// ErrPackContentNotFound is returned when no items are found in a given pack
	ErrPackContentNotFound = errors.New("pack content not found")
)

// Pack represents a pack with its metadata
type Pack struct {
	ID              uint      `json:"id"`
	UserID          uint      `json:"user_id"`
	PackName        string    `json:"pack_name"`
	PackDescription string    `json:"pack_description"`
	PackWeight      int       `json:"pack_weight"`
	PackItemsCount  int       `json:"pack_items_count"`
	SharingCode     *string   `json:"sharing_code,omitempty"`
	HasImage        bool      `json:"has_image"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Packs represents a collection of packs
type Packs []Pack

// PackCreateRequest represents the input for creating a pack (user endpoint)
type PackCreateRequest struct {
	PackName        string `json:"pack_name" binding:"required"`
	PackDescription string `json:"pack_description"`
}

// PackCreateAdminRequest represents the input for creating a pack (admin endpoint)
type PackCreateAdminRequest struct {
	UserID          uint   `json:"user_id" binding:"required"`
	PackName        string `json:"pack_name" binding:"required"`
	PackDescription string `json:"pack_description"`
}

// PackUpdateRequest represents the input for updating a pack
type PackUpdateRequest struct {
	PackName        string `json:"pack_name" binding:"required"`
	PackDescription string `json:"pack_description"`
}

// PackContent represents an item in a pack
type PackContent struct {
	ID         uint      `json:"id"`
	PackID     uint      `json:"pack_id"`
	ItemID     uint      `json:"item_id"`
	Quantity   int       `json:"quantity"`
	Worn       bool      `json:"worn"`
	Consumable bool      `json:"consumable"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PackContents represents a collection of pack contents
type PackContents []PackContent

// PackContentWithItem represents a pack content with inventory item details
type PackContentWithItem struct {
	PackContentID   uint   `json:"pack_content_id"`
	PackID          uint   `json:"pack_id"`
	InventoryID     uint   `json:"inventory_id"`
	ItemName        string `json:"item_name"`
	Category        string `json:"category"`
	ItemDescription string `json:"item_description"`
	Weight          int    `json:"weight"`
	ItemURL         string `json:"item_url"`
	Price           int    `json:"price"`
	Currency        string `json:"currency"`
	Quantity        int    `json:"quantity"`
	Worn            bool   `json:"worn"`
	Consumable      bool   `json:"consumable"`
}

// PackContentWithItems represents a collection of pack contents with item details
type PackContentWithItems []PackContentWithItem

// PackContentRequest represents the data required to add an item to a pack
type PackContentRequest struct {
	InventoryID uint `json:"inventory_id" binding:"required"`
	Quantity    int  `json:"quantity" binding:"required,min=1"`
	Worn        bool `json:"worn"`
	Consumable  bool `json:"consumable"`
}

// PackContentCreateRequest represents the input for creating pack content (admin)
// Note: User endpoint uses PackContentRequest which doesn't include PackID/ItemID
type PackContentCreateRequest struct {
	PackID     uint `json:"pack_id" binding:"required"`
	ItemID     uint `json:"item_id" binding:"required"`
	Quantity   int  `json:"quantity" binding:"required,min=1"`
	Worn       bool `json:"worn"`
	Consumable bool `json:"consumable"`
}

// PackContentUpdateRequest represents the input for updating pack content
type PackContentUpdateRequest struct {
	PackID     uint `json:"pack_id" binding:"required"`
	ItemID     uint `json:"item_id" binding:"required"`
	Quantity   int  `json:"quantity" binding:"required,min=1"`
	Worn       bool `json:"worn"`
	Consumable bool `json:"consumable"`
}

// LighterPackItem represents an item imported from LighterPack format
type LighterPackItem struct {
	ItemName   string `json:"item_name"`
	Category   string `json:"category"`
	Desc       string `json:"desc"`
	Qty        int    `json:"qty"`
	Weight     int    `json:"weight"`
	Unit       string `json:"unit"`
	URL        string `json:"url"`
	Price      int    `json:"price"`
	Worn       bool   `json:"worn"`
	Consumable bool   `json:"consumable"`
}

// LighterPack represents a collection of LighterPack items
type LighterPack []LighterPackItem

// ImportLighterPackResponse represents the response when importing from LighterPack
type ImportLighterPackResponse struct {
	Message string `json:"message"`
	PackID  uint   `json:"pack_id"`
}

// SharedPackResponse represents the response structure for shared pack endpoint
type SharedPackResponse struct {
	Pack     SharedPackInfo       `json:"pack"`
	Contents PackContentWithItems `json:"contents"`
}

// SharedPackInfo contains public metadata about a shared pack
// UserID and SharingCode are intentionally not included for security
// Note: pack_items_count is not included as it doesn't exist in DB schema
// Clients can count items from the contents array
type SharedPackInfo struct {
	ID              uint      `json:"id"`
	PackName        string    `json:"pack_name"`
	PackDescription string    `json:"pack_description"`
	HasImage        bool      `json:"has_image"`
	CreatedAt       time.Time `json:"created_at"`
}

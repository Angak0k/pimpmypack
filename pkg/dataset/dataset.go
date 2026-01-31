package dataset

import (
	"time"
)

type Account struct {
	ID                  uint      `json:"id"`
	Username            string    `json:"username"`
	Email               string    `json:"email"`
	Firstname           string    `json:"firstname"`
	Lastname            string    `json:"lastname"`
	Role                string    `json:"role"`
	Status              string    `json:"status"`
	PreferredCurrency   string    `json:"preferred_currency"`
	PreferredUnitSystem string    `json:"preferred_unit_system"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Accounts []Account

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

type Inventories []Inventory

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

type Packs []Pack

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

type PackContents []PackContent

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

type PackContentWithItems []PackContentWithItem

type PackContentRequest struct {
	InventoryID uint `json:"inventory_id"`
	Quantity    int  `json:"quantity"`
	Worn        bool `json:"worn"`
	Consumable  bool `json:"consumable"`
}

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

type LighterPack []LighterPackItem

type RegisterInput struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Firstname string `json:"firstname" binding:"required"`
	Lastname  string `json:"lastname" binding:"required"`
}

type User struct {
	ID                  uint      `json:"id"`
	Username            string    `json:"username"`
	Email               string    `json:"email"`
	Firstname           string    `json:"firstname"`
	Lastname            string    `json:"lastname"`
	Role                string    `json:"role"`
	Status              string    `json:"status"`
	Password            string    `json:"password"`
	LastPassword        string    `json:"last_password"`
	PreferredCurrency   string    `json:"preferred_currency"`
	PreferredUnitSystem string    `json:"preferred_unit_system"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type MailServer struct {
	MailServer   string `json:"mail_server"`
	MailPort     int    `json:"mail_port"`
	MailIdentity string `json:"mail_identity"`
	MailUsername string `json:"mail_username"`
	MailPassword string `json:"mail_password"`
}

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required"`
}

type PasswordUpdateInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

type Token struct {
	Token string `json:"token"`
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

type OkResponse struct {
	Response string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

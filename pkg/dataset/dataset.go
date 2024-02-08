package dataset

import (
	"time"
)

type Account struct {
	ID         uint      `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Firstname  string    `json:"firstname"`
	Lastname   string    `json:"lastname"`
	Role       string    `json:"role"`
	Status     string    `json:"status"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

type Accounts []Account

type Inventory struct {
	ID          uint      `json:"id"`
	User_id     uint      `json:"user_id"`
	Item_name   string    `json:"item_name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Weight      int       `json:"weight"`
	Weight_unit string    `json:"weight_unit"`
	Url         string    `json:"url"`
	Price       int       `json:"price"`
	Currency    string    `json:"currency"`
	Created_at  time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
}

type Inventories []Inventory

type Pack struct {
	ID               uint      `json:"id"`
	User_id          uint      `json:"user_id"`
	Pack_name        string    `json:"pack_name"`
	Pack_description string    `json:"pack_description"`
	Sharing_code     string    `json:"sharing_code"`
	Created_at       time.Time `json:"created_at"`
	Updated_at       time.Time `json:"updated_at"`
}

type Packs []Pack

type PackContent struct {
	ID         uint      `json:"id"`
	Pack_id    uint      `json:"pack_id"`
	Item_id    uint      `json:"item_id"`
	Quantity   int       `json:"quantity"`
	Worn       bool      `json:"worn"`
	Consumable bool      `json:"consumable"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

type PackContents []PackContent

type PackContentWithItem struct {
	Pack_content_id  uint   `json:"pack_content_id"`
	Pack_id          uint   `json:"pack_id"`
	Inventory_id     uint   `json:"inventory_id"`
	Item_name        string `json:"item_name"`
	Category         string `json:"category"`
	Item_description string `json:"item_description"`
	Weight           int    `json:"weight"`
	Weight_unit      string `json:"weight_unit"`
	Item_url         string `json:"item_url"`
	Price            int    `json:"price"`
	Currency         string `json:"currency"`
	Quantity         int    `json:"quantity"`
	Worn             bool   `json:"worn"`
	Consumable       bool   `json:"consumable"`
}

type PackContentWithItems []PackContentWithItem

type LighterPackItem struct {
	Item_name  string `json:"item_name"`
	Category   string `json:"category"`
	Desc       string `json:"desc"`
	Qty        int    `json:"qty"`
	Weight     int    `json:"weight"`
	Unit       string `json:"unit"`
	Url        string `json:"url"`
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
	ID           uint      `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Firstname    string    `json:"firstname"`
	Lastname     string    `json:"lastname"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	Password     string    `json:"password"`
	LastPassword string    `json:"last_password"`
	Created_at   time.Time `json:"created_at"`
	Updated_at   time.Time `json:"updated_at"`
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
	Password string `json:"password" binding:"required"`
}

type Token struct {
	Token string `json:"token"`
}

type OkResponse struct {
	Response string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

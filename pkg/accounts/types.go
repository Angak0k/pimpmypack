package accounts

import "time"

// Account represents a user account with public information
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

// Accounts represents a collection of accounts
type Accounts []Account

// User represents a user with authentication information
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

// RegisterInput represents the data required to register a new account
type RegisterInput struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Firstname string `json:"firstname" binding:"required"`
	Lastname  string `json:"lastname" binding:"required"`
}

// LoginInput represents the data required to login
type LoginInput struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"` // optional, defaults to false
}

// ForgotPasswordInput represents the data required to reset a password
type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required"`
}

// PasswordUpdateInput represents the data required to update a password
type PasswordUpdateInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// AccountUpdateInput represents the data users can update about their own account
// SECURITY: This type intentionally excludes role, status, username, and password fields
// to prevent privilege escalation and unauthorized modifications
type AccountUpdateInput struct {
	Email               string `json:"email" binding:"required"`
	Firstname           string `json:"firstname" binding:"required"`
	Lastname            string `json:"lastname" binding:"required"`
	PreferredCurrency   string `json:"preferred_currency"`
	PreferredUnitSystem string `json:"preferred_unit_system"`
}

// Token represents an authentication token
type Token struct {
	Token string `json:"token"`
}

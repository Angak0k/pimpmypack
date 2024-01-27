package accounts

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/dataset"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Register a new user account
// @Summary Register new user
// @Description Register a new user with username, password, email, firstname, and lastname
// @Tags Public
// @Accept  json
// @Produce  json
// @Param   input  body    dataset.RegisterInput  true  "Register Info"
// @Success 200 {object} dataset.RegisterResponse
// @Failure 400 {object} dataset.ErrorResponse
// @Router /register [post]
func Register(c *gin.Context) {

	var input dataset.RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := dataset.User{}

	user.Username = input.Username
	user.Password = input.Password
	user.Email = input.Email
	user.Firstname = input.Firstname
	user.Lastname = input.Lastname
	user.Role = "standard"
	user.Status = "pending"
	user.Created_at = time.Now().Truncate(time.Second)
	user.Updated_at = time.Now().Truncate(time.Second)

	err := saveUser(user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "registration success"})

}

func saveUser(u dataset.User) error {
	var id int

	err := database.Db().QueryRow("INSERT INTO account (username, email, firstname, lastname, role, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id;", u.Username, u.Email, u.Firstname, u.Lastname, u.Role, u.Status, u.Created_at, u.Updated_at).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	hashedPassword, err := security.HashPassword(u.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = database.Db().QueryRow("INSERT INTO password (user_id, password, updated_at) VALUES ($1,$2,$3) RETURNING id;", id, hashedPassword, u.Updated_at).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert password: %w", err)
	}

	return nil
}

// Update user password
// @Summary Update password
// @Description Update the password of the current logged-in user
// @Security Bearer
// @Tags Accounts
// @Accept  json
// @Produce  json
// @Param   password  body    string  true  "Updated Password"
// @Success 200 {string} string "Updated password successfully"
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /mypassword [put]
func PutMyPassword(c *gin.Context) {

	var updatedPassword string

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call BindJSON to bind the received JSON to updatedPassword.
	if err := c.BindJSON(&updatedPassword); err != nil {
		return
	}

	// Update the DB
	err = updatePassword(user_id, updatedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, updatedPassword)
}

func updatePassword(user_id uint, updatedPassword string) error {
	var lastPassword string
	// Get old password
	row := database.Db().QueryRow("SELECT password FROM password WHERE user_id = $1);", user_id)
	err := row.Scan(&lastPassword)
	if err != nil {
		return fmt.Errorf("failed to get old password: %w", err)
	}

	// Update DB
	statement, err := database.Db().Prepare("UPDATE password SET password = $1, last_password = $2, updated_at = $3 WHERE user_id = $4);")
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer statement.Close()

	hashedPassword, err := security.HashPassword(updatedPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = statement.Exec(hashedPassword, lastPassword, time.Now().Truncate(time.Second), user_id)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	return nil
}

// User login
// @Summary User login
// @Description Logs in a user by providing username and password
// @Tags Public
// @Accept  json
// @Produce  json
// @Param   username  body    string  true  "Username"
// @Param   password  body    string  true  "Password"
// @Success 200 {object} map[string]string "token"
// @Failure 403 {object} map[string]interface{} "error: credentials are incorrect or token generation failed"
// @Router /login [post]
func Login(c *gin.Context) {

	var input dataset.LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := loginCheck(input.Username, input.Password)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "credentials are incorrect or token generation failed."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})

}

func loginCheck(username string, password string) (string, error) {

	var err error
	var storedPassword string
	var id uint

	row := database.Db().QueryRow("SELECT password, user_id FROM password WHERE user_id = (SELECT id FROM account WHERE username = $1);", username)
	err = row.Scan(&storedPassword, &id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	err = security.VerifyPassword(password, storedPassword)

	if err != nil && errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return "", err
	}

	token, err := security.GenerateToken(id)

	if err != nil {
		return "", err
	}

	return token, nil

}

// Get my account information
// @Summary Get account info
// @Description Get information of the currently logged-in user
// @Security Bearer
// @Tags Accounts
// @Produce  json
// @Success 200 {object} dataset.Account "Account Information"
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 404 {object} map[string]interface{} "error: Account not found"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /myaccount [get]
func GetMyAccount(c *gin.Context) {

	user_id, err := security.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := findAccountById(user_id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if account != nil {
		c.IndentedJSON(http.StatusOK, *account)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Account not found"})
	}

}

// Update my account information
// @Summary Update account info
// @Description Update information of the currently logged-in user
// @Security Bearer
// @Tags Accounts
// @Accept  json
// @Produce  json
// @Param   input  body    dataset.Account  true  "Account Information"
// @Success 200 {object} dataset.Account
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /myaccount [put]
func PutMyAccount(c *gin.Context) {

	var updatedAccount dataset.Account

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call BindJSON to bind the received JSON to updatedAccount.
	if err := c.BindJSON(&updatedAccount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the DB
	err = updateAccountById(user_id, &updatedAccount)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, updatedAccount)
}

// Get all accounts
// @Summary [ADMIN] Get all accounts
// @Description Get all accounts - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Success 200 {object} dataset.Account
// @Failure 500 {object} map[string]interface{} "error"
// @Router /accounts [get]
func GetAccounts(c *gin.Context) {
	accounts, err := returnAccounts()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, *accounts)

}

func returnAccounts() (*dataset.Accounts, error) {
	var accounts dataset.Accounts

	rows, err := database.Db().Query("SELECT id, username, email, firstname, lastname, role, status, created_at, updated_at FROM account;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var account dataset.Account
		err := rows.Scan(&account.ID, &account.Username, &account.Email, &account.Firstname, &account.Lastname, &account.Role, &account.Status, &account.Created_at, &account.Updated_at)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &accounts, nil
}

// Get account by ID
// @Summary [ADMIN] Get account by ID
// @Description Get account by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param   id  path    int  true  "Account ID"
// @Success 200 {object} dataset.Account
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 404 {object} map[string]interface{} "error: Account not found"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /accounts/{id} [get]
func GetAccountByID(c *gin.Context) {

	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	id := uint(id64)

	// Call findAccountById function to lookup in database
	account, err := findAccountById(id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if account != nil {
		c.IndentedJSON(http.StatusOK, *account) // Dereference only if account is not nil
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Account not found"})
	}
}

func findAccountById(id uint) (*dataset.Account, error) {
	var account dataset.Account

	row := database.Db().QueryRow("SELECT id, username, email, firstname, lastname, role, status, created_at, updated_at FROM account WHERE id = $1;", id)
	err := row.Scan(&account.ID, &account.Username, &account.Email, &account.Firstname, &account.Lastname, &account.Role, &account.Status, &account.Created_at, &account.Updated_at)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Handle case when no rows are returned
			return nil, nil
		}
		return nil, err
	}

	return &account, nil
}

// Create a new account
// @Summary [ADMIN] Create a new account
// @Description Create a new account - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param   input  body    dataset.Account  true  "Account Information"
// @Success 201 {object} dataset.Account
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /accounts [post]
func PostAccount(c *gin.Context) {
	var newAccount dataset.Account

	// Call BindJSON to bind the received JSON to
	// newAccount.
	if err := c.BindJSON(&newAccount); err != nil {
		return
	}

	// Insert the new account into the database.
	err := insertAccount(&newAccount)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newAccount)

}

func insertAccount(a *dataset.Account) error {
	if a == nil {
		return errors.New("payload is empty")
	}
	a.Created_at = time.Now().Truncate(time.Second)
	a.Updated_at = time.Now().Truncate(time.Second)

	err := database.Db().QueryRow("INSERT INTO account (username, email, firstname, lastname, role, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id;", a.Username, a.Email, a.Firstname, a.Lastname, a.Role, a.Status, a.Created_at, a.Updated_at).Scan(&a.ID)

	if err != nil {
		return err
	}
	return nil
}

// Update account by ID
// @Summary [ADMIN] Update account by ID
// @Description Update account by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param   id  path    int  true  "Account ID"
// @Param   input  body    dataset.Account  true  "Account Information"
// @Success 200 {object} dataset.Account
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /accounts/{id} [put]
func PutAccountByID(c *gin.Context) {
	var updatedAccount dataset.Account

	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	id := uint(id64)

	// Call BindJSON to bind the received JSON to updatedAccount.
	if err := c.BindJSON(&updatedAccount); err != nil {
		return
	}
	// Update the DB
	err = updateAccountById(id, &updatedAccount)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, updatedAccount)
}

func updateAccountById(id uint, a *dataset.Account) error {
	if a == nil {
		return errors.New("payload is empty")
	}

	a.ID = id
	a.Updated_at = time.Now().Truncate(time.Second)
	statement, err := database.Db().Prepare("UPDATE account SET email=$1, firstname=$2, lastname=$3, status=$4, role=$5, updated_at=$6 WHERE id=$7 RETURNING username;")
	if err != nil {
		return err
	}
	err = statement.QueryRow(a.Email, a.Firstname, a.Lastname, a.Status, a.Role, a.Updated_at, a.ID).Scan(&a.Username)
	if err != nil {
		return err
	}
	return nil
}

// Delete account by ID
// @Summary [ADMIN] Delete account by ID
// @Description Delete account by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param   id  path    int  true  "Account ID"
// @Success 200 {string} string "Account deleted"
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /accounts/{id} [delete]
func DeleteAccountByID(c *gin.Context) {
	id := c.Param("id")
	err := deleteAccountById(id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Account deleted"})
}

func deleteAccountById(id string) error {
	statement, err := database.Db().Prepare("DELETE FROM account WHERE id=$1;")
	if err != nil {
		return err
	}
	_, err = statement.Exec(id)
	if err != nil {
		return err
	}

	return nil

}

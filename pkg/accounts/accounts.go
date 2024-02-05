package accounts

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/dataset"
	"github.com/Angak0k/pimpmypack/pkg/helper"
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
// @Param   input  body    dataset.RegisterInput  true  "Register Informations"
// @Success 200 {object} dataset.OkResponse
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

	err, emailSended := registerUser(user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !emailSended {
		c.JSON(http.StatusAccepted, gin.H{"message": "registration succeed but failed to send confirmation email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "registration succeed, please check your email to confirm your account"})

}

func registerUser(u dataset.User) (error, bool) {
	var id int

	confirmationCode, err := helper.GenerateRandomCode(30)
	if err != nil {
		return fmt.Errorf("failed to generate confirmation code: %w", err), false
	}

	err = database.Db().QueryRow("INSERT INTO account (username, email, firstname, lastname, role, status, confirmation_code, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8, $9) RETURNING id;", u.Username, u.Email, u.Firstname, u.Lastname, u.Role, u.Status, confirmationCode, u.Created_at, u.Updated_at).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err), false
	}

	u.ID = uint(id)

	hashedPassword, err := security.HashPassword(u.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err), false
	}

	err = database.Db().QueryRow("INSERT INTO password (user_id, password, updated_at) VALUES ($1,$2,$3) RETURNING id;", id, hashedPassword, u.Updated_at).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert password: %w", err), false
	}

	err = sendConfirmationEmail(u, confirmationCode)
	if err != nil {
		// we haven't succed to send the email but the user is created
		return nil, false
	}

	return nil, true
}

// Send confirmation email
func sendConfirmationEmail(u dataset.User, code string) error {
	// Send confirmation email
	mailRcpt := u.Email
	mailSubject := "PimpMyPack - Confirm your email address"
	mailBody := "Please confirm your email address by clicking on the following link: " + config.Scheme + "://" + config.HostName + "/api/confirmemail?id=" + strconv.Itoa(int(u.ID)) + "&code=" + code

	smtpClient := helper.SMTPClient{Server: config.MailServer}

	err := smtpClient.SendEmail(mailRcpt, mailSubject, mailBody)
	if err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}

// Confirm email address
// @Summary Confirm email address
// @Description Confirm email address by providing username and email
// @Tags Public
// @Produce  json
// @Success 200 {object} dataset.OkResponse
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /confirmemail [get]
func ConfirmEmail(c *gin.Context) {

	// Retrieve the confirmation code from the url query
	confirmationCode := c.Query("code")
	userID := c.Query("id")
	if confirmationCode == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid confirmation code or user ID"})
		return
	}

	err := confirmEmail(userID, confirmationCode)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email confirmed"})

}

func confirmEmail(id string, code string) error {

	// Check if the confirmation code is valid
	row := database.Db().QueryRow("SELECT id FROM account WHERE id = $1 AND confirmation_code = $2;", id, code)
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	// Update the DB
	statement, err := database.Db().Prepare("UPDATE account SET status = 'active' WHERE id = $1;")
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer statement.Close()

	_, err = statement.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	return nil
}

// Reset password
// @Summary Reset password
// @Description Send a new password to the user's email
// @Tags Public
// @Accept  json
// @Produce  json
// @Param   input  body    dataset.ForgotPasswordInput  true  "Email Address"
// @Success 200 {object} dataset.OkResponse
// @Failure 400 {object} dataset.ErrorResponse "Bad Request"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /forgotpassword [post]
func ForgotPassword(c *gin.Context) {

	var email dataset.ForgotPasswordInput

	if err := c.ShouldBindJSON(&email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := forgotPassword(email.Email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "new password sent"})

}

func forgotPassword(email string) error {

	newPassword, err := helper.GenerateRandomCode(10)
	if err != nil {
		return fmt.Errorf("failed to generate new password: %w", err)
	}

	user_id, err := getUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to process the request %w", err) // email not found but we don't want to leak this information
	}

	err = updatePassword(user_id, newPassword)
	if err != nil {
		return fmt.Errorf("failed to generate new password: %w", err)
	}

	mailRcpt := email
	mailSubject := "PimpMyPack - Your password has been reset"
	mailBody := "Hi! your password has been reset. If you did not request this, please contact us.\n\nYour new password is: " + newPassword

	smtpClient := helper.SMTPClient{Server: config.MailServer}

	err = smtpClient.SendEmail(mailRcpt, mailSubject, mailBody)
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}
	return nil
}

// User login
// @Summary User login
// @Description Log in a user by providing username and password
// @Tags Public
// @Produce  json
// @Param   input  body    dataset.LoginInput  true  "Credentials Info"
// @Success 200 {object} dataset.Token
// @Failure 401 {object} dataset.ErrorResponse
// @Router /login [post]
func Login(c *gin.Context) {

	var input dataset.LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, pending, err := loginCheck(input.Username, input.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "credentials are incorrect or token generation failed. "})
		return
	}

	if pending {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "account not yet confirmed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})

}

func loginCheck(username string, password string) (string, bool, error) {

	var err error
	var status string
	var storedPassword string
	var id uint

	row := database.Db().QueryRow("SELECT p.password, p.user_id, a.status FROM password AS p JOIN account AS a ON p.user_id = a.id WHERE a.username = $1;", username)
	err = row.Scan(&storedPassword, &id, &status)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}

	err = security.VerifyPassword(password, storedPassword)

	if err != nil && errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return "", false, err
	}

	token, err := security.GenerateToken(id)

	if err != nil {
		return "", false, err
	}

	if status != "active" {
		return token, true, nil
	}

	return token, false, nil

}

// Get my account information
// @Summary Get account info
// @Description Get information of the currently logged-in user
// @Security Bearer
// @Tags Accounts
// @Produce  json
// @Success 200 {object} dataset.Account "Account Information"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 404 {object} dataset.ErrorResponse "Account not found"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /v1/myaccount [get]
func GetMyAccount(c *gin.Context) {

	user_id, err := security.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
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

// Update user password
// @Summary Update password
// @Description Update the password of the current logged-in user
// @Security Bearer
// @Tags Accounts
// @Accept  json
// @Produce  json
// @Param   password  body dataset.PasswordUpdateInput  true  "New Password"
// @Success 200 {object} dataset.OkResponse "Password updated"
// @Failure 400 {object} dataset.ErrorResponse "Bad Request"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 500 {object} dataset.ErrorResponse "Internal Server Error"
// @Router /v1/mypassword [put]
func PutMyPassword(c *gin.Context) {

	var updatedPassword string

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Call BindJSON to bind the received JSON to updatedPassword.
	err = c.BindJSON(&updatedPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	row := database.Db().QueryRow("SELECT password FROM password WHERE user_id = $1;", user_id)
	err := row.Scan(&lastPassword)
	if err != nil {
		return fmt.Errorf("failed to get old password: %w", err)
	}

	// Update DB
	statement, err := database.Db().Prepare("UPDATE password SET password = $1, last_password = $2, updated_at = $3 WHERE user_id = $4;")
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

// Update my account information
// @Summary Update account info
// @Description Update information of the currently logged-in user
// @Security Bearer
// @Tags Accounts
// @Accept  json
// @Produce  json
// @Param   input  body    dataset.Account  true  "Account Information"
// @Success 200 {object} dataset.Account
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 401 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /v1/myaccount [put]
func PutMyAccount(c *gin.Context) {

	var updatedAccount dataset.Account

	user_id, err := security.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Call BindJSON to bind the received JSON to updatedAccount.
	if err := c.BindJSON(&updatedAccount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedAccount.ID = user_id

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
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/accounts [get]
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
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 404 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/accounts/{id} [get]
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
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/accounts [post]
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
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/accounts/{id} [put]
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
// @Success 200 {object} dataset.OkResponse
// @Failure 400 {object} dataset.ErrorResponse
// @Failure 500 {object} dataset.ErrorResponse
// @Router /admin/accounts/{id} [delete]
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

func getUserIDByEmail(email string) (uint, error) {
	var id uint
	row := database.Db().QueryRow("SELECT id FROM account WHERE email = $1;", email)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

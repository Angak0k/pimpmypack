package accounts

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
)

// Register a new user account
// @Summary Register new user
// @Description Register a new user with username, password, email, firstname, and lastname
// @Tags Public
// @Accept  json
// @Produce  json
// @Param   input  body    RegisterInput  true  "Register Informations"
// @Success 200 {object} apitypes.OkResponse
// @Failure 400 {object} apitypes.ErrorResponse
// @Router /register [post]
func Register(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		helper.LogAndSanitize(err, "register: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	user := User{}

	if helper.IsValidEmail(input.Email) {
		user.Email = input.Email
		user.Username = input.Username
		user.Password = input.Password
		user.Firstname = input.Firstname
		user.Lastname = input.Lastname
		user.Role = "standard"
		user.Status = "pending"
		user.CreatedAt = time.Now().Truncate(time.Second)
		user.UpdatedAt = time.Now().Truncate(time.Second)

		emailSended, err := registerUser(c.Request.Context(), user)

		if err != nil {
			helper.LogAndSanitize(err, "register: user registration failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}

		if !emailSended {
			c.JSON(http.StatusAccepted, gin.H{"message": "registration succeed but failed to send confirmation email"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "registration succeed, please check your email to confirm your account"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}
}

// Confirm email address
// @Summary Confirm email address
// @Description Confirm email address by providing username and email
// @Tags Public
// @Produce  json
// @Success 200 {object} apitypes.OkResponse
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /confirmemail [get]
func ConfirmEmail(c *gin.Context) {
	// Retrieve the confirmation code from the url query
	confirmationCode := c.Query("code")
	userID := c.Query("id")

	// LOCAL mode: Support simplified confirmation by username + email
	if config.Stage == "LOCAL" {
		username := c.Query("username")
		email := c.Query("email")

		// Method 1: Simplified confirmation with username + email (no code needed)
		if username != "" && email != "" {
			err := confirmUserByUsernameAndEmail(c.Request.Context(), username, email)
			if err != nil {
				helper.LogAndSanitize(err, "confirm email by username/email failed")
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Email confirmed successfully (LOCAL mode)"})
			return
		}

		// Method 2: Traditional with ID but accept any code
		if userID != "" && confirmationCode == "" {
			confirmationCode = "LOCAL_BYPASS"
		}
	}

	// Validate required params for traditional confirmation
	if confirmationCode == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid confirmation code or user ID"})
		return
	}

	err := confirmEmail(c.Request.Context(), userID, confirmationCode)

	if err != nil {
		helper.LogAndSanitize(err, "confirm email failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email confirmed"})
}

// Reset password
// @Summary Reset password
// @Description Send a new password to the user's email
// @Tags Public
// @Accept  json
// @Produce  json
// @Param   input  body    ForgotPasswordInput  true  "Email Address"
// @Success 200 {object} apitypes.OkResponse
// @Failure 400 {object} apitypes.ErrorResponse "Bad Request"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /forgotpassword [post]
func ForgotPassword(c *gin.Context) {
	var email ForgotPasswordInput

	if err := c.ShouldBindJSON(&email); err != nil {
		helper.LogAndSanitize(err, "forgot password: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	err := forgotPassword(c.Request.Context(), email.Email)

	if err != nil {
		helper.LogAndSanitize(err, "forgot password failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "new password sent"})
}

// User login
// @Summary User login
// @Description Log in a user by providing username and password
// @Description Returns access token (short-lived) and refresh token (long-lived)
// @Description Use remember_me to extend refresh token lifetime
// @Tags Public
// @Accept  json
// @Produce  json
// @Param   input  body    LoginInput  true  "Credentials Info"
// @Success 200 {object} security.TokenPairResponse "Login successful with access and refresh tokens"
// @Failure 400 {object} apitypes.ErrorResponse "Bad request"
// @Failure 401 {object} apitypes.ErrorResponse "Invalid credentials or account not confirmed"
// @Failure 500 {object} apitypes.ErrorResponse "Internal server error"
// @Router /login [post]
func Login(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		helper.LogAndSanitize(err, "login: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	tokenPair, userID, pending, err := loginCheck(c.Request.Context(), input.Username, input.Password, input.RememberMe)

	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			security.AuditLoginFailed(c, input.Username, "invalid credentials")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "credentials are incorrect"})
			return
		}
		security.AuditLoginFailed(c, input.Username, "authentication failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}

	if pending {
		security.AuditLoginFailed(c, input.Username, "account not yet confirmed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "account not yet confirmed"})
		return
	}

	security.AuditLoginSuccess(c, userID, input.RememberMe)
	c.JSON(http.StatusOK, tokenPair)
}

// Get my account information
// @Summary Get account info
// @Description Get information of the currently logged-in user
// @Security Bearer
// @Tags Accounts
// @Produce  json
// @Success 200 {object} Account "Account Information"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 404 {object} apitypes.ErrorResponse "Account not found"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/myaccount [get]
func GetMyAccount(c *gin.Context) {
	userID, err := security.ExtractTokenID(c)

	if err != nil {
		helper.LogAndSanitize(err, "get my account: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	account, err := findAccountByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNoAccountFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		helper.LogAndSanitize(err, "get my account: find account failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	if account != nil {
		c.IndentedJSON(http.StatusOK, *account)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Account object is null"})
	}
}

// Update user password
// @Summary Update password
// @Description Update the password of the current logged-in user
// @Security Bearer
// @Tags Accounts
// @Accept  json
// @Produce  json
// @Param   password  body PasswordUpdateInput  true  "Current and New Password"
// @Success 200 {object} apitypes.OkResponse "Password updated"
// @Failure 400 {object} apitypes.ErrorResponse "Bad Request"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /v1/mypassword [put]
func PutMyPassword(c *gin.Context) {
	var input PasswordUpdateInput

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "put my password: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		helper.LogAndSanitize(err, "put my password: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	// Get the current hashed password from DB
	var storedPassword string
	row := database.DB().QueryRowContext(c.Request.Context(), "SELECT password FROM password WHERE user_id = $1;", userID)
	err = row.Scan(&storedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current password"})
		return
	}

	// Verify the current password
	if err := security.VerifyPassword(input.CurrentPassword, storedPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current password is incorrect"})
		return
	}

	// Check if new password is the same as current password
	if input.CurrentPassword == input.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be different from current password"})
		return
	}

	// Update the password
	if err := updatePassword(c.Request.Context(), userID, input.NewPassword); err != nil {
		helper.LogAndSanitize(err, "put my password: update password failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated"})
}

// Update my account information
// @Summary Update account info
// @Description Update information of the currently logged-in user
// @Security Bearer
// @Tags Accounts
// @Accept  json
// @Produce  json
// @Param   input  body    Account  true  "Account Information"
// @Success 200 {object} Account
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 401 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /v1/myaccount [put]
func PutMyAccount(c *gin.Context) {
	var updatedAccount Account

	userID, err := security.ExtractTokenID(c)
	if err != nil {
		helper.LogAndSanitize(err, "put my account: extract token ID failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": helper.ErrMsgUnauthorized})
		return
	}

	// Call BindJSON to bind the received JSON to updatedAccount.
	if err := c.BindJSON(&updatedAccount); err != nil {
		helper.LogAndSanitize(err, "put my account: bind JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": helper.ErrMsgBadRequest})
		return
	}

	updatedAccount.ID = userID

	// Update the DB
	err = updateAccountByID(c.Request.Context(), userID, &updatedAccount)
	if err != nil {
		helper.LogAndSanitize(err, "put my account: update account failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
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
// @Success 200 {object} Account
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/accounts [get]
func GetAccounts(c *gin.Context) {
	accounts, err := returnAccounts(c.Request.Context())
	if err != nil {
		helper.LogAndSanitize(err, "get accounts: return accounts failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	c.IndentedJSON(http.StatusOK, *accounts)
}

// Get account by ID
// @Summary [ADMIN] Get account by ID
// @Description Get account by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param   id  path    int  true  "Account ID"
// @Success 200 {object} Account
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 404 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/accounts/{id} [get]
func GetAccountByID(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	id := uint(id64)

	// Call findAccountById function to lookup in database
	account, err := findAccountByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNoAccountFound) {
			// Handle the "not found" case specifically
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			// Handle other errors
			helper.LogAndSanitize(err, "get account by ID: find account failed")
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		}
	}

	if account != nil {
		c.IndentedJSON(http.StatusOK, *account) // Dereference only if account is not nil
	} else {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Account object is null"})
	}
}

// Create a new account
// @Summary [ADMIN] Create a new account
// @Description Create a new account - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param   input  body    Account  true  "Account Information"
// @Success 201 {object} Account
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/accounts [post]
func PostAccount(c *gin.Context) {
	var newAccount Account

	// Call BindJSON to bind the received JSON to
	// newAccount.
	if err := c.BindJSON(&newAccount); err != nil {
		return
	}

	if helper.IsValidEmail(newAccount.Email) {
		// Insert the new account into the database.
		err := insertAccount(c.Request.Context(), &newAccount)
		if err != nil {
			helper.LogAndSanitize(err, "post account: insert account failed")
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
			return
		}

		c.IndentedJSON(http.StatusCreated, newAccount)
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}
}

// Update account by ID
// @Summary [ADMIN] Update account by ID
// @Description Update account by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Accept  json
// @Produce  json
// @Param   id  path    int  true  "Account ID"
// @Param   input  body    Account  true  "Account Information"
// @Success 200 {object} Account
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/accounts/{id} [put]
func PutAccountByID(c *gin.Context) {
	var updatedAccount Account

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
	err = updateAccountByID(c.Request.Context(), id, &updatedAccount)
	if err != nil {
		helper.LogAndSanitize(err, "put account by ID: update account failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}
	c.IndentedJSON(http.StatusOK, updatedAccount)
}

// Delete account by ID
// @Summary [ADMIN] Delete account by ID
// @Description Delete account by ID - for admin use only
// @Security Bearer
// @Tags Internal
// @Produce  json
// @Param   id  path    int  true  "Account ID"
// @Success 200 {object} apitypes.OkResponse
// @Failure 400 {object} apitypes.ErrorResponse
// @Failure 500 {object} apitypes.ErrorResponse
// @Router /admin/accounts/{id} [delete]
func DeleteAccountByID(c *gin.Context) {
	id := c.Param("id")
	err := deleteAccountByID(c.Request.Context(), id)
	if err != nil {
		helper.LogAndSanitize(err, "delete account by ID: delete account failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Account deleted"})
}

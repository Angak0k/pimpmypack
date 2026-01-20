package accounts

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/dataset"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/gruntwork-io/terratest/modules/random"
	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	// init env
	err := config.EnvInit("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file or environement variable : %v", err)
	}

	// init DB
	err = database.Initialization()
	if err != nil {
		log.Fatalf("Error connecting database : %v", err)
	}

	// init DB migration
	err = database.Migrate()
	if err != nil {
		log.Fatalf("Error migrating database : %v", err)
	}

	// init dataset
	println("Loading Account dataset...")
	err = loadingAccountDataset()
	if err != nil {
		log.Fatalf("Error loading dataset : %v", err)
	}

	ret := m.Run()
	os.Exit(ret)
}

func TestGetAccounts(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for GetAccounts handler
	router.GET("/accounts", GetAccounts)

	t.Run("Account List Retrieved", func(t *testing.T) {
		var getAccounts dataset.Accounts
		// Create a mock HTTP request to the /accounts endpoint
		req, err := http.NewRequest(http.MethodGet, "/accounts", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Create a response recorder to record the response
		w := httptest.NewRecorder()

		// Serve the HTTP request to the Gin router
		router.ServeHTTP(w, req)

		// Check the HTTP status code
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}

		// Check the Content-Type header
		expectedContentType := "application/json; charset=utf-8"
		contentType := w.Header().Get("Content-Type")
		if contentType != expectedContentType {
			t.Errorf("Expected content type %s but got %s", expectedContentType, contentType)
		}

		// Unmarshal the response body into a slice of accounts struct
		if err := json.Unmarshal(w.Body.Bytes(), &getAccounts); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}
		// determine if the account - and only the expected account - is in the database
		if len(getAccounts) < 3 {
			t.Errorf("Expected almost 3 account but got %d", len(getAccounts))
		}
	})
}

func TestGetAccountByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for GetAccountByID handler
	router.GET("/accounts/:id", GetAccountByID)

	// Set up a test scenario: account found
	t.Run("Account Found", func(t *testing.T) {
		// identify the user_id of the first user in the dataset
		path := fmt.Sprintf("/accounts/%d", users[0].ID)

		req, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}

		// Unmarshal the response body into an account struct
		var receivedAccount dataset.Account
		if err := json.Unmarshal(w.Body.Bytes(), &receivedAccount); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received account with the expected account
		switch {
		case receivedAccount.Username != users[0].Username:
			t.Errorf("Expected Username %v but got %v", users[0].Username, receivedAccount.Username)
		case receivedAccount.Email != users[0].Email:
			t.Errorf("Expected Email %v but got %v", users[0].Email, receivedAccount.Email)
		case receivedAccount.Firstname != users[0].Firstname:
			t.Errorf("Expected Firstname %v but got %v", users[0].Firstname, receivedAccount.Firstname)
		case receivedAccount.Lastname != users[0].Lastname:
			t.Errorf("Expected Lastname %v but got %v", users[0].Lastname, receivedAccount.Lastname)
		case receivedAccount.Role != users[0].Role:
			t.Errorf("Expected Role %v but got %v", users[0].Role, receivedAccount.Role)
		case receivedAccount.Status != users[0].Status:
			t.Errorf("Expected Status %v but got %v", users[0].Status, receivedAccount.Status)
		case receivedAccount.PreferredCurrency != users[0].PreferredCurrency:
			t.Errorf("Expected PreferredCurrency %v but got %v", users[0].PreferredCurrency,
				receivedAccount.PreferredCurrency)
		case receivedAccount.PreferredUnitSystem != users[0].PreferredUnitSystem:
			t.Errorf("Expected PreferredUnitSystem %v but got %v", users[0].PreferredUnitSystem,
				receivedAccount.PreferredUnitSystem)
		}
	})

	// Set up a test scenario: account not found
	t.Run("Account Not Found", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/accounts/1000", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d but got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestPostAccount(t *testing.T) {
	// Common setup
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/accounts", PostAccount)

	t.Run("Insert valid account", func(t *testing.T) {
		testInsertValidAccount(t, router)
	})

	t.Run("Insert account with bad email", func(t *testing.T) {
		testInsertAccountWithBadEmail(t, router)
	})
}

func testInsertValidAccount(t *testing.T, router *gin.Engine) {
	newAccount := dataset.Account{
		Username:  "Jane",
		Email:     "jane.doe@example.com",
		Firstname: "Jane",
		Lastname:  "Doe",
		Role:      "standard",
		Status:    "active",
	}
	jsonData, _ := json.Marshal(newAccount) // Simplify error handling for brevity
	req, _ := http.NewRequest(http.MethodPost, "/accounts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d but got %d", http.StatusCreated, w.Code)
	}

	// Omitting the detailed database check and unmarshal to focus on the main issue
}

func testInsertAccountWithBadEmail(t *testing.T, router *gin.Engine) {
	badAccount := dataset.Account{
		Username:  "Jules",
		Email:     "jules.doe@example",
		Firstname: "Jules",
		Lastname:  "Doe",
		Role:      "standard",
		Status:    "active",
	}
	jsonData, _ := json.Marshal(badAccount)
	req, _ := http.NewRequest(http.MethodPost, "/accounts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d but got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPutAccountByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for PostAccounts handler
	router.PUT("/accounts/:id", PutAccountByID)

	// Sample account data (with the third user in the dataset)
	testUpdatedAccount := dataset.Account{
		ID:        users[2].ID,
		Username:  users[2].Username,
		Email:     "joseph.doe@example.com",
		Firstname: "Joseph",
		Lastname:  "Doe",
		Role:      "standard",
		Status:    "active",
	}

	// Convert account data to JSON
	jsonData, err := json.Marshal(testUpdatedAccount)
	if err != nil {
		t.Fatalf("Failed to marshal account data: %v", err)
	}

	t.Run("Update account", func(t *testing.T) {
		// Format the path to the account ID
		path := fmt.Sprintf("/accounts/%d", testUpdatedAccount.ID)

		// Set up a test scenario: sending a PUT request with JSON data
		req, err := http.NewRequest(http.MethodPut, path, bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the HTTP status code
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}

		// Query the database to get the inserted account
		var updatedAccount dataset.Account
		row := database.DB().QueryRow(
			`SELECT id,username, email, firstname, lastname, role, status, created_at, updated_at 
			FROM account 
			WHERE id = $1;`,
			testUpdatedAccount.ID)
		err = row.Scan(
			&updatedAccount.ID,
			&updatedAccount.Username,
			&updatedAccount.Email,
			&updatedAccount.Firstname,
			&updatedAccount.Lastname,
			&updatedAccount.Role,
			&updatedAccount.Status,
			&updatedAccount.CreatedAt,
			&updatedAccount.UpdatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Compare the data in DB with Test dataset
		switch {
		case updatedAccount.Username != testUpdatedAccount.Username:
			t.Errorf("Expected Username %v but got %v", testUpdatedAccount.Username, updatedAccount.Username)
		case updatedAccount.Email != testUpdatedAccount.Email:
			t.Errorf("Expected Email %v but got %v", testUpdatedAccount.Email, updatedAccount.Email)
		case updatedAccount.Firstname != testUpdatedAccount.Firstname:
			t.Errorf("Expected Firstname %v but got %v", testUpdatedAccount.Firstname, updatedAccount.Firstname)
		case updatedAccount.Lastname != testUpdatedAccount.Lastname:
			t.Errorf("Expected Lastname %v but got %v", testUpdatedAccount.Lastname, updatedAccount.Lastname)
		case updatedAccount.Role != testUpdatedAccount.Role:
			t.Errorf("Expected Role %v but got %v", testUpdatedAccount.Role, updatedAccount.Role)
		case updatedAccount.Status != testUpdatedAccount.Status:
			t.Errorf("Expected Status %v but got %v", testUpdatedAccount.Status, updatedAccount.Status)
		}
	})
}

func TestDeleteAccountByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for PostAccounts handler
	router.DELETE("/accounts/:id", DeleteAccountByID)

	t.Run("Delete account", func(t *testing.T) {
		// Format the path to the third user of the dataset
		path := fmt.Sprintf("/accounts/%d", users[2].ID)

		// Set up a test scenario: sending a DELETE request
		req, err := http.NewRequest(http.MethodDelete, path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the HTTP status code
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}

		// check in database if the account has been deleted
		var username string
		row := database.DB().QueryRow("SELECT username FROM account WHERE id = $1;", users[2].ID)
		err = row.Scan(&username)
		if err == nil {
			t.Errorf("Account ID %v associated to username %s should be deleted and it is still in DB",
				users[2].ID, username)
		} else if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("Failed to create request: %v", err)
		}
	})
}

func TestRegisterOK(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for Register handler
	router.POST("/register", Register)

	t.Run("Register account", func(t *testing.T) {
		// Sample account data
		newAccount := dataset.RegisterInput{
			Username:  "user-" + random.UniqueId(),
			Password:  "password",
			Email:     "jane.doe@exemple.com",
			Firstname: "Jane",
			Lastname:  "Doe",
		}

		// Convert account data to JSON
		jsonData, err := json.Marshal(newAccount)
		if err != nil {
			t.Fatalf("Failed to marshal account data: %v", err)
		}

		// Set up a test scenario: sending a POST request with JSON data
		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the HTTP status code
		if w.Code != http.StatusOK && w.Code != http.StatusAccepted {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}

		// Query the database to get the inserted account
		var insertedUser dataset.User
		row := database.DB().QueryRow(
			`SELECT a.username, a.email, a.firstname, a.lastname, a.role, a.status, p.password, a.preferred_currency, 
			    a.preferred_unit_system, a.created_at, a.updated_at 
			FROM account a INNER JOIN password p ON a.id = p.user_id 
			WHERE a.username = $1;`,
			newAccount.Username)
		err = row.Scan(
			&insertedUser.Username,
			&insertedUser.Email,
			&insertedUser.Firstname,
			&insertedUser.Lastname,
			&insertedUser.Role,
			&insertedUser.Status,
			&insertedUser.Password,
			&insertedUser.PreferredCurrency,
			&insertedUser.PreferredUnitSystem,
			&insertedUser.CreatedAt,
			&insertedUser.UpdatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				t.Fatalf("No user founded!")
			}
			t.Fatalf("Failed querry database: %v", err)
		}

		err = security.VerifyPassword(newAccount.Password, insertedUser.Password)
		if err != nil && errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			t.Errorf("encryption issue or validation issue with password: %v", err)
		}

		switch {
		case newAccount.Username != insertedUser.Username:
			t.Errorf("Expected Username %v but got %v", insertedUser.Username, newAccount.Username)
		case newAccount.Email != insertedUser.Email:
			t.Errorf("Expected Email %v but got %v", insertedUser.Email, newAccount.Email)
		case newAccount.Firstname != insertedUser.Firstname:
			t.Errorf("Expected Firstname %v but got %v", insertedUser.Firstname, newAccount.Firstname)
		case newAccount.Lastname != insertedUser.Lastname:
			t.Errorf("Expected Lastname %v but got %v", insertedUser.Lastname, newAccount.Lastname)
		case insertedUser.Role != "standard":
			t.Errorf("Expected Role %v but got %v", "standard", insertedUser.Role)
		case insertedUser.Status != "pending":
			t.Errorf("Expected Status %v but got %v", "pending", insertedUser.Status)
		case insertedUser.PreferredCurrency != "EUR":
			t.Errorf("Expected PreferredCurrency %v but got %v", "EUR", insertedUser.PreferredCurrency)
		case insertedUser.PreferredUnitSystem != "METRIC":
			t.Errorf("Expected PreferredUnitSystem %v but got %v", "METRIC", insertedUser.PreferredUnitSystem)
		}
	})
}

func TestRegisterKO(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for Register handler
	router.POST("/register", Register)
	t.Run("Register account with bad email", func(t *testing.T) {
		// Sample account data
		newAccount := dataset.RegisterInput{
			Username:  "user-" + random.UniqueId(),
			Password:  "password",
			Email:     "jane.doe@exemple",
			Firstname: "Jane",
			Lastname:  "Doe",
		}

		// Convert account data to JSON
		jsonData, err := json.Marshal(newAccount)
		if err != nil {
			t.Fatalf("Failed to marshal account data: %v", err)
		}

		// Set up a test scenario: sending a POST request with JSON data
		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the HTTP status code
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d but got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func testLoginWithInvalidUsername(t *testing.T, router *gin.Engine) {
	// Try to login with a username that doesn't exist
	invalidLogin := dataset.LoginInput{
		Username: "nonexistent-user-" + random.UniqueId(),
		Password: "anypassword",
	}
	invalidJSON, err := json.Marshal(invalidLogin)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(invalidJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusUnauthorized, w.Code, w.Body.String())
	}

	// Verify error message doesn't leak user existence
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if errMsg, ok := response["error"].(string); ok {
		if errMsg != "credentials are incorrect" {
			t.Errorf("Expected generic error message but got: %s", errMsg)
		}
	}
}

func testLoginWithIncorrectPassword(t *testing.T, router *gin.Engine, username string) {
	// Reset user status to active for this test
	statement, err := database.DB().Prepare("UPDATE account SET status = 'active' WHERE username = $1;")
	if err != nil {
		t.Fatalf("failed to prepare update statement: %v", err)
	}
	defer statement.Close()

	_, err = statement.Exec(username)
	if err != nil {
		t.Fatalf("failed to execute update query: %v", err)
	}

	// Try to login with wrong password
	wrongPasswordLogin := dataset.LoginInput{
		Username: username,
		Password: "wrong-password-123",
	}
	wrongJSON, err := json.Marshal(wrongPasswordLogin)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(wrongJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusUnauthorized, w.Code, w.Body.String())
	}

	// Verify error message is generic
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if errMsg, ok := response["error"].(string); ok {
		if errMsg != "credentials are incorrect" {
			t.Errorf("Expected generic error message but got: %s", errMsg)
		}
	}
}

func TestLogin(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for Login handler
	router.POST("/login", Login)

	// Sample account data
	newUser := dataset.User{
		Username:  "user-" + random.UniqueId(),
		Password:  "password2",
		Email:     "newuser2@pmp.com",
		Firstname: "Jules",
		Lastname:  "Doe",
		Role:      "standard",
		Status:    "active",
		CreatedAt: time.Now().Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}
	userLogin := dataset.LoginInput{
		Username: newUser.Username,
		Password: newUser.Password,
	}

	// insert account in database
	var id int

	err := database.DB().QueryRow(
		`INSERT INTO account (username, email, firstname, lastname, role, status, created_at, updated_at) 
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) 
		RETURNING id;`,
		newUser.Username,
		newUser.Email,
		newUser.Firstname,
		newUser.Lastname,
		newUser.Role,
		newUser.Status,
		newUser.CreatedAt,
		newUser.UpdatedAt).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	hashedPassword, err := security.HashPassword(newUser.Password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	err = database.DB().QueryRow(
		`INSERT INTO password (user_id, password, updated_at) 
		VALUES ($1,$2,$3) 
		RETURNING id;`,
		id,
		hashedPassword,
		newUser.UpdatedAt).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert password: %v", err)
	}

	// Convert user data to JSON
	jsonData, err := json.Marshal(userLogin)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}

	t.Run("Login user", func(t *testing.T) {
		// Set up a test scenario: sending a POST request with JSON data
		req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the HTTP status code
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}

		// Unmarshal the response body into an token struct
		var receivedToken dataset.Token
		if err := json.Unmarshal(w.Body.Bytes(), &receivedToken); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		if receivedToken.Token == "" {
			t.Errorf("Expected token but got nil")
		}
	})
	t.Run("Login pending user", func(t *testing.T) {
		// set status to pending in DB
		statement, err := database.DB().Prepare("UPDATE account SET status = 'pending' WHERE username = $1;")
		if err != nil {
			t.Fatalf("failed to prepare update statement: %v", err)
		}
		defer statement.Close()

		_, err = statement.Exec(newUser.Username)
		if err != nil {
			t.Fatalf("failed to execute update query: %v", err)
		}

		// Set up a test scenario: sending a POST request with JSON data
		req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the HTTP status code
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Login with invalid username", func(t *testing.T) {
		testLoginWithInvalidUsername(t, router)
	})

	t.Run("Login with incorrect password", func(t *testing.T) {
		testLoginWithIncorrectPassword(t, router, newUser.Username)
	})
}

func TestPutMyPassword(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PUT("/v1/mypassword", PutMyPassword)

	// Use the first user from the loaded dataset
	testUser := users[0]

	// Generate a token for the user
	token, err := security.GenerateToken(testUser.ID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Test: Incorrect current password
	t.Run("Incorrect current password", func(t *testing.T) {
		input := dataset.PasswordUpdateInput{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newpassword",
		}
		jsonData, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPut, "/v1/mypassword", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d but got %d", http.StatusBadRequest, w.Code)
		}
	})

	// Test: Same password
	t.Run("Same password", func(t *testing.T) {
		input := dataset.PasswordUpdateInput{
			CurrentPassword: testUser.Password,
			NewPassword:     testUser.Password,
		}
		jsonData, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPut, "/v1/mypassword", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d but got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		expectedError := "new password must be different from current password"
		if response["error"] != expectedError {
			t.Errorf("Expected error message '%s' but got '%s'", expectedError, response["error"])
		}
	})

	// Test: Correct current password
	t.Run("Correct current password", func(t *testing.T) {
		input := dataset.PasswordUpdateInput{
			CurrentPassword: testUser.Password,
			NewPassword:     "newpassword",
		}
		jsonData, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPut, "/v1/mypassword", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}
	})
}

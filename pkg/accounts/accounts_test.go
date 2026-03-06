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

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/gruntwork-io/terratest/modules/random"
	"golang.org/x/crypto/bcrypt"
)

// mockEmailSender is a no-op email sender for tests.
type mockEmailSender struct{}

func (m *mockEmailSender) SendEmail(_, _, _, _ string) error {
	return nil
}

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

	// inject mock email sender to avoid real SMTP calls in tests
	setEmailSender(&mockEmailSender{})

	// init dataset
	println("Loading Account dataset...")
	err = loadingAccountDataset()
	if err != nil {
		log.Fatalf("Error loading dataset : %v", err)
	}

	// Run tests
	ret := m.Run()

	// Cleanup test data
	println("Cleaning up test data...")
	err = cleanupAccountDataset()
	if err != nil {
		log.Printf("Warning: Error cleaning up dataset : %v", err)
	}

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
		var getAccounts Accounts
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
		var receivedAccount Account
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
	newAccount := Account{
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
	badAccount := Account{
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
	testUpdatedAccount := Account{
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
		var updatedAccount Account
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
		newAccount := RegisterInput{
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
		var insertedUser User
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
	t.Run("Register account with duplicate email", func(t *testing.T) {
		testRegisterWithDuplicateEmail(t, router)
	})
	t.Run("Register account with @ in username", func(t *testing.T) {
		testRegisterWithAtInUsername(t, router)
	})
	t.Run("Register account with bad email", func(t *testing.T) {
		testRegisterWithBadEmail(t, router)
	})
}

func testRegisterWithDuplicateEmail(t *testing.T, router *gin.Engine) {
	t.Helper()
	newAccount := RegisterInput{
		Username:  "user-" + random.UniqueId(),
		Password:  "password",
		Email:     users[0].Email,
		Firstname: "Duplicate",
		Lastname:  "Email",
	}

	jsonData, err := json.Marshal(newAccount)
	if err != nil {
		t.Fatalf("Failed to marshal account data: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusConflict, w.Code, w.Body.String())
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response["error"] != "email already in use" {
		t.Errorf("Expected error 'email already in use' but got '%s'", response["error"])
	}
}

func testRegisterWithAtInUsername(t *testing.T, router *gin.Engine) {
	t.Helper()
	newAccount := RegisterInput{
		Username:  "user@example.com",
		Password:  "password",
		Email:     "valid-" + random.UniqueId() + "@example.com",
		Firstname: "At",
		Lastname:  "Username",
	}

	jsonData, err := json.Marshal(newAccount)
	if err != nil {
		t.Fatalf("Failed to marshal account data: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response["error"] != "username must not contain @" {
		t.Errorf("Expected error 'username must not contain @' but got '%s'", response["error"])
	}
}

func testRegisterWithBadEmail(t *testing.T, router *gin.Engine) {
	t.Helper()
	newAccount := RegisterInput{
		Username:  "user-" + random.UniqueId(),
		Password:  "password",
		Email:     "jane.doe@exemple",
		Firstname: "Jane",
		Lastname:  "Doe",
	}

	jsonData, err := json.Marshal(newAccount)
	if err != nil {
		t.Fatalf("Failed to marshal account data: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d but got %d", http.StatusBadRequest, w.Code)
	}
}

func testLoginWithInvalidUsername(t *testing.T, router *gin.Engine) {
	// Try to login with a username that doesn't exist
	invalidLogin := LoginInput{
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
	wrongPasswordLogin := LoginInput{
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

func testLoginWithEmail(t *testing.T, router *gin.Engine, user User) {
	// Reset user status to active
	_, err := database.DB().Exec("UPDATE account SET status = 'active' WHERE username = $1;", user.Username)
	if err != nil {
		t.Fatalf("failed to reset user status: %v", err)
	}

	emailLogin := LoginInput{
		Username: user.Email,
		Password: user.Password,
	}
	emailJSON, err := json.Marshal(emailLogin)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(emailJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
	validateTokenPairResponse(t, response)
}

func testLoginWithEmailPendingUser(t *testing.T, router *gin.Engine, user User) {
	// Set status to pending
	_, err := database.DB().Exec("UPDATE account SET status = 'pending' WHERE username = $1;", user.Username)
	if err != nil {
		t.Fatalf("failed to set user status to pending: %v", err)
	}

	emailLogin := LoginInput{
		Username: user.Email,
		Password: user.Password,
	}
	emailJSON, err := json.Marshal(emailLogin)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(emailJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
}

func testLoginWithNonexistentEmail(t *testing.T, router *gin.Engine) {
	emailLogin := LoginInput{
		Username: "nonexistent-" + random.UniqueId() + "@example.com",
		Password: "anypassword",
	}
	emailJSON, err := json.Marshal(emailLogin)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(emailJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusUnauthorized, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	errorVal, exists := response["error"]
	if !exists {
		t.Fatalf("Expected 'error' field in response but it was missing. Full response: %+v", response)
	}
	errMsg, ok := errorVal.(string)
	if !ok {
		t.Fatalf("Expected 'error' field to be a string but got %T. Value: %v", errorVal, errorVal)
	}
	if errMsg != "credentials are incorrect" {
		t.Errorf("Expected generic error message %q but got: %q", "credentials are incorrect", errMsg)
	}
}

// validateTokenPairResponse validates the token pair response format
func validateTokenPairResponse(t *testing.T, response map[string]interface{}) {
	t.Helper()

	// Verify new fields
	if response["access_token"] == nil || response["access_token"] == "" {
		t.Errorf("Expected access_token but got nil or empty")
	}
	if response["refresh_token"] == nil || response["refresh_token"] == "" {
		t.Errorf("Expected refresh_token but got nil or empty")
	}

	// Verify backward compatibility - old 'token' field should exist
	if response["token"] == nil || response["token"] == "" {
		t.Errorf("Expected token (backward compatibility) but got nil or empty")
	}

	// Verify token and access_token have the same value
	if response["token"] != response["access_token"] {
		t.Errorf("Expected token and access_token to have the same value")
	}
}

func TestLogin(t *testing.T) {
	router, newUser, jsonData := setupLoginTestData(t)

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

		// Unmarshal the response body into token pair response
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		validateTokenPairResponse(t, response)
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

	t.Run("Login with email", func(t *testing.T) {
		testLoginWithEmail(t, router, newUser)
	})

	t.Run("Login with email of pending user", func(t *testing.T) {
		testLoginWithEmailPendingUser(t, router, newUser)
	})

	t.Run("Login with nonexistent email", func(t *testing.T) {
		testLoginWithNonexistentEmail(t, router)
	})

	t.Run("Login with invalid username", func(t *testing.T) {
		testLoginWithInvalidUsername(t, router)
	})

	t.Run("Login with incorrect password", func(t *testing.T) {
		testLoginWithIncorrectPassword(t, router, newUser.Username)
	})
}

func TestResendConfirmEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/resend-confirmemail", ResendConfirmEmail)

	// The second user in the dataset has status "pending"
	pendingUser := users[1]

	t.Run("Resend for pending account returns 200", func(t *testing.T) {
		input := ResendConfirmEmailInput{Email: pendingUser.Email}
		jsonData, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPost, "/resend-confirmemail", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response["message"] == "" {
			t.Error("Expected a message in response")
		}
	})

	t.Run("Resend for non-existent email returns 200 (anti-enumeration)", func(t *testing.T) {
		input := ResendConfirmEmailInput{Email: "nonexistent@example.com"}
		jsonData, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPost, "/resend-confirmemail", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Resend for active account returns 200 (anti-enumeration)", func(t *testing.T) {
		// The first user in the dataset has status "active"
		input := ResendConfirmEmailInput{Email: users[0].Email}
		jsonData, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPost, "/resend-confirmemail", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Resend with invalid email format returns 400", func(t *testing.T) {
		input := ResendConfirmEmailInput{Email: "not-an-email"}
		jsonData, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPost, "/resend-confirmemail", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d but got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Resend with empty body returns 400", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/resend-confirmemail", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d but got %d", http.StatusBadRequest, w.Code)
		}
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
		input := PasswordUpdateInput{
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
		input := PasswordUpdateInput{
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
		input := PasswordUpdateInput{
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

func TestPutMyAccountImagePosition(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PUT("/v1/myaccount", PutMyAccount)
	router.GET("/v1/myaccount", GetMyAccount)

	testUser := users[0]
	token, err := security.GenerateToken(testUser.ID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	t.Run("Default image position is 50/50", func(t *testing.T) {
		testDefaultImagePosition(t, router, token)
	})
	t.Run("Update image position to 30/70", func(t *testing.T) {
		testUpdateImagePosition(t, router, token, testUser)
	})
	t.Run("Omitting image position keeps existing values", func(t *testing.T) {
		testOmittedImagePositionPreserved(t, router, token, testUser)
	})
	t.Run("Out of range image position returns 400", func(t *testing.T) {
		testImagePositionOutOfRange(t, router, token, testUser)
	})
}

func testDefaultImagePosition(t *testing.T, router *gin.Engine, token string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, "/v1/myaccount", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 but got %d: %s", w.Code, w.Body.String())
	}

	var account Account
	if err := json.Unmarshal(w.Body.Bytes(), &account); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if account.ImagePositionX != 50 || account.ImagePositionY != 50 {
		t.Errorf("Expected default position 50/50 but got %d/%d",
			account.ImagePositionX, account.ImagePositionY)
	}
}

func testUpdateImagePosition(t *testing.T, router *gin.Engine, token string, user User) {
	t.Helper()
	posX := 30
	posY := 70
	input := AccountUpdateInput{
		Email:          user.Email,
		Firstname:      user.Firstname,
		Lastname:       user.Lastname,
		ImagePositionX: &posX,
		ImagePositionY: &posY,
	}
	jsonData, _ := json.Marshal(input)
	req, _ := http.NewRequest(http.MethodPut, "/v1/myaccount", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 but got %d: %s", w.Code, w.Body.String())
	}

	var account Account
	if err := json.Unmarshal(w.Body.Bytes(), &account); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if account.ImagePositionX != 30 || account.ImagePositionY != 70 {
		t.Errorf("Expected position 30/70 but got %d/%d",
			account.ImagePositionX, account.ImagePositionY)
	}
}

func testOmittedImagePositionPreserved(t *testing.T, router *gin.Engine, token string, user User) {
	t.Helper()
	input := AccountUpdateInput{
		Email:     user.Email,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
	}
	jsonData, _ := json.Marshal(input)
	req, _ := http.NewRequest(http.MethodPut, "/v1/myaccount", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 but got %d: %s", w.Code, w.Body.String())
	}

	var account Account
	if err := json.Unmarshal(w.Body.Bytes(), &account); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	// Should still be 30/70 from previous test
	if account.ImagePositionX != 30 || account.ImagePositionY != 70 {
		t.Errorf("Expected position 30/70 preserved but got %d/%d",
			account.ImagePositionX, account.ImagePositionY)
	}
}

func testImagePositionOutOfRange(t *testing.T, router *gin.Engine, token string, user User) {
	t.Helper()
	posX := 150
	posY := 50
	input := AccountUpdateInput{
		Email:          user.Email,
		Firstname:      user.Firstname,
		Lastname:       user.Lastname,
		ImagePositionX: &posX,
		ImagePositionY: &posY,
	}
	jsonData, _ := json.Marshal(input)
	req, _ := http.NewRequest(http.MethodPut, "/v1/myaccount", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 but got %d: %s", w.Code, w.Body.String())
	}
}

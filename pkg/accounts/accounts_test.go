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
	err = database.DatabaseInit()
	if err != nil {
		log.Fatalf("Error connecting database : %v", err)
	}

	// init DB migration
	err = database.DatabaseMigrate()
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
		req, err := http.NewRequest("GET", "/accounts", nil)
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

		req, err := http.NewRequest("GET", path, nil)
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
		}
	})

	// Set up a test scenario: account not found
	t.Run("Account Not Found", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/accounts/1000", nil)
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
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for PostAccounts handler
	router.POST("/accounts", PostAccount)

	// Sample account data
	newAccount := dataset.Account{
		Username:  "Jane",
		Email:     "jane.doe@example.com",
		Firstname: "Jane",
		Lastname:  "Doe",
		Role:      "standard",
		Status:    "active",
	}

	// Convert account data to JSON
	jsonData, err := json.Marshal(newAccount)
	if err != nil {
		t.Fatalf("Failed to marshal account data: %v", err)
	}

	t.Run("Insert account", func(t *testing.T) {

		// Set up a test scenario: sending a POST request with JSON data
		req, err := http.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the HTTP status code
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d but got %d", http.StatusCreated, w.Code)
		}

		// Query the database to get the inserted account
		var insertedAccount dataset.Account
		row := database.Db().QueryRow("SELECT id,username, email, firstname, lastname, role, status, created_at, updated_at FROM account WHERE username = $1;", newAccount.Username)
		err = row.Scan(&insertedAccount.ID, &insertedAccount.Username, &insertedAccount.Email, &insertedAccount.Firstname, &insertedAccount.Lastname, &insertedAccount.Role, &insertedAccount.Status, &insertedAccount.Created_at, &insertedAccount.Updated_at)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Unmarshal the response body into an account struct
		var receivedAccount dataset.Account
		if err := json.Unmarshal(w.Body.Bytes(), &receivedAccount); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received account with the expected account data
		switch {
		case receivedAccount.Username != insertedAccount.Username:
			t.Errorf("Expected Username %v but got %v", insertedAccount.Username, receivedAccount.Username)
		case receivedAccount.Email != insertedAccount.Email:
			t.Errorf("Expected Email %v but got %v", insertedAccount.Email, receivedAccount.Email)
		case receivedAccount.Firstname != insertedAccount.Firstname:
			t.Errorf("Expected Firstname %v but got %v", insertedAccount.Firstname, receivedAccount.Firstname)
		case receivedAccount.Lastname != insertedAccount.Lastname:
			t.Errorf("Expected Lastname %v but got %v", insertedAccount.Lastname, receivedAccount.Lastname)
		case receivedAccount.Role != insertedAccount.Role:
			t.Errorf("Expected Role %v but got %v", insertedAccount.Role, receivedAccount.Role)
		case receivedAccount.Status != insertedAccount.Status:
			t.Errorf("Expected Status %v but got %v", insertedAccount.Status, receivedAccount.Status)
		}
	})
}

func TestPutAccountByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for PostAccounts handler
	router.PUT("/accounts/:id", PutAccountByID)

	// Sample account data (with the third user in the dataset)
	TestUpdatedAccount := dataset.Account{
		ID:        users[2].ID,
		Username:  users[2].Username,
		Email:     "joseph.doe@example.com",
		Firstname: "Joseph",
		Lastname:  "Doe",
		Role:      "standard",
		Status:    "active",
	}

	// Convert account data to JSON
	jsonData, err := json.Marshal(TestUpdatedAccount)
	if err != nil {
		t.Fatalf("Failed to marshal account data: %v", err)
	}

	t.Run("Update account", func(t *testing.T) {

		// Format the path to the account ID
		path := fmt.Sprintf("/accounts/%d", TestUpdatedAccount.ID)

		// Set up a test scenario: sending a PUT request with JSON data
		req, err := http.NewRequest("PUT", path, bytes.NewBuffer(jsonData))
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
		row := database.Db().QueryRow("SELECT id,username, email, firstname, lastname, role, status, created_at, updated_at FROM account WHERE id = $1;", TestUpdatedAccount.ID)
		err = row.Scan(&updatedAccount.ID, &updatedAccount.Username, &updatedAccount.Email, &updatedAccount.Firstname, &updatedAccount.Lastname, &updatedAccount.Role, &updatedAccount.Status, &updatedAccount.Created_at, &updatedAccount.Updated_at)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Compare the data in DB with Test dataset
		switch {
		case updatedAccount.Username != TestUpdatedAccount.Username:
			t.Errorf("Expected Username %v but got %v", TestUpdatedAccount.Username, updatedAccount.Username)
		case updatedAccount.Email != TestUpdatedAccount.Email:
			t.Errorf("Expected Email %v but got %v", TestUpdatedAccount.Email, updatedAccount.Email)
		case updatedAccount.Firstname != TestUpdatedAccount.Firstname:
			t.Errorf("Expected Firstname %v but got %v", TestUpdatedAccount.Firstname, updatedAccount.Firstname)
		case updatedAccount.Lastname != TestUpdatedAccount.Lastname:
			t.Errorf("Expected Lastname %v but got %v", TestUpdatedAccount.Lastname, updatedAccount.Lastname)
		case updatedAccount.Role != TestUpdatedAccount.Role:
			t.Errorf("Expected Role %v but got %v", TestUpdatedAccount.Role, updatedAccount.Role)
		case updatedAccount.Status != TestUpdatedAccount.Status:
			t.Errorf("Expected Status %v but got %v", TestUpdatedAccount.Status, updatedAccount.Status)
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
		req, err := http.NewRequest("DELETE", path, nil)
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
		row := database.Db().QueryRow("SELECT username FROM account WHERE id = $1;", users[2].ID)
		err = row.Scan(&username)
		if err == nil {
			t.Errorf("Account ID %v associated to username %s should be deleted and it is still in DB", users[2].ID, username)
		} else if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("Failed to create request: %v", err)

		}
	})
}

func TestRegister(t *testing.T) {

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for Register handler
	router.POST("/register", Register)

	// Sample account data
	newAccount := dataset.RegisterInput{
		Username:  fmt.Sprintf("user-%s", random.UniqueId()),
		Password:  "password",
		Email:     "jane.doe@pmp.com",
		Firstname: "Jane",
		Lastname:  "Doe",
	}

	// Convert account data to JSON
	jsonData, err := json.Marshal(newAccount)
	if err != nil {
		t.Fatalf("Failed to marshal account data: %v", err)
	}

	t.Run("Register account", func(t *testing.T) {

		// Set up a test scenario: sending a POST request with JSON data
		req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
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
		var insertedUser dataset.User
		row := database.Db().QueryRow("SELECT a.username, a.email, a.firstname, a.lastname, a.role, a.status, p.password, a.created_at, a.updated_at FROM account a INNER JOIN password p ON a.id = p.user_id WHERE a.username = $1;", newAccount.Username)
		err = row.Scan(&insertedUser.Username, &insertedUser.Email, &insertedUser.Firstname, &insertedUser.Lastname, &insertedUser.Role, &insertedUser.Status, &insertedUser.Password, &insertedUser.Created_at, &insertedUser.Updated_at)
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
		}
	})
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
		Username:   fmt.Sprintf("user-%s", random.UniqueId()),
		Password:   "password2",
		Email:      "newuser2@pmp.com",
		Firstname:  "Jules",
		Lastname:   "Doe",
		Role:       "standard",
		Status:     "active",
		Created_at: time.Now().Truncate(time.Second),
		Updated_at: time.Now().Truncate(time.Second),
	}
	userLogin := dataset.LoginInput{
		Username: newUser.Username,
		Password: newUser.Password,
	}

	// insert account in database
	var id int

	err := database.Db().QueryRow("INSERT INTO account (username, email, firstname, lastname, role, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id;", newUser.Username, newUser.Email, newUser.Firstname, newUser.Lastname, newUser.Role, newUser.Status, newUser.Created_at, newUser.Updated_at).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	hashedPassword, err := security.HashPassword(newUser.Password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	err = database.Db().QueryRow("INSERT INTO password (user_id, password, updated_at) VALUES ($1,$2,$3) RETURNING id;", id, hashedPassword, newUser.Updated_at).Scan(&id)
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
		req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
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
		statement, err := database.Db().Prepare("UPDATE account SET status = 'pending' WHERE username = $1;")
		if err != nil {
			t.Fatalf("failed to prepare update statement: %v", err)
		}
		defer statement.Close()

		_, err = statement.Exec(newUser.Username)
		if err != nil {
			t.Fatalf("failed to execute update query: %v", err)
		}

		// Set up a test scenario: sending a POST request with JSON data
		req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
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
}

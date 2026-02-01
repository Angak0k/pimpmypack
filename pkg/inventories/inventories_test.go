package inventories

import (
	"bytes"
	"context"
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
	"github.com/google/go-cmp/cmp"
)

func TestMain(m *testing.M) {
	// init env
	err := config.EnvInit("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file or environment variable : %v", err)
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
	err = loadingInventoryDataset()
	if err != nil {
		log.Fatalf("Error loading dataset : %v", err)
	}

	// Run tests
	ret := m.Run()

	// Cleanup test data
	println("Cleaning up test data...")
	err = cleanupInventoryDataset()
	if err != nil {
		log.Printf("Warning: Error cleaning up dataset : %v", err)
	}

	os.Exit(ret)
}

func TestGetInventories(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for GetInventories handler
	router.GET("/inventories", GetInventories)

	t.Run("Inventories List Retrieved", func(t *testing.T) {
		// Create a mock HTTP request to the /inventories endpoint
		req, err := http.NewRequest(http.MethodGet, "/inventories", nil)
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

		// Unmarshal the response body into a slice of inventories struct
		var getInventories Inventories
		if err := json.Unmarshal(w.Body.Bytes(), &getInventories); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Debug: Print all inventories in the response
		t.Logf("Response inventories:")
		for i, inv := range getInventories {
			t.Logf("Response[%d]: UserID=%d, ItemName=%s", i, inv.UserID, inv.ItemName)
		}

		// Debug: Print expected inventory
		t.Logf("Expected inventory: UserID=%d, ItemName=%s", inventories[0].UserID, inventories[0].ItemName)

		// Check if we have at least 3 inventories
		if len(getInventories) < 3 {
			t.Errorf("Expected almost 3 inventory but got %d", len(getInventories))
			return
		}

		// Find and validate the expected inventory
		validateInventory(t, getInventories, inventories[0])
	})
}

// validateInventory checks if the expected inventory exists in the response and validates its fields
func validateInventory(t *testing.T, responseInventories Inventories, expectedInventory Inventory) {
	// Find the matching inventory in the response by both UserID and ItemName
	var foundInventory *Inventory
	for _, inv := range responseInventories {
		if inv.ItemName == expectedInventory.ItemName && inv.UserID == expectedInventory.UserID {
			foundInventory = &inv
			break
		}
	}

	if foundInventory == nil {
		t.Errorf("Expected inventory with UserID=%d and ItemName=%s not found in response",
			expectedInventory.UserID, expectedInventory.ItemName)
		return
	}

	// Compare the found inventory with the expected inventory
	compareInventoryFields(t, foundInventory, &expectedInventory)
}

// compareInventoryFields compares all fields between the found and expected inventory
func compareInventoryFields(t *testing.T, found, expected *Inventory) {
	switch {
	case !cmp.Equal(found.UserID, expected.UserID):
		t.Errorf("Expected UserID %v but got %v", expected.UserID, found.UserID)
	case !cmp.Equal(found.ItemName, expected.ItemName):
		t.Errorf("Expected Item Name %v but got %v", expected.ItemName, found.ItemName)
	case !cmp.Equal(found.Category, expected.Category):
		t.Errorf("Expected Category %v but got %v", expected.Category, found.Category)
	case !cmp.Equal(found.Description, expected.Description):
		t.Errorf("Expected Description %v but got %v", expected.Description, found.Description)
	case !cmp.Equal(found.Weight, expected.Weight):
		t.Errorf("Expected Weight %v but got %v", expected.Weight, found.Weight)
	case !cmp.Equal(found.URL, expected.URL):
		t.Errorf("Expected URL %v but got %v", expected.URL, found.URL)
	case !cmp.Equal(found.Price, expected.Price):
		t.Errorf("Expected Price %v but got %v", expected.Price, found.Price)
	case !cmp.Equal(found.Currency, expected.Currency):
		t.Errorf("Expected Currency %v but got %v", expected.Currency, found.Currency)
	}
}

func TestGetMyInventory(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for GetMyInventory handler
	router.GET("/myinventory", GetMyInventory)

	t.Run("Inventories List Retrieved", func(t *testing.T) {
		token, err := security.GenerateToken(users[0].ID)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}
		// Create a mock HTTP request to the /myinventory endpoint
		req, err := http.NewRequest(http.MethodGet, "/myinventory", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Set up a test scenario: sending a GET request
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

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

		// Unmarshal the response body into a slice of inventories struct
		var myInventories Inventories
		if err := json.Unmarshal(w.Body.Bytes(), &myInventories); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// determine if the inventory - and only the expected inventory - is the expected content
		if len(myInventories) < 3 {
			t.Errorf("Expected almost 3 inventory but got %d", len(myInventories))
		} else {
			switch {
			case !cmp.Equal(myInventories[0].UserID, inventories[0].UserID):
				t.Errorf("Expected UserID %v but got %v", inventories[0].UserID, myInventories[0].UserID)
			case !cmp.Equal(myInventories[0].ItemName, inventories[0].ItemName):
				t.Errorf("Expected Item Name %v but got %v", inventories[0].ItemName, myInventories[0].ItemName)
			case !cmp.Equal(myInventories[0].Category, inventories[0].Category):
				t.Errorf("Expected Category %v but got %v", inventories[0].Category, myInventories[0].Category)
			case !cmp.Equal(myInventories[0].Description, inventories[0].Description):
				t.Errorf("Expected Description %v but got %v", inventories[0].Description, myInventories[0].Description)
			case !cmp.Equal(myInventories[0].Weight, inventories[0].Weight):
				t.Errorf("Expected Weight %v but got %v", inventories[0].Weight, myInventories[0].Weight)
			case !cmp.Equal(myInventories[0].URL, inventories[0].URL):
				t.Errorf("Expected URL %v but got %v", inventories[0].URL, myInventories[0].URL)
			case !cmp.Equal(myInventories[0].Price, inventories[0].Price):
				t.Errorf("Expected Price %v but got %v", inventories[0].Price, myInventories[0].Price)
			case !cmp.Equal(myInventories[0].Currency, inventories[0].Currency):
				t.Errorf("Expected Currency %v but got %v", inventories[0].Currency, myInventories[0].Currency)
			case cmp.Equal(myInventories[1].UpdatedAt, inventories[1].UpdatedAt):
				t.Errorf("Expected UpdatedAt %v should be different than %v",
					inventories[1].UpdatedAt, myInventories[1].UpdatedAt)
			}
		}
	})
}

func TestGetInventoryByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for GetInventoryByID handler
	router.GET("/inventories/:id", GetInventoryByID)

	t.Run("Inventory Retrieved", func(t *testing.T) {
		// Create a mock HTTP request to the /inventories endpoint
		path := fmt.Sprintf("/inventories/%d", inventories[0].ID)
		req, err := http.NewRequest(http.MethodGet, path, nil)
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

		// Unmarshal the response body into an inventory struct
		var receivedInventory Inventory
		if err := json.Unmarshal(w.Body.Bytes(), &receivedInventory); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received Inventory with the expected Inventory
		switch {
		case receivedInventory.UserID != inventories[0].UserID:
			t.Errorf("Expected UserID %v but got %v", inventories[0].UserID, receivedInventory.UserID)
		case receivedInventory.ItemName != inventories[0].ItemName:
			t.Errorf("Expected ItemName %v but got %v", inventories[0].ItemName, receivedInventory.ItemName)
		case receivedInventory.Category != inventories[0].Category:
			t.Errorf("Expected Category %v but got %v", inventories[0].Category, receivedInventory.Category)
		case receivedInventory.Description != inventories[0].Description:
			t.Errorf("Expected Description %v but got %v", inventories[0].Description, receivedInventory.Description)
		case receivedInventory.Weight != inventories[0].Weight:
			t.Errorf("Expected Weight %v but got %v", inventories[0].Weight, receivedInventory.Weight)
		case receivedInventory.URL != inventories[0].URL:
			t.Errorf("Expected URL %v but got %v", inventories[0].URL, receivedInventory.URL)
		case receivedInventory.Price != inventories[0].Price:
			t.Errorf("Expected Price %v but got %v", inventories[0].Price, receivedInventory.Price)
		case receivedInventory.Currency != inventories[0].Currency:
			t.Errorf("Expected Currency %v but got %v", inventories[0].Currency, receivedInventory.Currency)
		}
	})

	// Set up a test scenario: inventory not found
	t.Run("Inventory Not Found", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/inventories/1000", nil)
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

func TestPostInventory(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for PostInventory handler
	router.POST("/inventories", PostInventory)

	// Sample inventory data
	newInventory := Inventory{
		UserID:      users[0].ID,
		ItemName:    "Light",
		Category:    "Outdoor Gear",
		Description: "Headed Light",
		Weight:      29, // Weight in grams
		URL:         "https://example.com/light",
		Price:       30,
		Currency:    "USD",
	}

	// Convert inventory data to JSON
	jsonData, err := json.Marshal(newInventory)
	if err != nil {
		t.Fatalf("Failed to marshal inventory data: %v", err)
	}

	t.Run("Insert Inventory", func(t *testing.T) {
		// Set up a test scenario: sending a POST request with JSON data
		req, err := http.NewRequest(http.MethodPost, "/inventories", bytes.NewBuffer(jsonData))
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

		// Query the database to get the inserted inventory
		var insertedInventory Inventory
		row := database.DB().QueryRow("SELECT * FROM inventory WHERE item_name = $1;", newInventory.ItemName)
		err = row.Scan(
			&insertedInventory.ID,
			&insertedInventory.UserID,
			&insertedInventory.ItemName,
			&insertedInventory.Category,
			&insertedInventory.Description,
			&insertedInventory.Weight,
			&insertedInventory.URL,
			&insertedInventory.Price,
			&insertedInventory.Currency,
			&insertedInventory.CreatedAt,
			&insertedInventory.UpdatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Unmarshal the response body into an inventory struct
		var receivedInventory Inventory
		if err := json.Unmarshal(w.Body.Bytes(), &receivedInventory); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received inventory with the expected inventory data
		switch {
		case receivedInventory.ID != insertedInventory.ID:
			t.Errorf("Expected ID %v but got %v", insertedInventory.ID, receivedInventory.ID)
		case receivedInventory.UserID != insertedInventory.UserID:
			t.Errorf("Expected User_ID %v but got %v", insertedInventory.UserID, receivedInventory.UserID)
		case receivedInventory.ItemName != insertedInventory.ItemName:
			t.Errorf("Expected ItemName %v but got %v", insertedInventory.ItemName, receivedInventory.ItemName)
		case receivedInventory.Category != insertedInventory.Category:
			t.Errorf("Expected Category %v but got %v", insertedInventory.Category, receivedInventory.Category)
		case receivedInventory.Description != insertedInventory.Description:
			t.Errorf("Expected Description %v but got %v", insertedInventory.Description, receivedInventory.Description)
		case receivedInventory.Weight != insertedInventory.Weight:
			t.Errorf("Expected Weight %v but got %v", insertedInventory.Weight, receivedInventory.Weight)
		case receivedInventory.URL != insertedInventory.URL:
			t.Errorf("Expected URL %v but got %v", insertedInventory.URL, receivedInventory.URL)
		case receivedInventory.Price != insertedInventory.Price:
			t.Errorf("Expected Price %v but got %v", insertedInventory.Price, receivedInventory.Price)
		case receivedInventory.Currency != insertedInventory.Currency:
			t.Errorf("Expected Currency %v but got %v", insertedInventory.Currency, receivedInventory.Currency)
		}
	})
}

func TestPutInventoryByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for Postinventorys handler
	router.PUT("/inventories/:id", PutInventoryByID)

	// Sample inventory data (with the first user of the dataset and the second inventory of the dataset)
	testUpdatedInventory := Inventory{
		ID:          inventories[1].ID,
		UserID:      users[0].ID,
		ItemName:    "Tent",
		Category:    "Outdoor Gear",
		Description: "Lightweight tent for camping",
		Weight:      1200, // Weight in grams
		URL:         "https://example.com/tent",
		Price:       200,
		Currency:    "USD",
	}

	// Convert inventory data to JSON
	jsonData, err := json.Marshal(testUpdatedInventory)
	if err != nil {
		t.Fatalf("Failed to marshal inventory data: %v", err)
	}

	t.Run("Update Inventory", func(t *testing.T) {
		// Set up a test scenario: sending a PUT request with JSON data
		path := fmt.Sprintf("/inventories/%d", testUpdatedInventory.ID)
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

		// Query the database to get the updated inventories
		var updatedInventory Inventory
		row := database.DB().QueryRow("SELECT * FROM inventory WHERE id = $1;", testUpdatedInventory.ID)
		err = row.Scan(
			&updatedInventory.ID,
			&updatedInventory.UserID,
			&updatedInventory.ItemName,
			&updatedInventory.Category,
			&updatedInventory.Description,
			&updatedInventory.Weight,
			&updatedInventory.URL,
			&updatedInventory.Price,
			&updatedInventory.Currency,
			&updatedInventory.CreatedAt,
			&updatedInventory.UpdatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Compare the data in DB with Test dataset
		switch {
		case updatedInventory.ItemName != testUpdatedInventory.ItemName:
			t.Errorf("Expected ItemName %v but got %v", testUpdatedInventory.ItemName, updatedInventory.ItemName)
		case updatedInventory.Category != testUpdatedInventory.Category:
			t.Errorf("Expected Category %v but got %v", testUpdatedInventory.Category, updatedInventory.Category)
		case updatedInventory.Description != testUpdatedInventory.Description:
			t.Errorf("Expected Description %v but got %v",
				testUpdatedInventory.Description, updatedInventory.Description)
		case updatedInventory.Weight != testUpdatedInventory.Weight:
			t.Errorf("Expected Weight %v but got %v", testUpdatedInventory.Weight, updatedInventory.Weight)
		case updatedInventory.URL != testUpdatedInventory.URL:
			t.Errorf("Expected URL %v but got %v", testUpdatedInventory.URL, updatedInventory.URL)
		case updatedInventory.Price != testUpdatedInventory.Price:
			t.Errorf("Expected Price %v but got %v", testUpdatedInventory.Price, updatedInventory.Price)
		}
	})
}

func TestDeleteInventoryByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for PostInventorys handler
	router.DELETE("/inventories/:id", DeleteInventoryByID)

	t.Run("Delete inventory", func(t *testing.T) {
		// Set up a test scenario: sending a DELETE request with the third inventory of the dataset
		path := fmt.Sprintf("/inventories/%d", inventories[2].ID)
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

		// check in database if the inventory has been deleted
		var itemName string
		row := database.DB().QueryRow("SELECT item_name FROM inventory WHERE id = $1;", inventories[2].ID)
		err = row.Scan(&itemName)
		if err == nil {
			t.Errorf("Inventory ID 3 associated to item_name %s should be deleted and it is still in DB", itemName)
		} else if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("Failed to create request: %v", err)
		}
	})
}

func TestCheckInventoryOwnership(t *testing.T) {
	t.Run("Test Pack Ownership", func(t *testing.T) {
		inventoryID := inventories[0].ID
		userID := users[0].ID
		wrongUserID := users[1].ID

		myInventory, err := checkInventoryOwnership(context.Background(), inventoryID, userID)
		if err != nil {
			t.Fatalf("Failed to check inventory ownership: %v", err)
		}
		if !myInventory {
			t.Errorf("Expected true but got false")
		}

		notmyInventory, err := checkInventoryOwnership(context.Background(), inventoryID, wrongUserID)
		if err != nil {
			t.Fatalf("Failed to check inventory ownership: %v", err)
		}
		if notmyInventory {
			t.Errorf("Expected false but got true")
		}
	})
}

func TestGetMyInventoryByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()
	// Define the endpoint for GetMyInventoryByID handler
	router.GET("/myinventory/:id", GetMyInventoryByID)

	t.Run("Item Retrieved", func(t *testing.T) {
		token, err := security.GenerateToken(users[0].ID)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}
		// Create a mock HTTP request to the /myinventory endpoint
		path := fmt.Sprintf("/myinventory/%d", inventories[0].ID)
		req, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Set up a test scenario: sending a GET request
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Create a response recorder to record the response
		w := httptest.NewRecorder()

		// Serve the HTTP request to the Gin router
		router.ServeHTTP(w, req)

		// Check the HTTP status code
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}

		// Unmarshal the response body into an inventory struct
		var receivedInventory Inventory
		if err := json.Unmarshal(w.Body.Bytes(), &receivedInventory); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received Inventory with the expected Inventory
		switch {
		case receivedInventory.UserID != inventories[0].UserID:
			t.Errorf("Expected UserID %v but got %v", inventories[0].UserID, receivedInventory.UserID)
		case receivedInventory.ItemName != inventories[0].ItemName:
			t.Errorf("Expected ItemName %v but got %v", inventories[0].ItemName, receivedInventory.ItemName)
		case receivedInventory.Category != inventories[0].Category:
			t.Errorf("Expected Category %v but got %v", inventories[0].Category, receivedInventory.Category)
		case receivedInventory.Description != inventories[0].Description:
			t.Errorf("Expected Description %v but got %v", inventories[0].Description, receivedInventory.Description)
		case receivedInventory.Weight != inventories[0].Weight:
			t.Errorf("Expected Weight %v but got %v", inventories[0].Weight, receivedInventory.Weight)
		case receivedInventory.URL != inventories[0].URL:
			t.Errorf("Expected URL %v but got %v", inventories[0].URL, receivedInventory.URL)
		case receivedInventory.Price != inventories[0].Price:
			t.Errorf("Expected Price %v but got %v", inventories[0].Price, receivedInventory.Price)
		case receivedInventory.Currency != inventories[0].Currency:
			t.Errorf("Expected Currency %v but got %v", inventories[0].Currency, receivedInventory.Currency)
		}
	})
	t.Run("Item Forbiden", func(t *testing.T) {
		token, err := security.GenerateToken(users[1].ID)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}
		// Create a mock HTTP request to the /myinventory endpoint
		path := fmt.Sprintf("/myinventory/%d", inventories[0].ID)
		req, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Set up a test scenario: sending a GET request
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Create a response recorder to record the response
		w := httptest.NewRecorder()

		// Serve the HTTP request to the Gin router
		router.ServeHTTP(w, req)

		// Check the HTTP status code
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d but got %d", http.StatusForbidden, w.Code)
		}
	})
}

// Helper functions to reduce cognitive complexity
func testFindExistingItem(ctx context.Context, t *testing.T) {
	expected := inventories[0]
	found, err := FindInventoryItemByAttributes(
		ctx,
		expected.UserID,
		expected.ItemName,
		expected.Category,
		expected.Description,
	)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find item, got nil")
	}

	if found.ID != expected.ID {
		t.Errorf("Expected ID %d, got %d", expected.ID, found.ID)
	}
	if found.ItemName != expected.ItemName {
		t.Errorf("Expected ItemName %s, got %s", expected.ItemName, found.ItemName)
	}
	if found.Category != expected.Category {
		t.Errorf("Expected Category %s, got %s", expected.Category, found.Category)
	}
}

func testItemNotFound(
	ctx context.Context,
	t *testing.T,
	userID uint,
	itemName, category, description string,
) {
	found, err := FindInventoryItemByAttributes(ctx, userID, itemName, category, description)

	if !errors.Is(err, ErrNoItemFound) {
		t.Fatalf("Expected ErrNoItemFound, got: %v", err)
	}

	if found != nil {
		t.Errorf("Expected nil, found item with ID %d", found.ID)
	}
}

func testFindEmptyDescription(ctx context.Context, t *testing.T) {
	var testItem Inventory
	testItem.UserID = users[0].ID
	testItem.ItemName = "Test Item Empty Desc"
	testItem.Category = "Test Category"
	testItem.Description = ""
	testItem.Weight = 100
	testItem.Price = 50
	testItem.Currency = "USD"

	err := InsertInventory(ctx, &testItem)
	if err != nil {
		t.Fatalf("Failed to insert test item: %v", err)
	}

	found, err := FindInventoryItemByAttributes(ctx, testItem.UserID, testItem.ItemName, testItem.Category, "")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find item with empty description, got nil")
	}

	if found.ID != testItem.ID {
		t.Errorf("Expected ID %d, got %d", testItem.ID, found.ID)
	}
}

func TestFindInventoryItemByAttributes(t *testing.T) {
	ctx := context.Background()

	t.Run("Find existing item by all attributes", func(t *testing.T) {
		testFindExistingItem(ctx, t)
	})

	t.Run("Item not found - different user", func(t *testing.T) {
		testItemNotFound(
			ctx, t, 9999,
			inventories[0].ItemName, inventories[0].Category, inventories[0].Description,
		)
	})

	t.Run("Item not found - different name", func(t *testing.T) {
		testItemNotFound(
			ctx, t, inventories[0].UserID,
			"Non-existent Item Name", inventories[0].Category, inventories[0].Description,
		)
	})

	t.Run("Item not found - different category", func(t *testing.T) {
		testItemNotFound(
			ctx, t, inventories[0].UserID,
			inventories[0].ItemName, "Different Category", inventories[0].Description,
		)
	})

	t.Run("Item not found - different description", func(t *testing.T) {
		testItemNotFound(
			ctx, t, inventories[0].UserID,
			inventories[0].ItemName, inventories[0].Category, "Different Description",
		)
	})

	t.Run("Find item with empty description", func(t *testing.T) {
		testFindEmptyDescription(ctx, t)
	})
}

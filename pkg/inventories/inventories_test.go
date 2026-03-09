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
	"time"

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

func TestGetInventoriesPackCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/inventories", GetInventories)

	t.Run("PackCount values are correct", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/inventories", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var allInventories Inventories
		if err := json.Unmarshal(w.Body.Bytes(), &allInventories); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Build a lookup by ID
		byID := make(map[uint]Inventory)
		for _, inv := range allInventories {
			byID[inv.ID] = inv
		}

		// Backpack (inventories[0]) is in 2 packs
		if got := byID[inventories[0].ID].PackCount; got != 2 {
			t.Errorf("Expected Backpack PackCount=2, got %d", got)
		}
		// Tent (inventories[1]) is in 1 pack
		if got := byID[inventories[1].ID].PackCount; got != 1 {
			t.Errorf("Expected Tent PackCount=1, got %d", got)
		}
		// Sleeping Bag (inventories[2]) is in 0 packs
		if got := byID[inventories[2].ID].PackCount; got != 0 {
			t.Errorf("Expected Sleeping Bag PackCount=0, got %d", got)
		}
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
	case !cmp.Equal(found.PackCount, expected.PackCount):
		t.Errorf("Expected PackCount %v but got %v", expected.PackCount, found.PackCount)
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

	// Sample inventory data using admin request struct
	newInventory := InventoryCreateAdminRequest{
		UserID:      users[0].ID,
		ItemName:    "Light",
		Category:    "Outdoor Gear",
		Description: "Headed Light",
		Weight:      29, // Weight in grams
		URL:         "https://example.com/light",
		Price:       30,
		Currency:    "USD",
	}

	// Debug: print the UserID
	t.Logf("DEBUG: users[0].ID = %d", users[0].ID)

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

		// Unmarshal the response body into an inventory struct
		var receivedInventory Inventory
		if err := json.Unmarshal(w.Body.Bytes(), &receivedInventory); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Query the database by the ID returned in the response for deterministic matching
		var insertedInventory Inventory
		row := database.DB().QueryRow(
			`SELECT id, user_id, item_name, category, description, weight, url, price, currency, created_at, updated_at
			 FROM inventory WHERE id = $1;`,
			receivedInventory.ID)
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
		row := database.DB().QueryRow(
			`SELECT id, user_id, item_name, category, description, weight, url, price, currency, created_at, updated_at
			 FROM inventory WHERE id = $1;`,
			testUpdatedInventory.ID)
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

func TestPostMyInventoryMerge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()

	t.Run("Merge with shared pack - quantities summed", func(t *testing.T) {
		// Setup merge test data
		err := createMergeTestData(ctx)
		if err != nil {
			t.Fatalf("Failed to create merge test data: %v", err)
		}
		defer cleanupMergeTestData(ctx)

		source := mergeTestItems[0]
		target := mergeTestItems[1]

		router := gin.Default()
		router.POST("/myinventory/merge", PostMyInventoryMerge)

		token, err := security.GenerateToken(users[0].ID)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		mergeReq := MergeInventoryRequest{
			SourceItemID: int(source.ID),
			TargetItemID: int(target.ID),
			ItemName:     "Merged Item",
			Category:     "Merged Category",
			Description:  "Merged description",
			Weight:       300,
			URL:          "https://example.com/merged",
			Price:        30,
			Currency:     "EUR",
			ImageSource:  "none",
		}

		jsonData, err := json.Marshal(mergeReq)
		if err != nil {
			t.Fatalf("Failed to marshal merge request: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, "/myinventory/merge", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
			return
		}

		var mergedItem Inventory
		if err := json.Unmarshal(w.Body.Bytes(), &mergedItem); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Verify merged item properties
		if mergedItem.ID != target.ID {
			t.Errorf("Expected merged item ID %d but got %d", target.ID, mergedItem.ID)
		}
		if mergedItem.ItemName != "Merged Item" {
			t.Errorf("Expected item name 'Merged Item' but got '%s'", mergedItem.ItemName)
		}
		if mergedItem.Category != "Merged Category" {
			t.Errorf("Expected category 'Merged Category' but got '%s'", mergedItem.Category)
		}
		if mergedItem.Weight != 300 {
			t.Errorf("Expected weight 300 but got %d", mergedItem.Weight)
		}

		// Verify target is in 2 packs (shared + non-shared reassigned)
		if mergedItem.PackCount != 2 {
			t.Errorf("Expected pack count 2 but got %d", mergedItem.PackCount)
		}

		// Verify source item is deleted
		_, err = findInventoryByID(ctx, source.ID)
		if !errors.Is(err, ErrNoItemFound) {
			t.Errorf("Expected source item to be deleted, got err: %v", err)
		}

		// Verify shared pack quantity was summed (2 + 3 = 5)
		var quantity int
		err = database.DB().QueryRowContext(ctx,
			`SELECT quantity FROM pack_content WHERE pack_id = $1 AND item_id = $2`,
			mergeTestPackIDs[0], target.ID).Scan(&quantity)
		if err != nil {
			t.Fatalf("Failed to query pack_content: %v", err)
		}
		if quantity != 5 {
			t.Errorf("Expected quantity 5 in shared pack but got %d", quantity)
		}

		// Verify non-shared pack was reassigned to target
		var nonSharedQuantity int
		err = database.DB().QueryRowContext(ctx,
			`SELECT quantity FROM pack_content WHERE pack_id = $1 AND item_id = $2`,
			mergeTestPackIDs[1], target.ID).Scan(&nonSharedQuantity)
		if err != nil {
			t.Fatalf("Failed to query non-shared pack_content: %v", err)
		}
		if nonSharedQuantity != 1 {
			t.Errorf("Expected quantity 1 in non-shared pack but got %d", nonSharedQuantity)
		}
	})

	t.Run("Merge with no pack overlap", func(t *testing.T) {
		err := createMergeTestData(ctx)
		if err != nil {
			t.Fatalf("Failed to create merge test data: %v", err)
		}
		defer cleanupMergeTestData(ctx)

		source := mergeTestItems[0]
		target := mergeTestItems[1]

		// Remove target from shared pack so there's no overlap
		_, err = database.DB().ExecContext(ctx,
			`DELETE FROM pack_content WHERE pack_id = $1 AND item_id = $2`,
			mergeTestPackIDs[0], target.ID)
		if err != nil {
			t.Fatalf("Failed to remove target from shared pack: %v", err)
		}

		router := gin.Default()
		router.POST("/myinventory/merge", PostMyInventoryMerge)

		token, err := security.GenerateToken(users[0].ID)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		mergeReq := MergeInventoryRequest{
			SourceItemID: int(source.ID),
			TargetItemID: int(target.ID),
			ItemName:     "No Overlap Merged",
			Category:     "Test",
			Description:  "",
			Weight:       150,
			URL:          "",
			Price:        15,
			Currency:     "USD",
			ImageSource:  "none",
		}

		jsonData, _ := json.Marshal(mergeReq)
		req, _ := http.NewRequest(http.MethodPost, "/myinventory/merge", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
			return
		}

		var mergedItem Inventory
		if err := json.Unmarshal(w.Body.Bytes(), &mergedItem); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Source had 2 packs (shared + non-shared), all reassigned to target
		if mergedItem.PackCount != 2 {
			t.Errorf("Expected pack count 2 but got %d", mergedItem.PackCount)
		}
	})

	t.Run("Merge with image_source=source", func(t *testing.T) {
		err := createMergeTestData(ctx)
		if err != nil {
			t.Fatalf("Failed to create merge test data: %v", err)
		}
		defer cleanupMergeTestData(ctx)

		source := mergeTestItems[0]
		target := mergeTestItems[1]

		// Insert a source image
		now := time.Now()
		_, err = database.DB().ExecContext(ctx,
			`INSERT INTO inventory_images (item_id, image_data, mime_type, file_size, width, height, uploaded_at, updated_at)
			VALUES ($1, $2, 'image/jpeg', 100, 50, 50, $3, $4)`,
			source.ID, []byte("fake-source-image"), now, now)
		if err != nil {
			t.Fatalf("Failed to insert source image: %v", err)
		}

		// Insert a target image
		_, err = database.DB().ExecContext(ctx,
			`INSERT INTO inventory_images (item_id, image_data, mime_type, file_size, width, height, uploaded_at, updated_at)
			VALUES ($1, $2, 'image/jpeg', 200, 100, 100, $3, $4)`,
			target.ID, []byte("fake-target-image"), now, now)
		if err != nil {
			t.Fatalf("Failed to insert target image: %v", err)
		}

		router := gin.Default()
		router.POST("/myinventory/merge", PostMyInventoryMerge)

		token, _ := security.GenerateToken(users[0].ID)
		mergeReq := MergeInventoryRequest{
			SourceItemID: int(source.ID),
			TargetItemID: int(target.ID),
			ItemName:     "Image Source Test",
			Category:     "Test",
			Weight:       100,
			Currency:     "USD",
			ImageSource:  "source",
		}

		jsonData, _ := json.Marshal(mergeReq)
		req, _ := http.NewRequest(http.MethodPost, "/myinventory/merge", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
			return
		}

		var mergedItem Inventory
		json.Unmarshal(w.Body.Bytes(), &mergedItem)

		// Verify the merged item has an image (source image was reassigned to target)
		if !mergedItem.HasImage {
			t.Errorf("Expected merged item to have image (source image reassigned)")
		}

		// Verify the image data is the source image
		var imageData []byte
		err = database.DB().QueryRowContext(ctx,
			`SELECT image_data FROM inventory_images WHERE item_id = $1`, target.ID).Scan(&imageData)
		if err != nil {
			t.Fatalf("Failed to query image: %v", err)
		}
		if string(imageData) != "fake-source-image" {
			t.Errorf("Expected source image data but got different data")
		}
	})

	t.Run("Merge with image_source=target", func(t *testing.T) {
		err := createMergeTestData(ctx)
		if err != nil {
			t.Fatalf("Failed to create merge test data: %v", err)
		}
		defer cleanupMergeTestData(ctx)

		source := mergeTestItems[0]
		target := mergeTestItems[1]

		// Insert a target image only
		now := time.Now()
		_, err = database.DB().ExecContext(ctx,
			`INSERT INTO inventory_images (item_id, image_data, mime_type, file_size, width, height, uploaded_at, updated_at)
			VALUES ($1, $2, 'image/jpeg', 200, 100, 100, $3, $4)`,
			target.ID, []byte("fake-target-image"), now, now)
		if err != nil {
			t.Fatalf("Failed to insert target image: %v", err)
		}

		router := gin.Default()
		router.POST("/myinventory/merge", PostMyInventoryMerge)

		token, _ := security.GenerateToken(users[0].ID)
		mergeReq := MergeInventoryRequest{
			SourceItemID: int(source.ID),
			TargetItemID: int(target.ID),
			ItemName:     "Image Target Test",
			Category:     "Test",
			Weight:       100,
			Currency:     "USD",
			ImageSource:  "target",
		}

		jsonData, _ := json.Marshal(mergeReq)
		req, _ := http.NewRequest(http.MethodPost, "/myinventory/merge", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
			return
		}

		var mergedItem Inventory
		json.Unmarshal(w.Body.Bytes(), &mergedItem)

		// Verify the merged item still has the target image
		if !mergedItem.HasImage {
			t.Errorf("Expected merged item to have image (target image kept)")
		}
	})

	t.Run("Merge with image_source=none", func(t *testing.T) {
		err := createMergeTestData(ctx)
		if err != nil {
			t.Fatalf("Failed to create merge test data: %v", err)
		}
		defer cleanupMergeTestData(ctx)

		source := mergeTestItems[0]
		target := mergeTestItems[1]

		// Insert images for both
		now := time.Now()
		_, _ = database.DB().ExecContext(ctx,
			`INSERT INTO inventory_images (item_id, image_data, mime_type, file_size, width, height, uploaded_at, updated_at)
			VALUES ($1, $2, 'image/jpeg', 100, 50, 50, $3, $4)`,
			source.ID, []byte("fake-source-image"), now, now)
		_, _ = database.DB().ExecContext(ctx,
			`INSERT INTO inventory_images (item_id, image_data, mime_type, file_size, width, height, uploaded_at, updated_at)
			VALUES ($1, $2, 'image/jpeg', 200, 100, 100, $3, $4)`,
			target.ID, []byte("fake-target-image"), now, now)

		router := gin.Default()
		router.POST("/myinventory/merge", PostMyInventoryMerge)

		token, _ := security.GenerateToken(users[0].ID)
		mergeReq := MergeInventoryRequest{
			SourceItemID: int(source.ID),
			TargetItemID: int(target.ID),
			ItemName:     "Image None Test",
			Category:     "Test",
			Weight:       100,
			Currency:     "USD",
			ImageSource:  "none",
		}

		jsonData, _ := json.Marshal(mergeReq)
		req, _ := http.NewRequest(http.MethodPost, "/myinventory/merge", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
			return
		}

		var mergedItem Inventory
		json.Unmarshal(w.Body.Bytes(), &mergedItem)

		// Verify no image on the merged item
		if mergedItem.HasImage {
			t.Errorf("Expected merged item to have no image with image_source=none")
		}
	})

	t.Run("Merge fails - source equals target", func(t *testing.T) {
		router := gin.Default()
		router.POST("/myinventory/merge", PostMyInventoryMerge)

		token, _ := security.GenerateToken(users[0].ID)
		mergeReq := MergeInventoryRequest{
			SourceItemID: int(inventories[0].ID),
			TargetItemID: int(inventories[0].ID),
			ItemName:     "Same Item",
			Category:     "Test",
			Currency:     "USD",
			ImageSource:  "none",
		}

		jsonData, _ := json.Marshal(mergeReq)
		req, _ := http.NewRequest(http.MethodPost, "/myinventory/merge", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d but got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Merge fails - ownership validation", func(t *testing.T) {
		router := gin.Default()
		router.POST("/myinventory/merge", PostMyInventoryMerge)

		// User 2 tries to merge user 1's items
		token, _ := security.GenerateToken(users[1].ID)
		mergeReq := MergeInventoryRequest{
			SourceItemID: int(inventories[0].ID),
			TargetItemID: int(inventories[1].ID),
			ItemName:     "Stolen Merge",
			Category:     "Test",
			Currency:     "USD",
			ImageSource:  "none",
		}

		jsonData, _ := json.Marshal(mergeReq)
		req, _ := http.NewRequest(http.MethodPost, "/myinventory/merge", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d but got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("Merge fails - source not found", func(t *testing.T) {
		router := gin.Default()
		router.POST("/myinventory/merge", PostMyInventoryMerge)

		token, _ := security.GenerateToken(users[0].ID)
		mergeReq := MergeInventoryRequest{
			SourceItemID: 999999,
			TargetItemID: int(inventories[0].ID),
			ItemName:     "Missing Source",
			Category:     "Test",
			Currency:     "USD",
			ImageSource:  "none",
		}

		jsonData, _ := json.Marshal(mergeReq)
		req, _ := http.NewRequest(http.MethodPost, "/myinventory/merge", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d but got %d", http.StatusNotFound, w.Code)
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

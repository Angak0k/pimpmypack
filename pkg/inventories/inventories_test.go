package inventories

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
	"github.com/Angak0k/pimpmypack/pkg/dataset"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
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
	err = loadingInventoryDataset()
	if err != nil {
		log.Fatalf("Error loading dataset : %v", err)
	}

	ret := m.Run()
	os.Exit(ret)
}

func TestGetInventories(t *testing.T) {
	var getInventories dataset.Inventories
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for GetInventories handler
	router.GET("/inventories", GetInventories)

	t.Run("Inventories List Retrieved", func(t *testing.T) {
		// Create a mock HTTP request to the /inventories endpoint
		req, err := http.NewRequest("GET", "/inventories", nil)
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
		if err := json.Unmarshal(w.Body.Bytes(), &getInventories); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}
		// determine if the inventory - and only the expected inventory - is in the database
		if len(getInventories) < 3 {
			t.Errorf("Expected almost 3 inventory but got %d", len(getInventories))
		} else {
			switch {
			case !cmp.Equal(getInventories[0].User_id, inventories[0].User_id):
				t.Errorf("Expected User_id %v but got %v", inventories[0].User_id, getInventories[0].User_id)
			case !cmp.Equal(getInventories[0].Item_name, inventories[0].Item_name):
				t.Errorf("Expected Item Name %v but got %v", inventories[0].Item_name, getInventories[0].Item_name)
			case !cmp.Equal(getInventories[0].Category, inventories[0].Category):
				t.Errorf("Expected Category %v but got %v", inventories[0].Category, getInventories[0].Category)
			case !cmp.Equal(getInventories[0].Description, inventories[0].Description):
				t.Errorf("Expected Description %v but got %v", inventories[0].Description, getInventories[0].Description)
			case !cmp.Equal(getInventories[0].Weight, inventories[0].Weight):
				t.Errorf("Expected Weight %v but got %v", inventories[0].Weight, getInventories[0].Weight)
			case !cmp.Equal(getInventories[0].Weight_unit, inventories[0].Weight_unit):
				t.Errorf("Expected Weight_unit %v but got %v", inventories[0].Weight_unit, getInventories[0].Weight_unit)
			case !cmp.Equal(getInventories[0].Url, inventories[0].Url):
				t.Errorf("Expected Url %v but got %v", inventories[0].Url, getInventories[0].Url)
			case !cmp.Equal(getInventories[0].Price, inventories[0].Price):
				t.Errorf("Expected Price %v but got %v", inventories[0].Price, getInventories[0].Price)
			case !cmp.Equal(getInventories[0].Currency, inventories[0].Currency):
				t.Errorf("Expected Currency %v but got %v", inventories[0].Currency, getInventories[0].Currency)
			}
		}
	})
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
		req, err := http.NewRequest("GET", "/myinventory", nil)
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
		var myInventories dataset.Inventories
		if err := json.Unmarshal(w.Body.Bytes(), &myInventories); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// determine if the inventory - and only the expected inventory - is the expected content
		if len(myInventories) < 3 {
			t.Errorf("Expected almost 3 inventory but got %d", len(myInventories))
		} else {
			switch {
			case !cmp.Equal(myInventories[0].User_id, inventories[0].User_id):
				t.Errorf("Expected User_id %v but got %v", inventories[0].User_id, myInventories[0].User_id)
			case !cmp.Equal(myInventories[0].Item_name, inventories[0].Item_name):
				t.Errorf("Expected Item Name %v but got %v", inventories[0].Item_name, myInventories[0].Item_name)
			case !cmp.Equal(myInventories[0].Category, inventories[0].Category):
				t.Errorf("Expected Category %v but got %v", inventories[0].Category, myInventories[0].Category)
			case !cmp.Equal(myInventories[0].Description, inventories[0].Description):
				t.Errorf("Expected Description %v but got %v", inventories[0].Description, myInventories[0].Description)
			case !cmp.Equal(myInventories[0].Weight, inventories[0].Weight):
				t.Errorf("Expected Weight %v but got %v", inventories[0].Weight, myInventories[0].Weight)
			case !cmp.Equal(myInventories[0].Weight_unit, inventories[0].Weight_unit):
				t.Errorf("Expected Weight_unit %v but got %v", inventories[0].Weight_unit, myInventories[0].Weight_unit)
			case !cmp.Equal(myInventories[0].Url, inventories[0].Url):
				t.Errorf("Expected Url %v but got %v", inventories[0].Url, myInventories[0].Url)
			case !cmp.Equal(myInventories[0].Price, inventories[0].Price):
				t.Errorf("Expected Price %v but got %v", inventories[0].Price, myInventories[0].Price)
			case !cmp.Equal(myInventories[0].Currency, inventories[0].Currency):
				t.Errorf("Expected Currency %v but got %v", inventories[0].Currency, myInventories[0].Currency)
			case cmp.Equal(myInventories[1].Updated_at, inventories[1].Updated_at):
				t.Errorf("Expected Updated_at %v should be different than %v", inventories[1].Updated_at, myInventories[1].Updated_at)
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
		req, err := http.NewRequest("GET", path, nil)
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
		var receivedInventory dataset.Inventory
		if err := json.Unmarshal(w.Body.Bytes(), &receivedInventory); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received Inventory with the expected Inventory
		switch {
		case receivedInventory.User_id != inventories[0].User_id:
			t.Errorf("Expected User_id %v but got %v", inventories[0].User_id, receivedInventory.User_id)
		case receivedInventory.Item_name != inventories[0].Item_name:
			t.Errorf("Expected Item_name %v but got %v", inventories[0].Item_name, receivedInventory.Item_name)
		case receivedInventory.Category != inventories[0].Category:
			t.Errorf("Expected Category %v but got %v", inventories[0].Category, receivedInventory.Category)
		case receivedInventory.Description != inventories[0].Description:
			t.Errorf("Expected Description %v but got %v", inventories[0].Description, receivedInventory.Description)
		case receivedInventory.Weight != inventories[0].Weight:
			t.Errorf("Expected Weight %v but got %v", inventories[0].Weight, receivedInventory.Weight)
		case receivedInventory.Weight_unit != inventories[0].Weight_unit:
			t.Errorf("Expected Weight_unit %v but got %v", inventories[0].Weight_unit, receivedInventory.Weight_unit)
		case receivedInventory.Url != inventories[0].Url:
			t.Errorf("Expected Url %v but got %v", inventories[0].Url, receivedInventory.Url)
		case receivedInventory.Price != inventories[0].Price:
			t.Errorf("Expected Price %v but got %v", inventories[0].Price, receivedInventory.Price)
		case receivedInventory.Currency != inventories[0].Currency:
			t.Errorf("Expected Currency %v but got %v", inventories[0].Currency, receivedInventory.Currency)
		}
	})

	// Set up a test scenario: inventory not found
	t.Run("Inventory Not Found", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/inventories/1000", nil)
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
	newInventory := dataset.Inventory{
		User_id:     users[0].ID,
		Item_name:   "Light",
		Category:    "Outdoor Gear",
		Description: "Headed Light",
		Weight:      29,
		Weight_unit: "METRIC",
		Url:         "https://example.com/light",
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
		req, err := http.NewRequest("POST", "/inventories", bytes.NewBuffer(jsonData))
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
		var insertedInventory dataset.Inventory
		row := database.Db().QueryRow("SELECT * FROM inventory WHERE item_name = $1;", newInventory.Item_name)
		err = row.Scan(&insertedInventory.ID, &insertedInventory.User_id, &insertedInventory.Item_name, &insertedInventory.Category, &insertedInventory.Description, &insertedInventory.Weight, &insertedInventory.Weight_unit, &insertedInventory.Url, &insertedInventory.Price, &insertedInventory.Currency, &insertedInventory.Created_at, &insertedInventory.Updated_at)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Unmarshal the response body into an inventory struct
		var receivedInventory dataset.Inventory
		if err := json.Unmarshal(w.Body.Bytes(), &receivedInventory); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received inventory with the expected inventory data
		switch {
		case receivedInventory.ID != insertedInventory.ID:
			t.Errorf("Expected ID %v but got %v", insertedInventory.ID, receivedInventory.ID)
		case receivedInventory.User_id != insertedInventory.User_id:
			t.Errorf("Expected User_ID %v but got %v", insertedInventory.User_id, receivedInventory.User_id)
		case receivedInventory.Item_name != insertedInventory.Item_name:
			t.Errorf("Expected Item_name %v but got %v", insertedInventory.Item_name, receivedInventory.Item_name)
		case receivedInventory.Category != insertedInventory.Category:
			t.Errorf("Expected Category %v but got %v", insertedInventory.Category, receivedInventory.Category)
		case receivedInventory.Description != insertedInventory.Description:
			t.Errorf("Expected Description %v but got %v", insertedInventory.Description, receivedInventory.Description)
		case receivedInventory.Weight != insertedInventory.Weight:
			t.Errorf("Expected Weight %v but got %v", insertedInventory.Weight, receivedInventory.Weight)
		case receivedInventory.Weight_unit != insertedInventory.Weight_unit:
			t.Errorf("Expected Weight_unit %v but got %v", insertedInventory.Weight_unit, receivedInventory.Weight_unit)
		case receivedInventory.Url != insertedInventory.Url:
			t.Errorf("Expected Url %v but got %v", insertedInventory.Url, receivedInventory.Url)
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
	TestUpdatedInventory := dataset.Inventory{
		ID:          inventories[1].ID,
		User_id:     users[0].ID,
		Item_name:   "Tent",
		Category:    "Outdoor Gear",
		Description: "Lightweight tent for camping",
		Weight:      1200,
		Weight_unit: "METRIC",
		Url:         "https://example.com/tent",
		Price:       200,
		Currency:    "USD",
	}

	// Convert inventory data to JSON
	jsonData, err := json.Marshal(TestUpdatedInventory)
	if err != nil {
		t.Fatalf("Failed to marshal inventory data: %v", err)
	}

	t.Run("Update Inventory", func(t *testing.T) {

		// Set up a test scenario: sending a PUT request with JSON data
		path := fmt.Sprintf("/inventories/%d", TestUpdatedInventory.ID)
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

		// Query the database to get the updated inventories
		var updatedInventory dataset.Inventory
		row := database.Db().QueryRow("SELECT * FROM inventory WHERE id = $1;", TestUpdatedInventory.ID)
		err = row.Scan(&updatedInventory.ID, &updatedInventory.User_id, &updatedInventory.Item_name, &updatedInventory.Category, &updatedInventory.Description, &updatedInventory.Weight, &updatedInventory.Weight_unit, &updatedInventory.Url, &updatedInventory.Price, &updatedInventory.Currency, &updatedInventory.Created_at, &updatedInventory.Updated_at)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Compare the data in DB with Test dataset
		switch {
		case updatedInventory.Item_name != TestUpdatedInventory.Item_name:
			t.Errorf("Expected Item_name %v but got %v", TestUpdatedInventory.Item_name, updatedInventory.Item_name)
		case updatedInventory.Category != TestUpdatedInventory.Category:
			t.Errorf("Expected Category %v but got %v", TestUpdatedInventory.Category, updatedInventory.Category)
		case updatedInventory.Description != TestUpdatedInventory.Description:
			t.Errorf("Expected Description %v but got %v", TestUpdatedInventory.Description, updatedInventory.Description)
		case updatedInventory.Weight != TestUpdatedInventory.Weight:
			t.Errorf("Expected Weight %v but got %v", TestUpdatedInventory.Weight, updatedInventory.Weight)
		case updatedInventory.Weight_unit != TestUpdatedInventory.Weight_unit:
			t.Errorf("Expected Weight_unit %v but got %v", TestUpdatedInventory.Weight_unit, updatedInventory.Weight_unit)
		case updatedInventory.Url != TestUpdatedInventory.Url:
			t.Errorf("Expected Url %v but got %v", TestUpdatedInventory.Url, updatedInventory.Url)
		case updatedInventory.Price != TestUpdatedInventory.Price:
			t.Errorf("Expected Price %v but got %v", TestUpdatedInventory.Price, updatedInventory.Price)
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

		// check in database if the inventory has been deleted
		var item_name string
		row := database.Db().QueryRow("SELECT item_name FROM inventory WHERE id = $1;", inventories[2].ID)
		err = row.Scan(&item_name)
		if err == nil {
			t.Errorf("Inventory ID 3 associated to item_name %s should be deleted and it is still in DB", item_name)
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

		myInventory, err := checkInventoryOwnership(inventoryID, userID)
		if err != nil {
			t.Fatalf("Failed to check inventory ownership: %v", err)
		}
		if !myInventory {
			t.Errorf("Expected true but got false")
		}

		notmyInventory, err := checkInventoryOwnership(inventoryID, wrongUserID)
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
		req, err := http.NewRequest("GET", path, nil)
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
		var receivedInventory dataset.Inventory
		if err := json.Unmarshal(w.Body.Bytes(), &receivedInventory); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received Inventory with the expected Inventory
		switch {
		case receivedInventory.User_id != inventories[0].User_id:
			t.Errorf("Expected User_id %v but got %v", inventories[0].User_id, receivedInventory.User_id)
		case receivedInventory.Item_name != inventories[0].Item_name:
			t.Errorf("Expected Item_name %v but got %v", inventories[0].Item_name, receivedInventory.Item_name)
		case receivedInventory.Category != inventories[0].Category:
			t.Errorf("Expected Category %v but got %v", inventories[0].Category, receivedInventory.Category)
		case receivedInventory.Description != inventories[0].Description:
			t.Errorf("Expected Description %v but got %v", inventories[0].Description, receivedInventory.Description)
		case receivedInventory.Weight != inventories[0].Weight:
			t.Errorf("Expected Weight %v but got %v", inventories[0].Weight, receivedInventory.Weight)
		case receivedInventory.Weight_unit != inventories[0].Weight_unit:
			t.Errorf("Expected Weight_unit %v but got %v", inventories[0].Weight_unit, receivedInventory.Weight_unit)
		case receivedInventory.Url != inventories[0].Url:
			t.Errorf("Expected Url %v but got %v", inventories[0].Url, receivedInventory.Url)
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
		req, err := http.NewRequest("GET", path, nil)
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

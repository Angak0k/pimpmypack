package packs

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
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
	println("Environement variables loaded")

	// init DB
	err = database.DatabaseInit()
	if err != nil {
		log.Fatalf("Error connecting database : %v", err)
	}
	println("Database connected")

	// init DB migration
	err = database.DatabaseMigrate()
	if err != nil {
		log.Fatalf("Error migrating database : %v", err)
	}
	println("Database migrated")

	// init dataset
	println("Loading dataset...")
	err = loadingPackDataset()
	if err != nil {
		log.Fatalf("Error loading dataset : %v", err)
	}
	println("Dataset loaded...")
	ret := m.Run()
	os.Exit(ret)
}

func TestGetPacks(t *testing.T) {
	var getPacks dataset.Packs
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for GetPacks handler
	router.GET("/packs", GetPacks)

	t.Run("Pack List Retrieved", func(t *testing.T) {
		// Create a mock HTTP request to the /packs endpoint
		req, err := http.NewRequest("GET", "/packs", nil)
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

		// Unmarshal the response body into a slice of packs struct
		if err := json.Unmarshal(w.Body.Bytes(), &getPacks); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}
		// determine if the pack - and only the expected pack - is in the database
		if len(getPacks) < 3 {
			t.Errorf("Expected almost 3 pack but got %d", len(getPacks))
		} else {
			switch {
			case !cmp.Equal(getPacks[0].User_id, packs[0].User_id):
				t.Errorf("Expected User ID %v but got %v", packs[0].User_id, getPacks[0].User_id)
			case !cmp.Equal(getPacks[0].Pack_name, packs[0].Pack_name):
				t.Errorf("Expected Pack Name %v but got %v", packs[0].Pack_name, getPacks[0].Pack_name)
			case !cmp.Equal(getPacks[0].Pack_description, packs[0].Pack_description):
				t.Errorf("Expected Pack Description %v but got %v", packs[0].Pack_description, getPacks[0].Pack_description)
			case !cmp.Equal(getPacks[0].Sharing_code, packs[0].Sharing_code):
				t.Errorf("Expected Sharing Code %v but got %v", packs[0].Sharing_code, getPacks[0].Sharing_code)
			case !cmp.Equal(getPacks[1].User_id, packs[1].User_id):
				t.Errorf("Expected User ID %v but got %v", packs[1].User_id, getPacks[1].User_id)
			case !cmp.Equal(getPacks[1].Pack_name, packs[1].Pack_name):
				t.Errorf("Expected Pack Name %v but got %v", packs[1].Pack_name, getPacks[1].Pack_name)
			case !cmp.Equal(getPacks[1].Pack_description, packs[1].Pack_description):
				t.Errorf("Expected Pack Description %v but got %v", packs[1].Pack_description, getPacks[1].Pack_description)
			case !cmp.Equal(getPacks[1].Sharing_code, packs[1].Sharing_code):
				t.Errorf("Expected Sharing Code %v but got %v", packs[1].Sharing_code, getPacks[1].Sharing_code)
			case !cmp.Equal(getPacks[2].User_id, packs[2].User_id):
				t.Errorf("Expected User ID %v but got %v", packs[2].User_id, getPacks[2].User_id)
			case !cmp.Equal(getPacks[2].Pack_name, packs[2].Pack_name):
				t.Errorf("Expected Pack Name %v but got %v", packs[2].Pack_name, getPacks[2].Pack_name)
			case !cmp.Equal(getPacks[2].Pack_description, packs[2].Pack_description):
				t.Errorf("Expected Pack Description %v but got %v", packs[2].Pack_description, getPacks[2].Pack_description)
			case !cmp.Equal(getPacks[2].Sharing_code, packs[2].Sharing_code):
				t.Errorf("Expected Sharing Code %v but got %v", packs[2].Sharing_code, getPacks[2].Sharing_code)
			}
		}
	})
}

func TestGetPackByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for GetPackByID handler
	router.GET("/packs/:id", GetPackByID)

	// Set up a test scenario: pack found
	t.Run("Pack Found", func(t *testing.T) {
		path := fmt.Sprintf("/packs/%d", packs[0].ID)
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}

		// Unmarshal the response body into a pack struct
		var receivedPack dataset.Pack
		if err := json.Unmarshal(w.Body.Bytes(), &receivedPack); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received pack with the expected pack
		switch {
		case receivedPack.User_id != packs[0].User_id:
			t.Errorf("Expected User ID %v but got %v", packs[0].User_id, receivedPack.User_id)
		case receivedPack.Pack_name != packs[0].Pack_name:
			t.Errorf("Expected Pack Name %v but got %v", packs[0].Pack_name, receivedPack.Pack_name)
		case receivedPack.Pack_description != packs[0].Pack_description:
			t.Errorf("Expected Pack Description %v but got %v", packs[0].Pack_description, receivedPack.Pack_description)
		case receivedPack.Sharing_code != packs[0].Sharing_code:
			t.Errorf("Expected Sharing Code %v but got %v", packs[0].Sharing_code, receivedPack.Sharing_code)
		}
	})

	// Set up a test scenario: pack not found
	t.Run("Pack Not Found", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/packs/1000", nil)
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

func TestPostPack(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for PostPacks handler
	router.POST("/packs", PostPack)

	// Sample pack data
	newPack := dataset.Pack{
		User_id:          users[0].ID,
		Pack_name:        "SomePack",
		Pack_description: "This is a new pack",
	}

	// Convert pack data to JSON
	jsonData, err := json.Marshal(newPack)
	if err != nil {
		t.Fatalf("Failed to marshal pack data: %v", err)
	}

	t.Run("Insert pack", func(t *testing.T) {

		// Set up a test scenario: sending a POST request with JSON data
		req, err := http.NewRequest("POST", "/packs", bytes.NewBuffer(jsonData))
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

		// Query the database to get the inserted pack
		var insertedPack dataset.Pack
		row := database.Db().QueryRow("SELECT id, user_id, pack_name, pack_description, sharing_code, created_at, updated_at FROM pack WHERE pack_name = $1;", newPack.Pack_name)
		err = row.Scan(&insertedPack.ID, &insertedPack.User_id, &insertedPack.Pack_name, &insertedPack.Pack_description, &insertedPack.Sharing_code, &insertedPack.Created_at, &insertedPack.Updated_at)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Unmarshal the response body into a pack struct
		var receivedPack dataset.Pack
		if err := json.Unmarshal(w.Body.Bytes(), &receivedPack); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received pack with the expected pack data
		switch {
		case receivedPack.User_id != insertedPack.User_id:
			t.Errorf("Expected User ID %v but got %v", insertedPack.User_id, receivedPack.User_id)
		case receivedPack.Pack_name != insertedPack.Pack_name:
			t.Errorf("Expected Pack Name %v but got %v", insertedPack.Pack_name, receivedPack.Pack_name)
		case receivedPack.Pack_description != insertedPack.Pack_description:
			t.Errorf("Expected Pack Description %v but got %v", insertedPack.Pack_description, receivedPack.Pack_description)
		case receivedPack.Sharing_code == "":
			t.Errorf("Expected a non empty Sharing Code but got %v", receivedPack.Sharing_code)
		}
	})
}

func TestPutPackByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for PutPacks handler
	router.PUT("/packs/:id", PutPackByID)

	// Sample pack data
	TestUpdatedPack := dataset.Pack{
		User_id:          users[1].ID,
		Pack_name:        "Amazing Pack",
		Pack_description: "Updated pack description",
	}

	// Convert pack data to JSON
	jsonData, err := json.Marshal(TestUpdatedPack)
	if err != nil {
		t.Fatalf("Failed to marshal pack data: %v", err)
	}

	t.Run("Update pack", func(t *testing.T) {

		// Set up a test scenario: sending a PUT request with JSON data
		path := fmt.Sprintf("/packs/%d", packs[2].ID)
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

		// Query the database to get the updated pack
		var updatedPack dataset.Pack
		row := database.Db().QueryRow("SELECT id, user_id, pack_name, pack_description, sharing_code, created_at, updated_at FROM pack WHERE id = $1;", packs[2].ID)
		err = row.Scan(&updatedPack.ID, &updatedPack.User_id, &updatedPack.Pack_name, &updatedPack.Pack_description, &updatedPack.Sharing_code, &updatedPack.Created_at, &updatedPack.Updated_at)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Compare the data in DB with Test dataset
		switch {
		case updatedPack.User_id != TestUpdatedPack.User_id:
			t.Errorf("Expected User ID %v but got %v", TestUpdatedPack.User_id, updatedPack.User_id)
		case updatedPack.Pack_name != TestUpdatedPack.Pack_name:
			t.Errorf("Expected Pack Name %v but got %v", TestUpdatedPack.Pack_name, updatedPack.Pack_name)
		case updatedPack.Pack_description != TestUpdatedPack.Pack_description:
			t.Errorf("Expected Pack Description %v but got %v", TestUpdatedPack.Pack_description, updatedPack.Pack_description)
		}
	})
}

func TestDeletePackByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for DeletePackByID handler
	router.DELETE("/packs/:id", DeletePackByID)

	t.Run("Delete pack", func(t *testing.T) {

		// Set up a test scenario: sending a DELETE request
		path := fmt.Sprintf("/packs/%d", packs[2].ID)
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

		// Check in the database if the pack has been deleted
		var packName string
		row := database.Db().QueryRow("SELECT pack_name FROM pack WHERE id = $1;", packs[2].ID)
		err = row.Scan(&packName)
		if err == nil {
			t.Errorf("Pack ID 3 associated to pack name %s should be deleted and it is still in DB", packName)
		} else if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("Failed to create request: %v", err)
		}
	})
}

func TestGetPackContents(t *testing.T) {
	var packContents dataset.PackContents
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for GetPackContents handler
	router.GET("/packcontents", GetPackContents)

	t.Run("PackContent List Retrieved", func(t *testing.T) {
		// Create a mock HTTP request to the /packs endpoint
		req, err := http.NewRequest("GET", "/packcontents", nil)
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

		// Unmarshal the response body into a slice of PackContent struct
		if err := json.Unmarshal(w.Body.Bytes(), &packContents); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}
		// determine if the packcontent - and only the expected packcontent - is in the database
		if len(packContents) < 4 {
			t.Errorf("Expected almost 4 packcontent but got %d", len(packContents))
		} else {
			switch {
			case !cmp.Equal(packContents[0].Pack_id, packItems[0].Pack_id):
				t.Errorf("Expected Pack ID %v but got %v", packItems[0].Pack_id, packContents[0].Pack_id)
			case !cmp.Equal(packContents[0].Item_id, packItems[0].Item_id):
				t.Errorf("Expected Item ID %v but got %v", packItems[0].Item_id, packContents[0].Item_id)
			case !cmp.Equal(packContents[0].Quantity, packItems[0].Quantity):
				t.Errorf("Expected Quantity %v but got %v", packItems[0].Quantity, packContents[0].Quantity)
			case !cmp.Equal(packContents[0].Worn, packItems[0].Worn):
				t.Errorf("Expected Worn %v but got %v", packItems[0].Worn, packContents[0].Worn)
			case !cmp.Equal(packContents[0].Consumable, packItems[0].Consumable):
				t.Errorf("Expected Consumable %v but got %v", packItems[0].Consumable, packContents[0].Consumable)
			case !cmp.Equal(packContents[1].Pack_id, packItems[1].Pack_id):
				t.Errorf("Expected Pack ID %v but got %v", packItems[1].Pack_id, packContents[1].Pack_id)
			case !cmp.Equal(packContents[1].Item_id, packItems[1].Item_id):
				t.Errorf("Expected Item ID %v but got %v", packItems[1].Item_id, packContents[1].Item_id)
			case !cmp.Equal(packContents[1].Quantity, packItems[1].Quantity):
				t.Errorf("Expected Quantity %v but got %v", packItems[1].Quantity, packContents[1].Quantity)
			case !cmp.Equal(packContents[1].Worn, packItems[1].Worn):
				t.Errorf("Expected Worn %v but got %v", packItems[1].Worn, packContents[1].Worn)
			case !cmp.Equal(packContents[1].Consumable, packItems[1].Consumable):
				t.Errorf("Expected Consumable %v but got %v", packItems[1].Consumable, packContents[1].Consumable)
			case !cmp.Equal(packContents[2].Pack_id, packItems[2].Pack_id):
				t.Errorf("Expected Pack ID %v but got %v", packItems[2].Pack_id, packContents[2].Pack_id)
			case !cmp.Equal(packContents[2].Item_id, packItems[2].Item_id):
				t.Errorf("Expected Item ID %v but got %v", packItems[2].Item_id, packContents[2].Item_id)
			case !cmp.Equal(packContents[2].Quantity, packItems[2].Quantity):
				t.Errorf("Expected Quantity %v but got %v", packItems[2].Quantity, packContents[2].Quantity)
			case !cmp.Equal(packContents[2].Worn, packItems[2].Worn):
				t.Errorf("Expected Worn %v but got %v", packItems[2].Worn, packContents[2].Worn)
			case !cmp.Equal(packContents[2].Consumable, packItems[2].Consumable):
				t.Errorf("Expected Consumable %v but got %v", packItems[2].Consumable, packContents[2].Consumable)
			}
		}
	})
}

func TestGetPackContentByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for GetPackByID handler
	router.GET("/packcontents/:id", GetPackContentByID)

	// Set up a test scenario: PackContent found
	t.Run("PackContent Found", func(t *testing.T) {
		path := fmt.Sprintf("/packcontents/%d", packItems[0].ID)
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d but got %d", http.StatusOK, w.Code)
		}

		// Unmarshal the response body into a PackContent struct
		var receivedPackContent dataset.PackContent
		if err := json.Unmarshal(w.Body.Bytes(), &receivedPackContent); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received PackContent with the expected PackContent
		switch {
		case receivedPackContent.Pack_id != packItems[0].Pack_id:
			t.Errorf("Expected Pack ID %v but got %v", packItems[0].Pack_id, receivedPackContent.Pack_id)
		case receivedPackContent.Item_id != packItems[0].Item_id:
			t.Errorf("Expected Item ID %v but got %v", packItems[0].Item_id, receivedPackContent.Item_id)
		case receivedPackContent.Quantity != packItems[0].Quantity:
			t.Errorf("Expected Quantity %v but got %v", packItems[0].Quantity, receivedPackContent.Quantity)
		case receivedPackContent.Worn != packItems[0].Worn:
			t.Errorf("Expected Worn %v but got %v", packItems[0].Worn, receivedPackContent.Worn)
		case receivedPackContent.Consumable != packItems[0].Consumable:
			t.Errorf("Expected Consumable %v but got %v", packItems[0].Consumable, receivedPackContent.Consumable)
		}
	})

	// Set up a test scenario: PackContent not found
	t.Run("PackContent Not Found", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/packcontents/1000", nil)
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

func TestPostPackContent(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for PostPackContents handler
	router.POST("/packcontents", PostPackContent)

	// Sample pack content data
	newPackContent := dataset.PackContent{
		Pack_id:    packs[1].ID,
		Item_id:    inventories_user_pack1[2].ID,
		Quantity:   10,
		Worn:       false,
		Consumable: false,
	}

	// Convert pack content data to JSON
	jsonData, err := json.Marshal(newPackContent)
	if err != nil {
		t.Fatalf("Failed to marshal pack content data: %v", err)
	}

	t.Run("Insert pack content", func(t *testing.T) {

		// Set up a test scenario: sending a POST request with JSON data
		req, err := http.NewRequest("POST", "/packcontents", bytes.NewBuffer(jsonData))
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

		// Query the database to get the inserted pack content
		var insertedPackContent dataset.PackContent
		row := database.Db().QueryRow("SELECT id, pack_id, item_id, quantity, worn, consumable, created_at, updated_at FROM pack_content WHERE pack_id = $1 AND item_id = $2;", packs[1].ID, newPackContent.Item_id)
		err = row.Scan(&insertedPackContent.ID, &insertedPackContent.Pack_id, &insertedPackContent.Item_id, &insertedPackContent.Quantity, &insertedPackContent.Worn, &insertedPackContent.Consumable, &insertedPackContent.Created_at, &insertedPackContent.Updated_at)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Unmarshal the response body into a pack content struct
		var receivedPackContent dataset.PackContent
		if err := json.Unmarshal(w.Body.Bytes(), &receivedPackContent); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Compare the received pack content with the expected pack content data
		switch {
		case receivedPackContent.Pack_id != insertedPackContent.Pack_id:
			t.Errorf("Expected Pack ID %v but got %v", insertedPackContent.Pack_id, receivedPackContent.Pack_id)
		case receivedPackContent.Item_id != insertedPackContent.Item_id:
			t.Errorf("Expected Item ID %v but got %v", insertedPackContent.Item_id, receivedPackContent.Item_id)
		case receivedPackContent.Quantity != insertedPackContent.Quantity:
			t.Errorf("Expected Quantity %v but got %v", insertedPackContent.Quantity, receivedPackContent.Quantity)
		case receivedPackContent.Worn != insertedPackContent.Worn:
			t.Errorf("Expected Worn %v but got %v", insertedPackContent.Worn, receivedPackContent.Worn)
		case receivedPackContent.Consumable != insertedPackContent.Consumable:
			t.Errorf("Expected Consumable %v but got %v", insertedPackContent.Consumable, receivedPackContent.Consumable)

		}
	})
}

func TestPutPackContentByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for PutPackContents handler
	router.PUT("/packcontents/:id", PutPackContentByID)

	// Sample pack content data
	TestUpdatedPackContent := dataset.PackContent{
		Pack_id:    packs[1].ID,
		Item_id:    packItems[2].Item_id,
		Quantity:   10,
		Worn:       false,
		Consumable: false,
	}

	// Convert pack content data to JSON
	jsonData, err := json.Marshal(TestUpdatedPackContent)
	if err != nil {
		t.Fatalf("Failed to marshal pack content data: %v", err)
	}

	t.Run("Update pack content", func(t *testing.T) {

		// Set up a test scenario: sending a PUT request with JSON data
		path := fmt.Sprintf("/packcontents/%d", packItems[2].ID)
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

		// Query the database to get the updated pack content
		var updatedPackContent dataset.PackContent
		row := database.Db().QueryRow("SELECT id, pack_id, item_id, quantity, worn, consumable, created_at, updated_at FROM pack_content WHERE id = $1;", packItems[2].ID)
		err = row.Scan(&updatedPackContent.ID, &updatedPackContent.Pack_id, &updatedPackContent.Item_id, &updatedPackContent.Quantity, &updatedPackContent.Worn, &updatedPackContent.Consumable, &updatedPackContent.Created_at, &updatedPackContent.Updated_at)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Compare the data in DB with Test dataset
		switch {
		case updatedPackContent.Pack_id != TestUpdatedPackContent.Pack_id:
			t.Errorf("Expected Pack ID %v but got %v", TestUpdatedPackContent.Pack_id, updatedPackContent.Pack_id)
		case updatedPackContent.Item_id != TestUpdatedPackContent.Item_id:
			t.Errorf("Expected Item ID %v but got %v", TestUpdatedPackContent.Item_id, updatedPackContent.Item_id)
		case updatedPackContent.Quantity != TestUpdatedPackContent.Quantity:
			t.Errorf("Expected Quantity %v but got %v", TestUpdatedPackContent.Quantity, updatedPackContent.Quantity)
		case updatedPackContent.Worn != TestUpdatedPackContent.Worn:
			t.Errorf("Expected Worn %v but got %v", TestUpdatedPackContent.Worn, updatedPackContent.Worn)
		case updatedPackContent.Consumable != TestUpdatedPackContent.Consumable:
			t.Errorf("Expected Consumable %v but got %v", TestUpdatedPackContent.Consumable, updatedPackContent.Consumable)
		}
	})
}

func TestDeletePackContentByID(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for DeletePackByID handler
	router.DELETE("/packscontent/:id", DeletePackContentByID)

	t.Run("Delete pack Item", func(t *testing.T) {

		// Set up a test scenario: sending a DELETE request
		path := fmt.Sprintf("/packscontent/%d", packItems[2].ID)
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

		// Check in the database if the pack has been deleted
		var pack_id int
		row := database.Db().QueryRow("SELECT pack_id FROM pack_content WHERE id = $1;", packItems[2].ID)
		err = row.Scan(&pack_id)
		if err == nil {
			t.Errorf("Pack Item ID 3 associated to pack content id %d should be deleted and it is still in DB", pack_id)
		} else if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("Failed to create request: %v", err)
		}
	})
}

func TestGetPackContentsByPackID(t *testing.T) {
	var packContentWithItems dataset.PackContentWithItems
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for DeletePackByID handler
	router.GET("/packs/:id/packcontents", GetPackContentsByPackID)

	t.Run("Retrieve fourth pack", func(t *testing.T) {

		path := fmt.Sprintf("/packs/%d/packcontents", packs[3].ID)
		req, err := http.NewRequest("GET", path, nil)
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

		// Unmarshal the response body into a slice of packs struct
		if err := json.Unmarshal(w.Body.Bytes(), &packContentWithItems); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// determine if the answer is correct
		if len(packContentWithItems) != len(packWithItems) {
			t.Errorf("Expected same number of items in the pack but got %d instead of %d", len(packContentWithItems), len(packWithItems))
		}
		switch {
		case packContentWithItems[0].Pack_id != packWithItems[0].Pack_id:
			t.Errorf("Expected Pack ID %v but got %v", packWithItems[0].Pack_id, packContentWithItems[0].Pack_id)
		case packContentWithItems[0].Item_name != packWithItems[0].Item_name:
			t.Errorf("Expected Item Name %v but got %v", packWithItems[0].Item_name, packContentWithItems[0].Item_name)
		case packContentWithItems[0].Category != packWithItems[0].Category:
			t.Errorf("Expected Category %v but got %v", packWithItems[0].Category, packContentWithItems[0].Category)
		case packContentWithItems[0].Item_description != packWithItems[0].Item_description:
			t.Errorf("Expected Item Description %v but got %v", packWithItems[0].Item_description, packContentWithItems[0].Item_description)
		case packContentWithItems[0].Weight != packWithItems[0].Weight:
			t.Errorf("Expected Weight %v but got %v", packWithItems[0].Weight, packContentWithItems[0].Weight)
		case packContentWithItems[0].Weight_unit != packWithItems[0].Weight_unit:
			t.Errorf("Expected Weight Unit %v but got %v", packWithItems[0].Weight_unit, packContentWithItems[0].Weight_unit)
		case packContentWithItems[0].Item_url != packWithItems[0].Item_url:
			t.Errorf("Expected Item URL %v but got %v", packWithItems[0].Item_url, packContentWithItems[0].Item_url)
		case packContentWithItems[0].Price != packWithItems[0].Price:
			t.Errorf("Expected Price %v but got %v", packWithItems[0].Price, packContentWithItems[0].Price)
		case packContentWithItems[0].Currency != packWithItems[0].Currency:
			t.Errorf("Expected Currency %v but got %v", packWithItems[0].Currency, packContentWithItems[0].Currency)
		case packContentWithItems[0].Quantity != packWithItems[0].Quantity:
			t.Errorf("Expected Quantity %v but got %v", packWithItems[0].Quantity, packContentWithItems[0].Quantity)
		case packContentWithItems[0].Worn != packWithItems[0].Worn:
			t.Errorf("Expected Worn %v but got %v", packWithItems[0].Worn, packContentWithItems[0].Worn)
		case packContentWithItems[0].Consumable != packWithItems[0].Consumable:
			t.Errorf("Expected Consumable %v but got %v", packWithItems[0].Consumable, packContentWithItems[0].Consumable)
		}
	})

	t.Run("Pack Not Found", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/packs/1000/packcontents", nil)
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

func TestImportFromLighterPack(t *testing.T) {
	// Read the CSV file
	csvData := `Item Name,Category,desc,qty,weight,unit,url,price,worn,consumable
item1,category1,description1,1,100,g,http://example.com,10,worn,consumable
item2,category2,description2,2,150,g,http://example2.com,20,,consumable`

	tests := []struct {
		name         string
		fileContents string
		expectedCode int
	}{
		{
			name:         "Valid CSV",
			fileContents: "Item Name,Category,desc,qty,weight,unit,url,price,worn,consumable\nitem1,category1,description1,1,100,g,http://example.com,10,worn,consumable",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid CSV File",
			fileContents: "some plain text",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "CSV Data from File",
			fileContents: csvData,
			expectedCode: http.StatusOK,
		},
	}

	token, err := security.GenerateToken(1)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for DeletePackByID handler
	router.POST("/importfromlighterpack", ImportFromLighterPack)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBuf := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuf)

			// Create a form file part with the CSV content
			fileWriter, err := bodyWriter.CreateFormFile("file", "test.csv")
			if err != nil {
				t.Fatalf("Failed to create form file: %v", err)
			}

			if _, err = fileWriter.Write([]byte(tt.fileContents)); err != nil {
				t.Fatalf("Failed to write file contents: %v", err)
			}

			contentType := bodyWriter.FormDataContentType()
			bodyWriter.Close()

			req, err := http.NewRequest(http.MethodPost, "/importfromlighterpack", bodyBuf)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", contentType)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status code %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}

func TestCheckPackOwnership(t *testing.T) {
	testCases := []struct {
		packID   uint
		userID   uint
		expected bool
		name     string
	}{
		{
			packID:   packs[3].ID,
			userID:   packs[3].User_id,
			expected: true,
			name:     "Owner checks their own pack",
		},
		{
			packID:   packs[2].ID,
			userID:   packs[3].User_id,
			expected: false,
			name:     "Non-owner checks someone else's pack",
		},
	}

	// Loop through each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function under test
			test, err := checkPackOwnership(tc.packID, tc.userID)
			if err != nil {
				t.Fatalf("Failed to check pack ownership: %v", err)
			}
			if test != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, test)
			}
		})
	}
}

func TestFindPackIdBySharingCode(t *testing.T) {
	testCases := []struct {
		sharingCode string
		expected    uint
		name        string
	}{
		{
			sharingCode: packs[3].Sharing_code,
			expected:    packs[3].ID,
			name:        "Valid sharing code",
		},
		{
			sharingCode: "invalid",
			expected:    0,
			name:        "Invalid sharing code",
		},
	}

	// Loop through each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function under test
			test, err := findPackIdBySharingCode(tc.sharingCode)
			if err != nil {
				t.Fatalf("Failed to find pack ID by sharing code: %v", err)
			}
			if test != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, test)
			}
		})
	}
}

func TestGetPackBySharingCode(t *testing.T) {
	testCases := []struct {
		sharingCode  string
		responseCode int
		expected     dataset.PackContents
		name         string
	}{
		{
			sharingCode:  packs[0].Sharing_code,
			responseCode: http.StatusOK,
			expected:     packItems[0:1],
			name:         "Valid sharing code",
		},
		{
			sharingCode:  "invalid",
			responseCode: http.StatusNotFound,
			expected:     dataset.PackContents{},
			name:         "Invalid sharing code",
		},
	}
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router instance
	router := gin.Default()

	// Define the endpoint for DeletePackByID handler
	router.GET("/sharedlist/:sharing_code", SharedList)

	// Loop through each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var returnedPackContents dataset.PackContents
			// Set up a test scenario: sending a GET request
			path := fmt.Sprintf("/sharedlist/%s", tc.sharingCode)
			req, err := http.NewRequest("GET", path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.responseCode {
				t.Errorf("Expected status code %d but got %d", tc.responseCode, w.Code)
			}

			if tc.responseCode == http.StatusOK {
				// Unmarshal the response body into a slice of packs struct
				if err := json.Unmarshal(w.Body.Bytes(), &returnedPackContents); err != nil {
					t.Fatalf("Failed to unmarshal response body: %v", err)
				}

				// determine if the answer is correct
				if reflect.DeepEqual(returnedPackContents, tc.expected) {
					t.Errorf("Expected %v but got %v", tc.expected, returnedPackContents)
				}
			}
		})
	}
}

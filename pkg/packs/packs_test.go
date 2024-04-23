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
	err = database.Initialization()
	if err != nil {
		log.Fatalf("Error connecting database : %v", err)
	}
	println("Database connected")

	// init DB migration
	err = database.Migrate()
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
		req, err := http.NewRequest(http.MethodGet, "/packs", nil)
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
			case !cmp.Equal(getPacks[0].UserID, packs[0].UserID):
				t.Errorf("Expected User ID %v but got %v", packs[0].UserID, getPacks[0].UserID)
			case !cmp.Equal(getPacks[0].PackName, packs[0].PackName):
				t.Errorf("Expected Pack Name %v but got %v", packs[0].PackName, getPacks[0].PackName)
			case !cmp.Equal(getPacks[0].PackDescription, packs[0].PackDescription):
				t.Errorf("Expected Pack Description %v but got %v", packs[0].PackDescription,
					getPacks[0].PackDescription)
			case !cmp.Equal(getPacks[0].SharingCode, packs[0].SharingCode):
				t.Errorf("Expected Sharing Code %v but got %v", packs[0].SharingCode, getPacks[0].SharingCode)
			case !cmp.Equal(getPacks[1].UserID, packs[1].UserID):
				t.Errorf("Expected User ID %v but got %v", packs[1].UserID, getPacks[1].UserID)
			case !cmp.Equal(getPacks[1].PackName, packs[1].PackName):
				t.Errorf("Expected Pack Name %v but got %v", packs[1].PackName, getPacks[1].PackName)
			case !cmp.Equal(getPacks[1].PackDescription, packs[1].PackDescription):
				t.Errorf("Expected Pack Description %v but got %v", packs[1].PackDescription,
					getPacks[1].PackDescription)
			case !cmp.Equal(getPacks[1].SharingCode, packs[1].SharingCode):
				t.Errorf("Expected Sharing Code %v but got %v", packs[1].SharingCode, getPacks[1].SharingCode)
			case !cmp.Equal(getPacks[2].UserID, packs[2].UserID):
				t.Errorf("Expected User ID %v but got %v", packs[2].UserID, getPacks[2].UserID)
			case !cmp.Equal(getPacks[2].PackName, packs[2].PackName):
				t.Errorf("Expected Pack Name %v but got %v", packs[2].PackName, getPacks[2].PackName)
			case !cmp.Equal(getPacks[2].PackDescription, packs[2].PackDescription):
				t.Errorf("Expected Pack Description %v but got %v", packs[2].PackDescription,
					getPacks[2].PackDescription)
			case !cmp.Equal(getPacks[2].SharingCode, packs[2].SharingCode):
				t.Errorf("Expected Sharing Code %v but got %v", packs[2].SharingCode, getPacks[2].SharingCode)
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
		req, err := http.NewRequest(http.MethodGet, path, nil)
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
		case receivedPack.UserID != packs[0].UserID:
			t.Errorf("Expected User ID %v but got %v", packs[0].UserID, receivedPack.UserID)
		case receivedPack.PackName != packs[0].PackName:
			t.Errorf("Expected Pack Name %v but got %v", packs[0].PackName, receivedPack.PackName)
		case receivedPack.PackDescription != packs[0].PackDescription:
			t.Errorf("Expected Pack Description %v but got %v", packs[0].PackDescription, receivedPack.PackDescription)
		case receivedPack.SharingCode != packs[0].SharingCode:
			t.Errorf("Expected Sharing Code %v but got %v", packs[0].SharingCode, receivedPack.SharingCode)
		}
	})

	// Set up a test scenario: pack not found
	t.Run("Pack Not Found", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/packs/1000", nil)
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
		UserID:          users[0].ID,
		PackName:        "SomePack",
		PackDescription: "This is a new pack",
	}

	// Convert pack data to JSON
	jsonData, err := json.Marshal(newPack)
	if err != nil {
		t.Fatalf("Failed to marshal pack data: %v", err)
	}

	t.Run("Insert pack", func(t *testing.T) {
		// Set up a test scenario: sending a POST request with JSON data
		req, err := http.NewRequest(http.MethodPost, "/packs", bytes.NewBuffer(jsonData))
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
		row := database.DB().QueryRow(
			`SELECT id, user_id, pack_name, pack_description, sharing_code, created_at, updated_at 
			FROM pack 
			WHERE pack_name = $1;`,
			newPack.PackName)
		err = row.Scan(
			&insertedPack.ID,
			&insertedPack.UserID,
			&insertedPack.PackName,
			&insertedPack.PackDescription,
			&insertedPack.SharingCode,
			&insertedPack.CreatedAt,
			&insertedPack.UpdatedAt)
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
		case receivedPack.UserID != insertedPack.UserID:
			t.Errorf("Expected User ID %v but got %v", insertedPack.UserID, receivedPack.UserID)
		case receivedPack.PackName != insertedPack.PackName:
			t.Errorf("Expected Pack Name %v but got %v", insertedPack.PackName, receivedPack.PackName)
		case receivedPack.PackDescription != insertedPack.PackDescription:
			t.Errorf("Expected Pack Description %v but got %v", insertedPack.PackDescription,
				receivedPack.PackDescription)
		case receivedPack.SharingCode == "":
			t.Errorf("Expected a non empty Sharing Code but got %v", receivedPack.SharingCode)
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
	testUpdatedPack := dataset.Pack{
		UserID:          users[1].ID,
		PackName:        "Amazing Pack",
		PackDescription: "Updated pack description",
	}

	// Convert pack data to JSON
	jsonData, err := json.Marshal(testUpdatedPack)
	if err != nil {
		t.Fatalf("Failed to marshal pack data: %v", err)
	}

	t.Run("Update pack", func(t *testing.T) {
		// Set up a test scenario: sending a PUT request with JSON data
		path := fmt.Sprintf("/packs/%d", packs[2].ID)
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

		// Query the database to get the updated pack
		var updatedPack dataset.Pack
		row := database.DB().QueryRow(
			`SELECT id, user_id, pack_name, pack_description, sharing_code, created_at, updated_at 
			FROM pack 
			WHERE id = $1;`,
			packs[2].ID)
		err = row.Scan(
			&updatedPack.ID,
			&updatedPack.UserID,
			&updatedPack.PackName,
			&updatedPack.PackDescription,
			&updatedPack.SharingCode,
			&updatedPack.CreatedAt,
			&updatedPack.UpdatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Compare the data in DB with Test dataset
		switch {
		case updatedPack.UserID != testUpdatedPack.UserID:
			t.Errorf("Expected User ID %v but got %v", testUpdatedPack.UserID, updatedPack.UserID)
		case updatedPack.PackName != testUpdatedPack.PackName:
			t.Errorf("Expected Pack Name %v but got %v", testUpdatedPack.PackName, updatedPack.PackName)
		case updatedPack.PackDescription != testUpdatedPack.PackDescription:
			t.Errorf("Expected Pack Description %v but got %v", testUpdatedPack.PackDescription,
				updatedPack.PackDescription)
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

		// Check in the database if the pack has been deleted
		var packName string
		row := database.DB().QueryRow("SELECT pack_name FROM pack WHERE id = $1;", packs[2].ID)
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
		req, err := http.NewRequest(http.MethodGet, "/packcontents", nil)
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
			case !cmp.Equal(packContents[0].PackID, packItems[0].PackID):
				t.Errorf("Expected Pack ID %v but got %v", packItems[0].PackID, packContents[0].PackID)
			case !cmp.Equal(packContents[0].ItemID, packItems[0].ItemID):
				t.Errorf("Expected Item ID %v but got %v", packItems[0].ItemID, packContents[0].ItemID)
			case !cmp.Equal(packContents[0].Quantity, packItems[0].Quantity):
				t.Errorf("Expected Quantity %v but got %v", packItems[0].Quantity, packContents[0].Quantity)
			case !cmp.Equal(packContents[0].Worn, packItems[0].Worn):
				t.Errorf("Expected Worn %v but got %v", packItems[0].Worn, packContents[0].Worn)
			case !cmp.Equal(packContents[0].Consumable, packItems[0].Consumable):
				t.Errorf("Expected Consumable %v but got %v", packItems[0].Consumable, packContents[0].Consumable)
			case !cmp.Equal(packContents[1].PackID, packItems[1].PackID):
				t.Errorf("Expected Pack ID %v but got %v", packItems[1].PackID, packContents[1].PackID)
			case !cmp.Equal(packContents[1].ItemID, packItems[1].ItemID):
				t.Errorf("Expected Item ID %v but got %v", packItems[1].ItemID, packContents[1].ItemID)
			case !cmp.Equal(packContents[1].Quantity, packItems[1].Quantity):
				t.Errorf("Expected Quantity %v but got %v", packItems[1].Quantity, packContents[1].Quantity)
			case !cmp.Equal(packContents[1].Worn, packItems[1].Worn):
				t.Errorf("Expected Worn %v but got %v", packItems[1].Worn, packContents[1].Worn)
			case !cmp.Equal(packContents[1].Consumable, packItems[1].Consumable):
				t.Errorf("Expected Consumable %v but got %v", packItems[1].Consumable, packContents[1].Consumable)
			case !cmp.Equal(packContents[2].PackID, packItems[2].PackID):
				t.Errorf("Expected Pack ID %v but got %v", packItems[2].PackID, packContents[2].PackID)
			case !cmp.Equal(packContents[2].ItemID, packItems[2].ItemID):
				t.Errorf("Expected Item ID %v but got %v", packItems[2].ItemID, packContents[2].ItemID)
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
		req, err := http.NewRequest(http.MethodGet, path, nil)
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
		case receivedPackContent.PackID != packItems[0].PackID:
			t.Errorf("Expected Pack ID %v but got %v", packItems[0].PackID, receivedPackContent.PackID)
		case receivedPackContent.ItemID != packItems[0].ItemID:
			t.Errorf("Expected Item ID %v but got %v", packItems[0].ItemID, receivedPackContent.ItemID)
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
		req, err := http.NewRequest(http.MethodGet, "/packcontents/1000", nil)
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
		PackID:     packs[1].ID,
		ItemID:     inventoriesUserPack1[2].ID,
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
		req, err := http.NewRequest(http.MethodPost, "/packcontents", bytes.NewBuffer(jsonData))
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
		row := database.DB().QueryRow(
			`SELECT id, pack_id, item_id, quantity, worn, consumable, created_at, updated_at 
			FROM pack_content 
			WHERE pack_id = $1 AND item_id = $2;`,
			packs[1].ID, newPackContent.ItemID)
		err = row.Scan(
			&insertedPackContent.ID,
			&insertedPackContent.PackID,
			&insertedPackContent.ItemID,
			&insertedPackContent.Quantity,
			&insertedPackContent.Worn,
			&insertedPackContent.Consumable,
			&insertedPackContent.CreatedAt,
			&insertedPackContent.UpdatedAt)
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
		case receivedPackContent.PackID != insertedPackContent.PackID:
			t.Errorf("Expected Pack ID %v but got %v", insertedPackContent.PackID, receivedPackContent.PackID)
		case receivedPackContent.ItemID != insertedPackContent.ItemID:
			t.Errorf("Expected Item ID %v but got %v", insertedPackContent.ItemID, receivedPackContent.ItemID)
		case receivedPackContent.Quantity != insertedPackContent.Quantity:
			t.Errorf("Expected Quantity %v but got %v", insertedPackContent.Quantity, receivedPackContent.Quantity)
		case receivedPackContent.Worn != insertedPackContent.Worn:
			t.Errorf("Expected Worn %v but got %v", insertedPackContent.Worn, receivedPackContent.Worn)
		case receivedPackContent.Consumable != insertedPackContent.Consumable:
			t.Errorf("Expected Consumable %v but got %v", insertedPackContent.Consumable,
				receivedPackContent.Consumable)
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
	testUpdatedPackContent := dataset.PackContent{
		PackID:     packs[1].ID,
		ItemID:     packItems[2].ItemID,
		Quantity:   10,
		Worn:       false,
		Consumable: false,
	}

	// Convert pack content data to JSON
	jsonData, err := json.Marshal(testUpdatedPackContent)
	if err != nil {
		t.Fatalf("Failed to marshal pack content data: %v", err)
	}

	t.Run("Update pack content", func(t *testing.T) {
		// Set up a test scenario: sending a PUT request with JSON data
		path := fmt.Sprintf("/packcontents/%d", packItems[2].ID)
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

		// Query the database to get the updated pack content
		var updatedPackContent dataset.PackContent
		row := database.DB().QueryRow(
			`SELECT id, pack_id, item_id, quantity, worn, consumable, created_at, updated_at 
			FROM pack_content 
			WHERE id = $1;`,
			packItems[2].ID)
		err = row.Scan(
			&updatedPackContent.ID,
			&updatedPackContent.PackID,
			&updatedPackContent.ItemID,
			&updatedPackContent.Quantity,
			&updatedPackContent.Worn,
			&updatedPackContent.Consumable,
			&updatedPackContent.CreatedAt,
			&updatedPackContent.UpdatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No rows were returned!")
			}
			t.Fatalf("Failed to run request: %v", err)
		}

		// Compare the data in DB with Test dataset
		switch {
		case updatedPackContent.PackID != testUpdatedPackContent.PackID:
			t.Errorf("Expected Pack ID %v but got %v", testUpdatedPackContent.PackID, updatedPackContent.PackID)
		case updatedPackContent.ItemID != testUpdatedPackContent.ItemID:
			t.Errorf("Expected Item ID %v but got %v", testUpdatedPackContent.ItemID, updatedPackContent.ItemID)
		case updatedPackContent.Quantity != testUpdatedPackContent.Quantity:
			t.Errorf("Expected Quantity %v but got %v", testUpdatedPackContent.Quantity, updatedPackContent.Quantity)
		case updatedPackContent.Worn != testUpdatedPackContent.Worn:
			t.Errorf("Expected Worn %v but got %v", testUpdatedPackContent.Worn, updatedPackContent.Worn)
		case updatedPackContent.Consumable != testUpdatedPackContent.Consumable:
			t.Errorf("Expected Consumable %v but got %v", testUpdatedPackContent.Consumable,
				updatedPackContent.Consumable)
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

		// Check in the database if the pack has been deleted
		var packID int
		row := database.DB().QueryRow("SELECT pack_id FROM pack_content WHERE id = $1;", packItems[2].ID)
		err = row.Scan(&packID)
		if err == nil {
			t.Errorf("Pack Item ID 3 associated to pack content id %d should be deleted and it is still in DB", packID)
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
		req, err := http.NewRequest(http.MethodGet, path, nil)
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
			t.Errorf("Expected same number of items in the pack but got %d instead of %d", len(packContentWithItems),
				len(packWithItems))
		}
		switch {
		case packContentWithItems[0].PackID != packWithItems[0].PackID:
			t.Errorf("Expected Pack ID %v but got %v", packWithItems[0].PackID, packContentWithItems[0].PackID)
		case packContentWithItems[0].ItemName != packWithItems[0].ItemName:
			t.Errorf("Expected Item Name %v but got %v", packWithItems[0].ItemName, packContentWithItems[0].ItemName)
		case packContentWithItems[0].Category != packWithItems[0].Category:
			t.Errorf("Expected Category %v but got %v", packWithItems[0].Category, packContentWithItems[0].Category)
		case packContentWithItems[0].ItemDescription != packWithItems[0].ItemDescription:
			t.Errorf("Expected Item Description %v but got %v", packWithItems[0].ItemDescription,
				packContentWithItems[0].ItemDescription)
		case packContentWithItems[0].Weight != packWithItems[0].Weight:
			t.Errorf("Expected Weight %v but got %v", packWithItems[0].Weight, packContentWithItems[0].Weight)
		case packContentWithItems[0].WeightUnit != packWithItems[0].WeightUnit:
			t.Errorf("Expected Weight Unit %v but got %v", packWithItems[0].WeightUnit,
				packContentWithItems[0].WeightUnit)
		case packContentWithItems[0].ItemURL != packWithItems[0].ItemURL:
			t.Errorf("Expected Item URL %v but got %v", packWithItems[0].ItemURL, packContentWithItems[0].ItemURL)
		case packContentWithItems[0].Price != packWithItems[0].Price:
			t.Errorf("Expected Price %v but got %v", packWithItems[0].Price, packContentWithItems[0].Price)
		case packContentWithItems[0].Currency != packWithItems[0].Currency:
			t.Errorf("Expected Currency %v but got %v", packWithItems[0].Currency, packContentWithItems[0].Currency)
		case packContentWithItems[0].Quantity != packWithItems[0].Quantity:
			t.Errorf("Expected Quantity %v but got %v", packWithItems[0].Quantity, packContentWithItems[0].Quantity)
		case packContentWithItems[0].Worn != packWithItems[0].Worn:
			t.Errorf("Expected Worn %v but got %v", packWithItems[0].Worn, packContentWithItems[0].Worn)
		case packContentWithItems[0].Consumable != packWithItems[0].Consumable:
			t.Errorf("Expected Consumable %v but got %v", packWithItems[0].Consumable,
				packContentWithItems[0].Consumable)
		}
	})

	t.Run("Pack Not Found", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/packs/1000/packcontents", nil)
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
			name: "Valid CSV",
			fileContents: "Item Name,Category,desc,qty,weight,unit,url,price,worn,consumable\nitem1,category1," +
				"description1,1,100,g,http://example.com,10,worn,consumable",
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
			userID:   packs[3].UserID,
			expected: true,
			name:     "Owner checks their own pack",
		},
		{
			packID:   packs[2].ID,
			userID:   packs[3].UserID,
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

func TestFindPackIDBySharingCode(t *testing.T) {
	testCases := []struct {
		sharingCode string
		expected    uint
		name        string
	}{
		{
			sharingCode: packs[3].SharingCode,
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
			test, err := findPackIDBySharingCode(tc.sharingCode)
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
			sharingCode:  packs[0].SharingCode,
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
			path := "/sharedlist/" + tc.sharingCode
			req, err := http.NewRequest(http.MethodGet, path, nil)
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

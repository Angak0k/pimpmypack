package images

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
)

var testStorage *DBImageStorage

func TestMain(m *testing.M) {
	// Init env
	err := config.EnvInit("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file or environment variable: %v", err)
	}

	// Init DB
	err = database.Initialization()
	if err != nil {
		log.Fatalf("Error connecting database: %v", err)
	}

	// Init DB migration
	err = database.Migrate()
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	// Load test data
	err = loadImageTestData()
	if err != nil {
		log.Fatalf("Error loading test data: %v", err)
	}

	// Create testStorage instance
	testStorage = NewDBImageStorage()

	// Run tests
	ret := m.Run()

	// Cleanup test data
	if err := cleanupImageTestData(); err != nil {
		log.Printf("Warning: failed to cleanup test data: %v", err)
	}

	os.Exit(ret)
}

func TestDBImageStorage_Save(t *testing.T) {
	ctx := context.Background()
	packID := uint(999) // From testPacks[0]

	// Clean up image after test (pack cleanup handled by TestMain)
	defer func() {
		_ = testStorage.Delete(ctx, packID)
	}()

	// Test data
	testData := []byte("test image data")
	testMetadata := ImageMetadata{
		MimeType: "image/jpeg",
		FileSize: len(testData),
		Width:    800,
		Height:   600,
	}

	// Save image
	err := testStorage.Save(ctx, packID, testData, testMetadata)
	if err != nil {
		t.Fatalf("Failed to save image: %v", err)
	}

	// Verify image exists
	exists, err := testStorage.Exists(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Image should exist after save")
	}

	// Retrieve and verify
	img, err := testStorage.Get(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to get image: %v", err)
	}
	if img == nil {
		t.Fatal("Image should not be nil")
	}
	if string(img.Data) != string(testData) {
		t.Errorf("Data mismatch: expected %s, got %s", testData, img.Data)
	}
	if img.Metadata.Width != 800 {
		t.Errorf("Width mismatch: expected 800, got %d", img.Metadata.Width)
	}
}

func TestDBImageStorage_Update(t *testing.T) {
	ctx := context.Background()
	packID := uint(1000) // From testPacks[1]

	// Clean up image after test (pack cleanup handled by TestMain)
	defer func() {
		_ = testStorage.Delete(ctx, packID)
	}()

	// Save initial image
	initialData := []byte("initial data")
	initialMetadata := ImageMetadata{
		MimeType: "image/jpeg",
		FileSize: len(initialData),
		Width:    640,
		Height:   480,
	}

	err := testStorage.Save(ctx, packID, initialData, initialMetadata)
	if err != nil {
		t.Fatalf("Failed to save initial image: %v", err)
	}

	// Update image
	updatedData := []byte("updated data")
	updatedMetadata := ImageMetadata{
		MimeType: "image/jpeg",
		FileSize: len(updatedData),
		Width:    1920,
		Height:   1080,
	}

	err = testStorage.Save(ctx, packID, updatedData, updatedMetadata)
	if err != nil {
		t.Fatalf("Failed to update image: %v", err)
	}

	// Verify updated data
	img, err := testStorage.Get(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to get updated image: %v", err)
	}
	if string(img.Data) != string(updatedData) {
		t.Errorf("Data not updated: expected %s, got %s", updatedData, img.Data)
	}
	if img.Metadata.Width != 1920 {
		t.Errorf("Width not updated: expected 1920, got %d", img.Metadata.Width)
	}
}

func TestDBImageStorage_Get(t *testing.T) {
	ctx := context.Background()
	packID := uint(1001) // From testPacks[2]

	// Clean up image after test (pack cleanup handled by TestMain)
	defer func() {
		_ = testStorage.Delete(ctx, packID)
	}()

	// Test getting non-existent image
	img, err := testStorage.Get(ctx, packID)
	if err == nil {
		t.Fatal("Get should return ErrNotFound for non-existent image")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
	if img != nil {
		t.Error("Image should be nil when ErrNotFound is returned")
	}

	// Save image
	testData := []byte("test data for get")
	testMetadata := ImageMetadata{
		MimeType: "image/jpeg",
		FileSize: len(testData),
		Width:    1024,
		Height:   768,
	}

	err = testStorage.Save(ctx, packID, testData, testMetadata)
	if err != nil {
		t.Fatalf("Failed to save image: %v", err)
	}

	// Get existing image
	img, err = testStorage.Get(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to get image: %v", err)
	}
	if img == nil {
		t.Fatal("Image should not be nil")
	}
	if string(img.Data) != string(testData) {
		t.Errorf("Data mismatch: expected %s, got %s", testData, img.Data)
	}
}

func TestDBImageStorage_Delete(t *testing.T) {
	ctx := context.Background()
	packID := uint(1002) // From testPacks[3]

	// No cleanup needed - delete test verifies deletion (pack cleanup handled by TestMain)

	// Save image
	testData := []byte("test data for delete")
	testMetadata := ImageMetadata{
		MimeType: "image/jpeg",
		FileSize: len(testData),
		Width:    800,
		Height:   600,
	}

	err := testStorage.Save(ctx, packID, testData, testMetadata)
	if err != nil {
		t.Fatalf("Failed to save image: %v", err)
	}

	// Verify it exists
	exists, err := testStorage.Exists(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Image should exist before delete")
	}

	// Delete image
	err = testStorage.Delete(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to delete image: %v", err)
	}

	// Verify it doesn't exist
	exists, err = testStorage.Exists(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to check existence after delete: %v", err)
	}
	if exists {
		t.Error("Image should not exist after delete")
	}

	// Test idempotency - delete again should not error
	err = testStorage.Delete(ctx, packID)
	if err != nil {
		t.Errorf("Delete should be idempotent: %v", err)
	}
}

func TestDBImageStorage_Exists(t *testing.T) {
	ctx := context.Background()
	packID := uint(1003) // From testPacks[4]

	// Clean up image after test (pack cleanup handled by TestMain)
	defer func() {
		_ = testStorage.Delete(ctx, packID)
	}()

	// Check non-existent image
	exists, err := testStorage.Exists(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Image should not exist initially")
	}

	// Save image
	testData := []byte("test data for exists")
	testMetadata := ImageMetadata{
		MimeType: "image/jpeg",
		FileSize: len(testData),
		Width:    800,
		Height:   600,
	}

	err = testStorage.Save(ctx, packID, testData, testMetadata)
	if err != nil {
		t.Fatalf("Failed to save image: %v", err)
	}

	// Check existing image
	exists, err = testStorage.Exists(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Image should exist after save")
	}
}

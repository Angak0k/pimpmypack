package images

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
)

var storage *DBImageStorage

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

	// Create storage instance
	storage = NewDBImageStorage()

	ret := m.Run()
	os.Exit(ret)
}

func TestDBImageStorage_Save(t *testing.T) {
	ctx := context.Background()
	packID := uint(999)

	// Clean up after test
	defer func() {
		_ = storage.Delete(ctx, packID)
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
	err := storage.Save(ctx, packID, testData, testMetadata)
	if err != nil {
		t.Fatalf("Failed to save image: %v", err)
	}

	// Verify image exists
	exists, err := storage.Exists(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Image should exist after save")
	}

	// Retrieve and verify
	img, err := storage.Get(ctx, packID)
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
	packID := uint(1000)

	// Clean up after test
	defer func() {
		_ = storage.Delete(ctx, packID)
	}()

	// Save initial image
	initialData := []byte("initial data")
	initialMetadata := ImageMetadata{
		MimeType: "image/jpeg",
		FileSize: len(initialData),
		Width:    640,
		Height:   480,
	}

	err := storage.Save(ctx, packID, initialData, initialMetadata)
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

	err = storage.Save(ctx, packID, updatedData, updatedMetadata)
	if err != nil {
		t.Fatalf("Failed to update image: %v", err)
	}

	// Verify updated data
	img, err := storage.Get(ctx, packID)
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
	packID := uint(1001)

	// Clean up after test
	defer func() {
		_ = storage.Delete(ctx, packID)
	}()

	// Test getting non-existent image
	img, err := storage.Get(ctx, packID)
	if err != nil {
		t.Fatalf("Get should not error for non-existent image: %v", err)
	}
	if img != nil {
		t.Error("Image should be nil for non-existent pack")
	}

	// Save image
	testData := []byte("test data for get")
	testMetadata := ImageMetadata{
		MimeType: "image/jpeg",
		FileSize: len(testData),
		Width:    1024,
		Height:   768,
	}

	err = storage.Save(ctx, packID, testData, testMetadata)
	if err != nil {
		t.Fatalf("Failed to save image: %v", err)
	}

	// Get existing image
	img, err = storage.Get(ctx, packID)
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
	packID := uint(1002)

	// Save image
	testData := []byte("test data for delete")
	testMetadata := ImageMetadata{
		MimeType: "image/jpeg",
		FileSize: len(testData),
		Width:    800,
		Height:   600,
	}

	err := storage.Save(ctx, packID, testData, testMetadata)
	if err != nil {
		t.Fatalf("Failed to save image: %v", err)
	}

	// Verify it exists
	exists, err := storage.Exists(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Image should exist before delete")
	}

	// Delete image
	err = storage.Delete(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to delete image: %v", err)
	}

	// Verify it doesn't exist
	exists, err = storage.Exists(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to check existence after delete: %v", err)
	}
	if exists {
		t.Error("Image should not exist after delete")
	}

	// Test idempotency - delete again should not error
	err = storage.Delete(ctx, packID)
	if err != nil {
		t.Errorf("Delete should be idempotent: %v", err)
	}
}

func TestDBImageStorage_Exists(t *testing.T) {
	ctx := context.Background()
	packID := uint(1003)

	// Clean up after test
	defer func() {
		_ = storage.Delete(ctx, packID)
	}()

	// Check non-existent image
	exists, err := storage.Exists(ctx, packID)
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

	err = storage.Save(ctx, packID, testData, testMetadata)
	if err != nil {
		t.Fatalf("Failed to save image: %v", err)
	}

	// Check existing image
	exists, err = storage.Exists(ctx, packID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Image should exist after save")
	}
}

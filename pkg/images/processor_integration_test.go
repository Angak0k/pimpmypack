package images

import (
	"errors"
	"os"
	"testing"
)

// TestProcessImage_RealJPEG tests processing with a real JPEG file
func TestProcessImage_RealJPEG(t *testing.T) {
	data, err := os.ReadFile("testdata/valid.jpg")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	processed, err := ProcessImage(data)
	if err != nil {
		t.Fatalf("Failed to process valid JPEG: %v", err)
	}

	if processed.Metadata.MimeType != "image/jpeg" {
		t.Errorf("Expected mime type image/jpeg, got %s", processed.Metadata.MimeType)
	}

	if processed.Metadata.Width != 800 || processed.Metadata.Height != 600 {
		t.Errorf("Expected dimensions 800x600, got %dx%d",
			processed.Metadata.Width, processed.Metadata.Height)
	}
}

// TestProcessImage_RealPNG tests processing with a real PNG file
func TestProcessImage_RealPNG(t *testing.T) {
	data, err := os.ReadFile("testdata/valid.png")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	processed, err := ProcessImage(data)
	if err != nil {
		t.Fatalf("Failed to process valid PNG: %v", err)
	}

	if processed.Metadata.MimeType != "image/jpeg" {
		t.Errorf("Expected mime type image/jpeg (converted), got %s", processed.Metadata.MimeType)
	}

	if processed.Metadata.Width != 640 || processed.Metadata.Height != 480 {
		t.Errorf("Expected dimensions 640x480, got %dx%d",
			processed.Metadata.Width, processed.Metadata.Height)
	}
}

// TestProcessImage_RealWebP tests processing with a real WebP file
func TestProcessImage_RealWebP(t *testing.T) {
	data, err := os.ReadFile("testdata/valid.webp")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	processed, err := ProcessImage(data)
	if err != nil {
		t.Fatalf("Failed to process valid WebP: %v", err)
	}

	if processed.Metadata.MimeType != "image/jpeg" {
		t.Errorf("Expected mime type image/jpeg (converted), got %s", processed.Metadata.MimeType)
	}

	if processed.Metadata.Width != 500 || processed.Metadata.Height != 500 {
		t.Errorf("Expected dimensions 500x500, got %dx%d",
			processed.Metadata.Width, processed.Metadata.Height)
	}
}

// TestProcessImage_RealLargeImage tests resize with a real large image
func TestProcessImage_RealLargeImage(t *testing.T) {
	data, err := os.ReadFile("testdata/large.jpg")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	processed, err := ProcessImage(data)
	if err != nil {
		t.Fatalf("Failed to process large image: %v", err)
	}

	// Should be resized to MaxDimension (1920)
	if processed.Metadata.Width > MaxDimension || processed.Metadata.Height > MaxDimension {
		t.Errorf("Image not resized correctly: %dx%d (max: %d)",
			processed.Metadata.Width, processed.Metadata.Height, MaxDimension)
	}

	// Original was 3000x3000, should be resized to 1920x1920
	if processed.Metadata.Width != 1920 || processed.Metadata.Height != 1920 {
		t.Errorf("Expected resized dimensions 1920x1920, got %dx%d",
			processed.Metadata.Width, processed.Metadata.Height)
	}
}

// TestProcessImage_RealTooLarge tests file size validation with an oversized payload
func TestProcessImage_RealTooLarge(t *testing.T) {
	// Create an in-memory payload that is larger than the maximum allowed size.
	// Using a generated buffer avoids committing a large binary test asset.
	data := make([]byte, 40*1024*1024) // 40MB

	_, err := ProcessImage(data)
	if err == nil {
		t.Fatal("Expected error for too large image, got nil")
	}

	if !errors.Is(err, ErrTooLarge) {
		t.Errorf("Expected ErrTooLarge, got %v", err)
	}
}

// TestProcessImage_RealInvalidFormat tests format validation with real invalid file
func TestProcessImage_RealInvalidFormat(t *testing.T) {
	data, err := os.ReadFile("testdata/invalid.txt")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	_, err = ProcessImage(data)
	if err == nil {
		t.Fatal("Expected error for invalid format, got nil")
	}

	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("Expected ErrInvalidFormat, got %v", err)
	}
}

// TestProcessImage_RealCorrupted tests corrupted file handling
func TestProcessImage_RealCorrupted(t *testing.T) {
	data, err := os.ReadFile("testdata/corrupted.jpg")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	_, err = ProcessImage(data)
	if err == nil {
		t.Fatal("Expected error for corrupted image, got nil")
	}

	// Should get corrupted error (not invalid format, since magic bytes match)
	if !errors.Is(err, ErrCorrupted) {
		t.Errorf("Expected ErrCorrupted, got %v", err)
	}
}

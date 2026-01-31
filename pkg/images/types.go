package images

import (
	"context"
	"errors"
)

var (
	// ErrNotFound is returned when an image is not found
	ErrNotFound = errors.New("image not found")
	// ErrInvalidFormat is returned when the image format is not supported
	ErrInvalidFormat = errors.New("invalid image format")
	// ErrTooLarge is returned when the image exceeds size limits
	ErrTooLarge = errors.New("image too large")
	// ErrCorrupted is returned when the image is corrupted
	ErrCorrupted = errors.New("image corrupted")
)

// ImageMetadata contains metadata about a processed image
type ImageMetadata struct {
	MimeType string
	FileSize int
	Width    int
	Height   int
}

// ProcessedImage contains the processed image data and metadata
type ProcessedImage struct {
	Data     []byte
	Metadata ImageMetadata
}

// Image represents an image with its data and metadata
type Image struct {
	Data     []byte
	Metadata ImageMetadata
}

// ImageStorage defines the interface for image storage operations
// This allows for different storage backends (database, S3, etc.)
type ImageStorage interface {
	// Save stores an image for a pack
	Save(ctx context.Context, packID uint, data []byte, metadata ImageMetadata) error

	// Get retrieves an image for a pack
	// Returns ErrNotFound if the image doesn't exist
	Get(ctx context.Context, packID uint) (*Image, error)

	// Delete removes an image for a pack
	// Returns nil if the image doesn't exist (idempotent)
	Delete(ctx context.Context, packID uint) error

	// Exists checks if an image exists for a pack
	Exists(ctx context.Context, packID uint) (bool, error)
}

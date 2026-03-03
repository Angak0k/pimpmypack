package images

import (
	"context"
	"errors"
)

const (
	// ErrMsgNoImageProvided is the error message when no image file is provided
	ErrMsgNoImageProvided = "no image file provided"
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

// AccountImageStorage defines the interface for account profile image storage operations
type AccountImageStorage interface {
	// Save stores a profile image for an account
	Save(ctx context.Context, accountID uint, data []byte, metadata ImageMetadata) error

	// Get retrieves a profile image for an account
	// Returns ErrNotFound if the image doesn't exist
	Get(ctx context.Context, accountID uint) (*Image, error)

	// Delete removes a profile image for an account
	// Returns nil if the image doesn't exist (idempotent)
	Delete(ctx context.Context, accountID uint) error

	// Exists checks if a profile image exists for an account
	Exists(ctx context.Context, accountID uint) (bool, error)
}

// InventoryImageStorage defines the interface for inventory item image storage operations
type InventoryImageStorage interface {
	// Save stores an image for an inventory item
	Save(ctx context.Context, itemID uint, data []byte, metadata ImageMetadata) error

	// Get retrieves an image for an inventory item
	// Returns ErrNotFound if the image doesn't exist
	Get(ctx context.Context, itemID uint) (*Image, error)

	// Delete removes an image for an inventory item
	// Returns nil if the image doesn't exist (idempotent)
	Delete(ctx context.Context, itemID uint) error

	// Exists checks if an image exists for an inventory item
	Exists(ctx context.Context, itemID uint) (bool, error)
}

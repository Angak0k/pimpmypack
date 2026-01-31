package images

import (
	"context"
)

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
	// Returns nil, nil if the image doesn't exist
	Get(ctx context.Context, packID uint) (*Image, error)

	// Delete removes an image for a pack
	// Returns nil if the image doesn't exist (idempotent)
	Delete(ctx context.Context, packID uint) error

	// Exists checks if an image exists for a pack
	Exists(ctx context.Context, packID uint) (bool, error)
}

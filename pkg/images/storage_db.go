package images

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
)

// DBImageStorage implements ImageStorage using PostgreSQL database
type DBImageStorage struct{}

// NewDBImageStorage creates a new database image storage instance
func NewDBImageStorage() *DBImageStorage {
	return &DBImageStorage{}
}

// Save stores or updates an image for a pack in the database
// Uses UPSERT (INSERT ... ON CONFLICT ... DO UPDATE) for idempotent behavior
func (s *DBImageStorage) Save(ctx context.Context, packID uint, data []byte, metadata ImageMetadata) error {
	query := `
		INSERT INTO pack_images (pack_id, image_data, mime_type, file_size, width, height, uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (pack_id)
		DO UPDATE SET
			image_data = EXCLUDED.image_data,
			mime_type = EXCLUDED.mime_type,
			file_size = EXCLUDED.file_size,
			width = EXCLUDED.width,
			height = EXCLUDED.height,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := database.DB().ExecContext(
		ctx,
		query,
		packID,
		data,
		metadata.MimeType,
		metadata.FileSize,
		metadata.Width,
		metadata.Height,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}

// Get retrieves an image for a pack from the database
// Returns nil, nil if the image doesn't exist
func (s *DBImageStorage) Get(ctx context.Context, packID uint) (*Image, error) {
	query := `
		SELECT image_data, mime_type, file_size, width, height
		FROM pack_images
		WHERE pack_id = $1
	`

	var img Image
	err := database.DB().QueryRowContext(ctx, query, packID).Scan(
		&img.Data,
		&img.Metadata.MimeType,
		&img.Metadata.FileSize,
		&img.Metadata.Width,
		&img.Metadata.Height,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return &img, nil
}

// Delete removes an image for a pack from the database
// Returns nil if the image doesn't exist (idempotent)
func (s *DBImageStorage) Delete(ctx context.Context, packID uint) error {
	query := `DELETE FROM pack_images WHERE pack_id = $1`

	_, err := database.DB().ExecContext(ctx, query, packID)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	// Note: SQL DELETE is idempotent - deleting a non-existent row succeeds
	return nil
}

// Exists checks if an image exists for a pack
func (s *DBImageStorage) Exists(ctx context.Context, packID uint) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pack_images WHERE pack_id = $1)`

	var exists bool
	err := database.DB().QueryRowContext(ctx, query, packID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check image existence: %w", err)
	}

	return exists, nil
}

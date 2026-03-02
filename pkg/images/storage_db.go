package images

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
)

// dbImageStore is a shared base for image storage backed by PostgreSQL.
// It is parameterized by table name and ID column name, so that both
// pack_images and account_images can reuse the same CRUD logic.
type dbImageStore struct {
	table    string // e.g. "pack_images", "account_images"
	idColumn string // e.g. "pack_id", "account_id"
}

func (s *dbImageStore) save(ctx context.Context, ownerID uint, data []byte, metadata ImageMetadata) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (%s, image_data, mime_type, file_size, width, height, uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (%s)
		DO UPDATE SET
			image_data = EXCLUDED.image_data,
			mime_type = EXCLUDED.mime_type,
			file_size = EXCLUDED.file_size,
			width = EXCLUDED.width,
			height = EXCLUDED.height,
			updated_at = EXCLUDED.updated_at
	`, s.table, s.idColumn, s.idColumn)

	now := time.Now()
	_, err := database.DB().ExecContext(ctx, query,
		ownerID, data, metadata.MimeType, metadata.FileSize,
		metadata.Width, metadata.Height, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to save image in %s: %w", s.table, err)
	}
	return nil
}

func (s *dbImageStore) get(ctx context.Context, ownerID uint) (*Image, error) {
	query := fmt.Sprintf(`
		SELECT image_data, mime_type, file_size, width, height
		FROM %s WHERE %s = $1
	`, s.table, s.idColumn)

	var img Image
	err := database.DB().QueryRowContext(ctx, query, ownerID).Scan(
		&img.Data, &img.Metadata.MimeType, &img.Metadata.FileSize,
		&img.Metadata.Width, &img.Metadata.Height,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get image from %s: %w", s.table, err)
	}
	return &img, nil
}

func (s *dbImageStore) delete(ctx context.Context, ownerID uint) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1`, s.table, s.idColumn)
	_, err := database.DB().ExecContext(ctx, query, ownerID)
	if err != nil {
		return fmt.Errorf("failed to delete image from %s: %w", s.table, err)
	}
	return nil
}

func (s *dbImageStore) exists(ctx context.Context, ownerID uint) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1)`, s.table, s.idColumn)
	var exists bool
	err := database.DB().QueryRowContext(ctx, query, ownerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check image existence in %s: %w", s.table, err)
	}
	return exists, nil
}

// DBImageStorage implements ImageStorage using PostgreSQL database (pack_images table)
type DBImageStorage struct {
	store dbImageStore
}

// NewDBImageStorage creates a new database image storage instance for packs
func NewDBImageStorage() *DBImageStorage {
	return &DBImageStorage{store: dbImageStore{table: "pack_images", idColumn: "pack_id"}}
}

func (s *DBImageStorage) Save(ctx context.Context, packID uint, data []byte, metadata ImageMetadata) error {
	return s.store.save(ctx, packID, data, metadata)
}

func (s *DBImageStorage) Get(ctx context.Context, packID uint) (*Image, error) {
	return s.store.get(ctx, packID)
}

func (s *DBImageStorage) Delete(ctx context.Context, packID uint) error {
	return s.store.delete(ctx, packID)
}

func (s *DBImageStorage) Exists(ctx context.Context, packID uint) (bool, error) {
	return s.store.exists(ctx, packID)
}

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

// DBStorage wraps dbImageStore with exported methods, implementing
// ImageStorage, AccountImageStorage, and InventoryImageStorage interfaces.
// Use the specific constructor functions to create instances for each table.
type DBStorage struct {
	store dbImageStore
}

func (s *DBStorage) Save(ctx context.Context, id uint, data []byte, metadata ImageMetadata) error {
	return s.store.save(ctx, id, data, metadata)
}

func (s *DBStorage) Get(ctx context.Context, id uint) (*Image, error) {
	return s.store.get(ctx, id)
}

func (s *DBStorage) Delete(ctx context.Context, id uint) error {
	return s.store.delete(ctx, id)
}

func (s *DBStorage) Exists(ctx context.Context, id uint) (bool, error) {
	return s.store.exists(ctx, id)
}

// NewDBImageStorage creates a new database image storage instance for packs
func NewDBImageStorage() *DBStorage {
	return &DBStorage{store: dbImageStore{table: "pack_images", idColumn: "pack_id"}}
}

// NewDBAccountImageStorage creates a new database account image storage instance
func NewDBAccountImageStorage() *DBStorage {
	return &DBStorage{store: dbImageStore{table: "account_images", idColumn: "account_id"}}
}

// NewDBInventoryImageStorage creates a new database inventory image storage instance
func NewDBInventoryImageStorage() *DBStorage {
	return &DBStorage{store: dbImageStore{table: "inventory_images", idColumn: "item_id"}}
}

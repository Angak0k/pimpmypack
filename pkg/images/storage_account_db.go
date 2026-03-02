package images

import (
	"context"
)

// DBAccountImageStorage implements AccountImageStorage using PostgreSQL database (account_images table)
type DBAccountImageStorage struct {
	store dbImageStore
}

// NewDBAccountImageStorage creates a new database account image storage instance
func NewDBAccountImageStorage() *DBAccountImageStorage {
	return &DBAccountImageStorage{store: dbImageStore{table: "account_images", idColumn: "account_id"}}
}

func (s *DBAccountImageStorage) Save(ctx context.Context, accountID uint, data []byte, metadata ImageMetadata) error {
	return s.store.save(ctx, accountID, data, metadata)
}

func (s *DBAccountImageStorage) Get(ctx context.Context, accountID uint) (*Image, error) {
	return s.store.get(ctx, accountID)
}

func (s *DBAccountImageStorage) Delete(ctx context.Context, accountID uint) error {
	return s.store.delete(ctx, accountID)
}

func (s *DBAccountImageStorage) Exists(ctx context.Context, accountID uint) (bool, error) {
	return s.store.exists(ctx, accountID)
}

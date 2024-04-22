package migration

import (
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	// Import the PostgreSQL driver to register it with database/sql.
	// This allows us to use PostgreSQL with the migrate tool.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migration_scripts/*.sql
var fs embed.FS

func Migration(dbURL string) error {
	d, err := iofs.New(fs, "migration_scripts")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		return err
	}
	defer m.Close()

	// Apply all available migrations
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

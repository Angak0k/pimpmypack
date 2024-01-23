package migration

import (
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migration_scripts/*.sql
var fs embed.FS

func Migration(dbUrl string) error {

	d, err := iofs.New(fs, "migration_scripts")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, dbUrl)
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

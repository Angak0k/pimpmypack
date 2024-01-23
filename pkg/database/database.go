package database

import (
	"database/sql"
	"fmt"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database/migration"
	_ "github.com/lib/pq"
)

var db *sql.DB

func DbUrl() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", config.DbUser, config.DbPassword, config.DbHost, config.DbPort, config.DbName)
}

func DatabaseInit() error {
	var err error
	db, err = sql.Open("postgres", DbUrl())
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func DatabaseMigrate() error {
	err := migration.Migration(DbUrl())
	if err != nil {
		return err
	}
	return nil
}

// Getter for db var
func Db() *sql.DB {
	return db
}

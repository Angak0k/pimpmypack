package database

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database/migration"

	// Import the PostgreSQL driver to register it with database/sql.
	// This allows us to use PostgreSQL with the standard SQL package.
	_ "github.com/lib/pq"
)

var db *sql.DB

func DBUrl() string {
	hostPort := net.JoinHostPort(config.DBHost, strconv.Itoa(config.DBPort))
	return fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=disable",
		config.DBUser,
		config.DBPassword,
		hostPort,
		config.DBName,
	)
}

func Initialization() error {
	var err error
	db, err = sql.Open("postgres", DBUrl())
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func Migrate() error {
	err := migration.Migration(DBUrl())
	if err != nil {
		return err
	}
	return nil
}

// Getter for db var
func DB() *sql.DB {
	return db
}

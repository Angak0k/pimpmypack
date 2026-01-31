package main

import (
	"fmt"
	"log"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
)

func main() {
	// Init env
	err := config.EnvInit("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Init DB
	err = database.Initialization()
	if err != nil {
		log.Fatalf("Error connecting database: %v", err)
	}

	// Force set schema_migrations to clean state
	// Drop the dirty flag and set version to 10
	query := "UPDATE schema_migrations SET version = 10, dirty = false " +
		"WHERE version = 11 AND dirty = true;"
	_, err = database.DB().Exec(query)
	if err != nil {
		log.Fatalf("Error fixing migration: %v", err)
	}

	// Drop pack_images table if it exists (may be partially created)
	_, err = database.DB().Exec("DROP TABLE IF EXISTS pack_images CASCADE;")
	if err != nil {
		log.Fatalf("Error dropping pack_images: %v", err)
	}

	fmt.Println("Migration fixed successfully. Version set to 10, pack_images table dropped.")
	fmt.Println("Now run your tests again - they will re-apply migration 11.")
}

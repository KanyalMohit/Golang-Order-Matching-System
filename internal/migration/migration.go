package migration

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs all pending database migrations
func RunMigrations(db *sql.DB) error {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to get project root: %v", err)
	}

	migrationsPath := filepath.Join(projectRoot, "migrations")
	log.Printf("Looking for migrations in: %s", migrationsPath)

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %v", err)
	}

	// Check if we need to force a version
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("could not get migration version: %v", err)
	}

	if dirty {
		log.Printf("Found dirty database at version %d, forcing version", version)
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("could not force version: %v", err)
		}
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

// RollbackLastMigration rolls back the last applied migration
func RollbackLastMigration(db *sql.DB) error {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to get project root: %v", err)
	}

	migrationsPath := filepath.Join(projectRoot, "migrations")
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %v", err)
	}

	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("could not rollback migration: %v", err)
	}

	log.Println("Rollback completed successfully")
	return nil
}

// getProjectRoot returns the absolute path to the project root directory
func getProjectRoot() (string, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %v", err)
	}

	// If we're in cmd/server, go up one level
	if filepath.Base(wd) == "server" {
		wd = filepath.Dir(wd)
	}
	// If we're in cmd, go up one more level
	if filepath.Base(wd) == "cmd" {
		wd = filepath.Dir(wd)
	}

	return wd, nil
}

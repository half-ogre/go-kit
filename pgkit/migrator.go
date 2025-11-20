package pgkit

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/half-ogre/go-kit/kit"
)

// Migrator is an interface for running database migrations
type Migrator interface {
	RunMigrations(db DB, dirPath string) error
	RunMigrationsToVersion(db DB, dirPath string, toVersion int) error
}

// migrator implements Migrator
type migrator struct{}

// parseMigrationVersion extracts the version number from a migration filename
// Expected format: {number}_{description}.sql
// Returns the version number and an error if the format is invalid
func parseMigrationVersion(filename string) (int, error) {
	// Remove .sql extension
	nameWithoutExt := strings.TrimSuffix(filename, ".sql")
	if nameWithoutExt == filename {
		return 0, fmt.Errorf("migration file must have .sql extension")
	}

	// Split on underscore
	parts := strings.SplitN(nameWithoutExt, "_", 2)
	if len(parts) < 2 {
		return 0, fmt.Errorf("migration filename must be in format: {number}_{description}.sql")
	}

	// Parse the version number
	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("migration filename must start with a number: %w", err)
	}

	return version, nil
}

func (m *migrator) RunMigrations(db DB, dirPath string) error {
	return m.runMigrations(db, dirPath, 0)
}

func (m *migrator) RunMigrationsToVersion(db DB, dirPath string, toVersion int) error {
	if toVersion <= 0 {
		return fmt.Errorf("toVersion must be greater than 0")
	}
	return m.runMigrations(db, dirPath, toVersion)
}

func (m *migrator) runMigrations(db DB, dirPath string, toVersion int) error {
	if db == nil {
		return fmt.Errorf("database connection cannot be nil")
	}
	if dirPath == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	migrationsFS := os.DirFS(dirPath)

	// Create migrations tracking table
	_, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS pgkit_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return kit.WrapError(err, "failed to create pgkit_migrations table")
	}

	// Get all migration files
	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return kit.WrapError(err, "failed to read migration directory")
	}

	var filenames []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			// Validate migration filename format
			_, err := parseMigrationVersion(entry.Name())
			if err != nil {
				return kit.WrapError(err, "invalid migration filename: %s", entry.Name())
			}
			filenames = append(filenames, entry.Name())
		}
	}
	sort.Strings(filenames) // Ensure alphabetical order

	// Validate toVersion exists if specified
	if toVersion > 0 {
		found := false
		for _, filename := range filenames {
			version, _ := parseMigrationVersion(filename) // Already validated above
			if version == toVersion {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("migration with version %d not found", toVersion)
		}
	}

	// Run each migration if not already applied
	for _, filename := range filenames {
		version, _ := parseMigrationVersion(filename) // Already validated above

		var exists bool
		err := db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM pgkit_migrations WHERE filename = $1)", filename).Scan(&exists)
		if err != nil {
			return kit.WrapError(err, "failed to check migration %s", filename)
		}

		// Check if this is the target version
		isTargetVersion := toVersion > 0 && version == toVersion

		if exists {
			// If we've reached the target version and it's already applied, we're done
			if isTargetVersion {
				break
			}
			continue // Skip already applied migrations
		}

		// Read and execute migration
		content, err := fs.ReadFile(migrationsFS, filename)
		if err != nil {
			return kit.WrapError(err, "failed to read migration %s", filename)
		}

		_, err = db.Exec(context.Background(), string(content))
		if err != nil {
			return kit.WrapError(err, "failed to execute migration %s", filename)
		}

		// Record migration as applied
		_, err = db.Exec(context.Background(), "INSERT INTO pgkit_migrations (filename) VALUES ($1)", filename)
		if err != nil {
			return kit.WrapError(err, "failed to record migration %s", filename)
		}

		// Stop if we've reached the target version
		if isTargetVersion {
			break
		}
	}

	return nil
}

// NewMigrator creates a new Migrator
func NewMigrator() Migrator {
	return &migrator{}
}

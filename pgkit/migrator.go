package pgkit

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/half-ogre/go-kit/kit"
)

// Migrator is an interface for running database migrations
type Migrator interface {
	RunMigrations(db DB, dirPath string) error
}

// migrator implements Migrator
type migrator struct{}

func (m *migrator) RunMigrations(db DB, dirPath string) error {
	if db == nil {
		return fmt.Errorf("database connection cannot be nil")
	}
	if dirPath == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	migrationsFS := os.DirFS(dirPath)

	// Create migrations tracking table
	_, err := db.Exec(`
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
			filenames = append(filenames, entry.Name())
		}
	}
	sort.Strings(filenames) // Ensure alphabetical order

	// Run each migration if not already applied
	for _, filename := range filenames {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pgkit_migrations WHERE filename = $1)", filename).Scan(&exists)
		if err != nil {
			return kit.WrapError(err, "failed to check migration %s", filename)
		}

		if exists {
			continue // Skip already applied migrations
		}

		// Read and execute migration
		content, err := fs.ReadFile(migrationsFS, filename)
		if err != nil {
			return kit.WrapError(err, "failed to read migration %s", filename)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			return kit.WrapError(err, "failed to execute migration %s", filename)
		}

		// Record migration as applied
		_, err = db.Exec("INSERT INTO pgkit_migrations (filename) VALUES ($1)", filename)
		if err != nil {
			return kit.WrapError(err, "failed to record migration %s", filename)
		}
	}

	return nil
}

// NewMigrator creates a new Migrator
func NewMigrator() Migrator {
	return &migrator{}
}

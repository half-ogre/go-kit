//go:build acceptance

package pgkit_test

import (
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMigrations(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping acceptance test - DATABASE_URL not set")
	}

	dbURL := os.Getenv("DATABASE_URL")

	t.Run("creates_migrations_table_on_first_run", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()

		migrator := pgkit.NewMigrator()
		err := migrator.RunMigrations(db, "testdata")

		require.NoError(t, err)
		var tableExists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_name = 'pgkit_migrations'
			)
		`).Scan(&tableExists)
		require.NoError(t, err)
		assert.True(t, tableExists, "pgkit_migrations table should exist")
		var columnCount int
		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM information_schema.columns
			WHERE table_name = 'pgkit_migrations'
		`).Scan(&columnCount)
		require.NoError(t, err)
		assert.Equal(t, 3, columnCount, "pgkit_migrations table should have 3 columns")
	})

	t.Run("applies_all_migrations_in_order", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()

		migrator := pgkit.NewMigrator()
		err := migrator.RunMigrations(db, "testdata")

		require.NoError(t, err)
		var migrationCount int
		err = db.QueryRow("SELECT COUNT(*) FROM pgkit_migrations").Scan(&migrationCount)
		require.NoError(t, err)
		assert.Equal(t, 4, migrationCount, "should have 4 migrations applied")
		rows, err := db.Query("SELECT filename FROM pgkit_migrations ORDER BY id")
		require.NoError(t, err)
		defer rows.Close()
		expectedFiles := []string{
			"001_create_users.sql",
			"002_add_email_to_users.sql",
			"003_add_index_on_email.sql",
			"004_add_status_column.sql",
		}
		var actualFiles []string
		for rows.Next() {
			var filename string
			err := rows.Scan(&filename)
			require.NoError(t, err)
			actualFiles = append(actualFiles, filename)
		}
		assert.Equal(t, expectedFiles, actualFiles, "migrations should be applied in alphabetical order")
	})

	t.Run("creates_expected_schema", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()

		migrator := pgkit.NewMigrator()
		err := migrator.RunMigrations(db, "testdata")

		require.NoError(t, err)
		var tableExists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_name = 'test_users'
			)
		`).Scan(&tableExists)
		require.NoError(t, err)
		assert.True(t, tableExists, "test_users table should exist")
		var columnNames []string
		rows, err := db.Query(`
			SELECT column_name
			FROM information_schema.columns
			WHERE table_name = 'test_users'
			ORDER BY ordinal_position
		`)
		require.NoError(t, err)
		defer rows.Close()
		for rows.Next() {
			var colName string
			err := rows.Scan(&colName)
			require.NoError(t, err)
			columnNames = append(columnNames, colName)
		}
		expectedColumns := []string{"id", "name", "created_at", "email", "status"}
		assert.Equal(t, expectedColumns, columnNames, "test_users should have expected columns")
		var indexExists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM pg_indexes
				WHERE tablename = 'test_users' AND indexname = 'idx_test_users_email'
			)
		`).Scan(&indexExists)
		require.NoError(t, err)
		assert.True(t, indexExists, "email index should exist")
	})

	t.Run("skips_already_applied_migrations", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()
		migrator := pgkit.NewMigrator()
		err := migrator.RunMigrations(db, "testdata")
		require.NoError(t, err)
		var countBefore int
		err = db.QueryRow("SELECT COUNT(*) FROM pgkit_migrations").Scan(&countBefore)
		require.NoError(t, err)

		err = migrator.RunMigrations(db, "testdata")

		require.NoError(t, err)
		var countAfter int
		err = db.QueryRow("SELECT COUNT(*) FROM pgkit_migrations").Scan(&countAfter)
		require.NoError(t, err)
		assert.Equal(t, countBefore, countAfter, "migration count should not increase on second run")
	})

	t.Run("is_idempotent", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()

		migrator := pgkit.NewMigrator()
		for i := 0; i < 3; i++ {
			err := migrator.RunMigrations(db, "testdata")
			require.NoError(t, err, "run %d should succeed", i+1)
		}

		var migrationCount int
		err := db.QueryRow("SELECT COUNT(*) FROM pgkit_migrations").Scan(&migrationCount)
		require.NoError(t, err)
		assert.Equal(t, 4, migrationCount, "should have exactly 4 migrations after multiple runs")
	})

	t.Run("records_applied_timestamp", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()

		migrator := pgkit.NewMigrator()
		err := migrator.RunMigrations(db, "testdata")

		require.NoError(t, err)
		var nullTimestamps int
		err = db.QueryRow("SELECT COUNT(*) FROM pgkit_migrations WHERE applied_at IS NULL").Scan(&nullTimestamps)
		require.NoError(t, err)
		assert.Equal(t, 0, nullTimestamps, "all migrations should have applied_at timestamps")
	})

	t.Run("can_insert_data_after_migrations", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()
		migrator := pgkit.NewMigrator()
		err := migrator.RunMigrations(db, "testdata")
		require.NoError(t, err)

		_, err = db.Exec(`
			INSERT INTO test_users (name, email, status)
			VALUES ($1, $2, $3)
		`, "Test User", "test@example.com", "active")

		require.NoError(t, err)
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM test_users").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "should have 1 user after insert")
		var name, email, status string
		err = db.QueryRow("SELECT name, email, status FROM test_users LIMIT 1").Scan(&name, &email, &status)
		require.NoError(t, err)
		assert.Equal(t, "Test User", name)
		assert.Equal(t, "test@example.com", email)
		assert.Equal(t, "active", status)
	})

	t.Run("returns_error_when_directory_does_not_exist", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()

		migrator := pgkit.NewMigrator()
		err := migrator.RunMigrations(db, "nonexistent-directory")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read migration directory")
	})

	t.Run("returns_error_when_database_connection_is_nil", func(t *testing.T) {
		migrator := pgkit.NewMigrator()
		err := migrator.RunMigrations(nil, "testdata")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection cannot be nil")
	})

	t.Run("returns_error_when_directory_path_is_empty", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()

		migrator := pgkit.NewMigrator()
		err := migrator.RunMigrations(db, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "directory path cannot be empty")
	})
}

// setupTestDB creates a fresh test database for each test
func setupTestDB(t *testing.T, dbURL string) pgkit.DB {
	t.Helper()

	db, err := pgkit.NewDB(dbURL)
	require.NoError(t, err)

	// Clean up any existing test tables and migrations
	cleanupTestDB(t, db)

	return db
}

// cleanupTestDB drops all test tables and the pgkit_migrations table
func cleanupTestDB(t *testing.T, db pgkit.DB) {
	t.Helper()

	// Drop test_users table if it exists
	_, err := db.Exec("DROP TABLE IF EXISTS test_users CASCADE")
	require.NoError(t, err)

	// Drop pgkit_migrations table if it exists
	_, err = db.Exec("DROP TABLE IF EXISTS pgkit_migrations CASCADE")
	require.NoError(t, err)
}

//go:build acceptance

package pgkit_test

import (
	"os"
	"testing"
	"time"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListMigrationsFromDir(t *testing.T) {
	t.Run("lists_all_migrations_from_testdata_directory", func(t *testing.T) {
		migrations, err := pgkit.ListMigrationsFromDir("testdata")

		require.NoError(t, err)
		assert.Len(t, migrations, 4)
		assert.Equal(t, 1, migrations[0].Version)
		assert.Equal(t, "create_users", migrations[0].Description)
		assert.Equal(t, "001_create_users.sql", migrations[0].Filename)
		assert.False(t, migrations[0].Applied)
		assert.Nil(t, migrations[0].AppliedAt)
		assert.Equal(t, 2, migrations[1].Version)
		assert.Equal(t, "add_email_to_users", migrations[1].Description)
		assert.Equal(t, "002_add_email_to_users.sql", migrations[1].Filename)
		assert.Equal(t, 3, migrations[2].Version)
		assert.Equal(t, "add_index_on_email", migrations[2].Description)
		assert.Equal(t, "003_add_index_on_email.sql", migrations[2].Filename)
		assert.Equal(t, 4, migrations[3].Version)
		assert.Equal(t, "add_status_column", migrations[3].Description)
		assert.Equal(t, "004_add_status_column.sql", migrations[3].Filename)
	})

	t.Run("returns_migrations_sorted_by_version_number", func(t *testing.T) {
		migrations, err := pgkit.ListMigrationsFromDir("testdata")

		require.NoError(t, err)
		for i := 0; i < len(migrations)-1; i++ {
			assert.Less(t, migrations[i].Version, migrations[i+1].Version, "migrations should be sorted by version")
		}
	})

	t.Run("returns_error_when_directory_does_not_exist", func(t *testing.T) {
		migrations, err := pgkit.ListMigrationsFromDir("nonexistent")

		assert.Nil(t, migrations)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read migration directory")
	})

	t.Run("returns_error_when_directory_path_is_empty", func(t *testing.T) {
		migrations, err := pgkit.ListMigrationsFromDir("")

		assert.Nil(t, migrations)
		assert.EqualError(t, err, "directory path cannot be empty")
	})
}

func TestListMigrations(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")

	t.Run("lists_all_migrations_with_applied_status_and_timestamps", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		migrator := pgkit.NewMigrator()

		// Record time before running migrations
		beforeRun := time.Now()

		// First, run some migrations
		err := migrator.RunMigrationsToVersion(db, "testdata", 2)
		require.NoError(t, err)

		// Record time after running migrations
		afterRun := time.Now()

		// Now list all migrations
		migrations, err := migrator.ListMigrations(db, "testdata")

		require.NoError(t, err)
		assert.Len(t, migrations, 4)

		// First migration - applied
		assert.Equal(t, 1, migrations[0].Version)
		assert.Equal(t, "create_users", migrations[0].Description)
		assert.Equal(t, "001_create_users.sql", migrations[0].Filename)
		assert.True(t, migrations[0].Applied)
		assert.NotNil(t, migrations[0].AppliedAt)
		assert.True(t, migrations[0].AppliedAt.After(beforeRun) || migrations[0].AppliedAt.Equal(beforeRun),
			"applied timestamp should be after or equal to beforeRun")
		assert.True(t, migrations[0].AppliedAt.Before(afterRun) || migrations[0].AppliedAt.Equal(afterRun),
			"applied timestamp should be before or equal to afterRun")

		// Second migration - applied
		assert.Equal(t, 2, migrations[1].Version)
		assert.Equal(t, "add_email_to_users", migrations[1].Description)
		assert.Equal(t, "002_add_email_to_users.sql", migrations[1].Filename)
		assert.True(t, migrations[1].Applied)
		assert.NotNil(t, migrations[1].AppliedAt)
		assert.True(t, migrations[1].AppliedAt.After(beforeRun) || migrations[1].AppliedAt.Equal(beforeRun),
			"applied timestamp should be after or equal to beforeRun")
		assert.True(t, migrations[1].AppliedAt.Before(afterRun) || migrations[1].AppliedAt.Equal(afterRun),
			"applied timestamp should be before or equal to afterRun")

		// Third migration - not applied
		assert.Equal(t, 3, migrations[2].Version)
		assert.Equal(t, "add_index_on_email", migrations[2].Description)
		assert.Equal(t, "003_add_index_on_email.sql", migrations[2].Filename)
		assert.False(t, migrations[2].Applied)
		assert.Nil(t, migrations[2].AppliedAt)

		// Fourth migration - not applied
		assert.Equal(t, 4, migrations[3].Version)
		assert.Equal(t, "add_status_column", migrations[3].Description)
		assert.Equal(t, "004_add_status_column.sql", migrations[3].Filename)
		assert.False(t, migrations[3].Applied)
		assert.Nil(t, migrations[3].AppliedAt)
	})

	t.Run("returns_migrations_sorted_by_version_number", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		migrator := pgkit.NewMigrator()

		migrations, err := migrator.ListMigrations(db, "testdata")

		require.NoError(t, err)
		for i := 0; i < len(migrations)-1; i++ {
			assert.Less(t, migrations[i].Version, migrations[i+1].Version, "migrations should be sorted by version")
		}
	})

	t.Run("returns_all_migrations_as_not_applied_when_no_migrations_have_run", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		migrator := pgkit.NewMigrator()

		migrations, err := migrator.ListMigrations(db, "testdata")

		require.NoError(t, err)
		assert.Len(t, migrations, 4)
		for _, m := range migrations {
			assert.False(t, m.Applied, "migration %s should not be applied", m.Filename)
			assert.Nil(t, m.AppliedAt, "migration %s should not have applied timestamp", m.Filename)
		}
	})

	t.Run("returns_error_when_directory_does_not_exist", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		migrator := pgkit.NewMigrator()

		migrations, err := migrator.ListMigrations(db, "nonexistent")

		assert.Nil(t, migrations)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read migration directory")
	})

	t.Run("returns_error_when_directory_path_is_empty", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		migrator := pgkit.NewMigrator()

		migrations, err := migrator.ListMigrations(db, "")

		assert.Nil(t, migrations)
		assert.EqualError(t, err, "directory path cannot be empty")
	})

	t.Run("returns_error_when_database_connection_is_nil", func(t *testing.T) {
		migrator := pgkit.NewMigrator()

		migrations, err := migrator.ListMigrations(nil, "testdata")

		assert.Nil(t, migrations)
		assert.EqualError(t, err, "database connection cannot be nil")
	})
}

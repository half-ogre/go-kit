package pgkit

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestRunMigrations(t *testing.T) {
	t.Run("successfully_creates_migrations_table_and_runs_all_new_migrations", func(t *testing.T) {
		execCallCount := 0
		var execQueries []string
		queryRowCallCount := 0
		var queryRowQueries []string
		var queryRowArgs []string

		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				execCallCount++
				execQueries = append(execQueries, query)
				return nil, nil
			},
			QueryRowFake: func(ctx context.Context, query string, args ...any) Row {
				queryRowCallCount++
				queryRowQueries = append(queryRowQueries, query)
				if len(args) > 0 {
					queryRowArgs = append(queryRowArgs, args[0].(string))
				}
				return &FakeRow{
					ScanFake: func(dest ...any) error {
						// All migrations are new (don't exist yet)
						*dest[0].(*bool) = false
						return nil
					},
				}
			},
		}

		migrator := NewMigrator()
		err := migrator.RunMigrations(fakeDB, "testdata")

		assert.NoError(t, err)
		// Should call Exec: 1 for CREATE TABLE + 2 migrations + 2 INSERTs = 5 times
		assert.Equal(t, 5, execCallCount)
		// Should call QueryRow: 2 times (once per migration file)
		assert.Equal(t, 2, queryRowCallCount)
		// Verify CREATE TABLE was called first
		assert.Contains(t, execQueries[0], "CREATE TABLE IF NOT EXISTS pgkit_migrations")
		// Verify migrations checked in alphabetical order
		assert.Equal(t, []string{"001_initial.sql", "002_add_email.sql"}, queryRowArgs)
		// Verify migrations executed
		assert.Contains(t, execQueries[1], "CREATE TABLE users")
		assert.Contains(t, execQueries[2], "INSERT INTO pgkit_migrations")
		assert.Contains(t, execQueries[3], "ALTER TABLE users ADD COLUMN email")
		assert.Contains(t, execQueries[4], "INSERT INTO pgkit_migrations")
	})

	t.Run("skips_migrations_that_have_already_been_applied", func(t *testing.T) {
		execCallCount := 0
		queryRowCallCount := 0

		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				execCallCount++
				return nil, nil
			},
			QueryRowFake: func(ctx context.Context, query string, args ...any) Row {
				queryRowCallCount++
				filename := args[0].(string)
				return &FakeRow{
					ScanFake: func(dest ...any) error {
						// First migration already exists, second is new
						if filename == "001_initial.sql" {
							*dest[0].(*bool) = true
						} else {
							*dest[0].(*bool) = false
						}
						return nil
					},
				}
			},
		}

		migrator := NewMigrator()
		err := migrator.RunMigrations(fakeDB, "testdata")

		assert.NoError(t, err)
		// Should call Exec: 1 for CREATE TABLE + 1 migration + 1 INSERT = 3 times
		assert.Equal(t, 3, execCallCount)
		// Should call QueryRow: 2 times (check both migrations)
		assert.Equal(t, 2, queryRowCallCount)
	})

	t.Run("returns_error_when_creating_migrations_table_fails", func(t *testing.T) {
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, assert.AnError
			},
		}

		migrator := NewMigrator()
		err := migrator.RunMigrations(fakeDB, "testdata")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create pgkit_migrations table")
	})

	t.Run("returns_error_when_checking_migration_existence_fails", func(t *testing.T) {
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			QueryRowFake: func(ctx context.Context, query string, args ...any) Row {
				return &FakeRow{
					ScanFake: func(dest ...any) error {
						return assert.AnError
					},
				}
			},
		}

		migrator := NewMigrator()
		err := migrator.RunMigrations(fakeDB, "testdata")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check migration")
	})

	t.Run("returns_error_when_executing_migration_fails", func(t *testing.T) {
		execCallCount := 0
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				execCallCount++
				// CREATE TABLE succeeds, but first migration execution fails
				if execCallCount > 1 {
					return nil, assert.AnError
				}
				return nil, nil
			},
			QueryRowFake: func(ctx context.Context, query string, args ...any) Row {
				return &FakeRow{
					ScanFake: func(dest ...any) error {
						*dest[0].(*bool) = false
						return nil
					},
				}
			},
		}

		migrator := NewMigrator()
		err := migrator.RunMigrations(fakeDB, "testdata")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute migration")
	})

	t.Run("returns_error_when_recording_migration_fails", func(t *testing.T) {
		execCallCount := 0
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				execCallCount++
				// CREATE TABLE and migration execution succeed, but INSERT fails
				if execCallCount == 3 {
					return nil, assert.AnError
				}
				return nil, nil
			},
			QueryRowFake: func(ctx context.Context, query string, args ...any) Row {
				return &FakeRow{
					ScanFake: func(dest ...any) error {
						*dest[0].(*bool) = false
						return nil
					},
				}
			},
		}

		migrator := NewMigrator()
		err := migrator.RunMigrations(fakeDB, "testdata")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to record migration")
	})

	t.Run("returns_an_error_when_database_connection_is_nil", func(t *testing.T) {
		migrator := NewMigrator()
		err := migrator.RunMigrations(nil, "testdata")

		assert.EqualError(t, err, "database connection cannot be nil")
	})

	t.Run("returns_an_error_when_directory_path_is_empty", func(t *testing.T) {
		fakeDB := &FakeDB{}

		migrator := NewMigrator()
		err := migrator.RunMigrations(fakeDB, "")

		assert.EqualError(t, err, "directory path cannot be empty")
	})
}

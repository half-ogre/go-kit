package pgkit

import (
	"context"
	"database/sql"
	"testing"
	"time"

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

func TestRunMigrationsToVersion(t *testing.T) {
	t.Run("runs_migrations_up_to_specified_version", func(t *testing.T) {
		execCallCount := 0
		var execQueries []string
		queryRowCallCount := 0
		var queryRowArgs []string

		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				execCallCount++
				execQueries = append(execQueries, query)
				return nil, nil
			},
			QueryRowFake: func(ctx context.Context, query string, args ...any) Row {
				queryRowCallCount++
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
		err := migrator.RunMigrationsToVersion(fakeDB, "testdata", 1)

		assert.NoError(t, err)
		// Should call Exec: 1 for CREATE TABLE + 1 migration + 1 INSERT = 3 times
		assert.Equal(t, 3, execCallCount)
		// Should call QueryRow: 1 time (only first migration)
		assert.Equal(t, 1, queryRowCallCount)
		// Verify only first migration was checked
		assert.Equal(t, []string{"001_initial.sql"}, queryRowArgs)
		// Verify CREATE TABLE was called first
		assert.Contains(t, execQueries[0], "CREATE TABLE IF NOT EXISTS pgkit_migrations")
		// Verify only first migration executed
		assert.Contains(t, execQueries[1], "CREATE TABLE users")
		assert.Contains(t, execQueries[2], "INSERT INTO pgkit_migrations")
	})

	t.Run("stops_at_target_version_when_already_applied", func(t *testing.T) {
		execCallCount := 0
		queryRowCallCount := 0

		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				execCallCount++
				return nil, nil
			},
			QueryRowFake: func(ctx context.Context, query string, args ...any) Row {
				queryRowCallCount++
				return &FakeRow{
					ScanFake: func(dest ...any) error {
						// First migration already exists
						*dest[0].(*bool) = true
						return nil
					},
				}
			},
		}

		migrator := NewMigrator()
		err := migrator.RunMigrationsToVersion(fakeDB, "testdata", 1)

		assert.NoError(t, err)
		// Should call Exec: 1 for CREATE TABLE only
		assert.Equal(t, 1, execCallCount)
		// Should call QueryRow: 1 time (check first migration)
		assert.Equal(t, 1, queryRowCallCount)
	})

	t.Run("returns_error_when_version_not_found", func(t *testing.T) {
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
		}

		migrator := NewMigrator()
		err := migrator.RunMigrationsToVersion(fakeDB, "testdata", 999)

		assert.EqualError(t, err, "migration with version 999 not found")
	})

	t.Run("returns_error_when_toVersion_is_zero_or_negative", func(t *testing.T) {
		fakeDB := &FakeDB{}

		migrator := NewMigrator()
		err := migrator.RunMigrationsToVersion(fakeDB, "testdata", 0)

		assert.EqualError(t, err, "toVersion must be greater than 0")

		err = migrator.RunMigrationsToVersion(fakeDB, "testdata", -1)

		assert.EqualError(t, err, "toVersion must be greater than 0")
	})

	t.Run("applies_multiple_migrations_up_to_target", func(t *testing.T) {
		execCallCount := 0
		var execQueries []string
		queryRowCallCount := 0

		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				execCallCount++
				execQueries = append(execQueries, query)
				return nil, nil
			},
			QueryRowFake: func(ctx context.Context, query string, args ...any) Row {
				queryRowCallCount++
				return &FakeRow{
					ScanFake: func(dest ...any) error {
						// All migrations are new
						*dest[0].(*bool) = false
						return nil
					},
				}
			},
		}

		migrator := NewMigrator()
		err := migrator.RunMigrationsToVersion(fakeDB, "testdata", 2)

		assert.NoError(t, err)
		// Should call Exec: 1 for CREATE TABLE + 2 migrations + 2 INSERTs = 5 times
		assert.Equal(t, 5, execCallCount)
		// Should call QueryRow: 2 times (both migrations)
		assert.Equal(t, 2, queryRowCallCount)
		// Verify both migrations executed
		assert.Contains(t, execQueries[1], "CREATE TABLE users")
		assert.Contains(t, execQueries[3], "ALTER TABLE users ADD COLUMN email")
	})
}

func TestParseMigrationVersion(t *testing.T) {
	t.Run("parses_valid_migration_filename", func(t *testing.T) {
		version, err := parseMigrationVersion("001_initial.sql")

		assert.NoError(t, err)
		assert.Equal(t, 1, version)
	})

	t.Run("parses_multi_digit_version", func(t *testing.T) {
		version, err := parseMigrationVersion("123_my_migration.sql")

		assert.NoError(t, err)
		assert.Equal(t, 123, version)
	})

	t.Run("returns_error_when_missing_sql_extension", func(t *testing.T) {
		version, err := parseMigrationVersion("001_initial.txt")

		assert.Equal(t, 0, version)
		assert.EqualError(t, err, "migration file must have .sql extension")
	})

	t.Run("returns_error_when_missing_underscore", func(t *testing.T) {
		version, err := parseMigrationVersion("001initial.sql")

		assert.Equal(t, 0, version)
		assert.EqualError(t, err, "migration filename must be in format: {number}_{description}.sql")
	})

	t.Run("returns_error_when_version_is_not_a_number", func(t *testing.T) {
		version, err := parseMigrationVersion("abc_initial.sql")

		assert.Equal(t, 0, version)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "migration filename must start with a number")
	})

	t.Run("handles_description_with_underscores", func(t *testing.T) {
		version, err := parseMigrationVersion("42_add_user_email_column.sql")

		assert.NoError(t, err)
		assert.Equal(t, 42, version)
	})
}

func TestListMigrationsFromDir(t *testing.T) {
	t.Run("returns_all_migrations_from_directory", func(t *testing.T) {
		migrations, err := ListMigrationsFromDir("testdata")

		assert.NoError(t, err)
		assert.Len(t, migrations, 2)
		assert.Equal(t, 1, migrations[0].Version)
		assert.Equal(t, "initial", migrations[0].Description)
		assert.Equal(t, "001_initial.sql", migrations[0].Filename)
		assert.False(t, migrations[0].Applied)
		assert.Nil(t, migrations[0].AppliedAt)
		assert.Equal(t, 2, migrations[1].Version)
		assert.Equal(t, "add_email", migrations[1].Description)
		assert.Equal(t, "002_add_email.sql", migrations[1].Filename)
		assert.False(t, migrations[1].Applied)
		assert.Nil(t, migrations[1].AppliedAt)
	})

	t.Run("returns_migrations_sorted_by_version", func(t *testing.T) {
		migrations, err := ListMigrationsFromDir("testdata")

		assert.NoError(t, err)
		assert.Len(t, migrations, 2)
		for i := 0; i < len(migrations)-1; i++ {
			assert.Less(t, migrations[i].Version, migrations[i+1].Version)
		}
	})

	t.Run("returns_error_when_directory_path_is_empty", func(t *testing.T) {
		migrations, err := ListMigrationsFromDir("")

		assert.Nil(t, migrations)
		assert.EqualError(t, err, "directory path cannot be empty")
	})

	t.Run("returns_error_when_directory_does_not_exist", func(t *testing.T) {
		migrations, err := ListMigrationsFromDir("nonexistent")

		assert.Nil(t, migrations)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read migration directory")
	})

	t.Run("returns_empty_list_when_directory_has_no_sql_files", func(t *testing.T) {
		migrations, err := ListMigrationsFromDir(".")

		assert.NoError(t, err)
		assert.Empty(t, migrations)
	})
}

func TestListMigrations(t *testing.T) {
	t.Run("returns_all_migrations_from_directory_with_applied_status_and_timestamps", func(t *testing.T) {
		appliedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		nextCallCount := 0
		fakeRows := &FakeRows{
			NextFake: func() bool {
				nextCallCount++
				return nextCallCount <= 1
			},
			ScanFake: func(dest ...any) error {
				*dest[0].(*string) = "001_initial.sql"
				*dest[1].(*time.Time) = appliedTime
				return nil
			},
			CloseFake: func() error { return nil },
			ErrFake:   func() error { return nil },
		}
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			QueryFake: func(ctx context.Context, query string, args ...any) (Rows, error) {
				return fakeRows, nil
			},
		}

		migrator := NewMigrator()
		migrations, err := migrator.ListMigrations(fakeDB, "testdata")

		assert.NoError(t, err)
		assert.Len(t, migrations, 2)
		assert.Equal(t, 1, migrations[0].Version)
		assert.Equal(t, "initial", migrations[0].Description)
		assert.Equal(t, "001_initial.sql", migrations[0].Filename)
		assert.True(t, migrations[0].Applied)
		assert.NotNil(t, migrations[0].AppliedAt)
		assert.Equal(t, appliedTime, *migrations[0].AppliedAt)
		assert.Equal(t, 2, migrations[1].Version)
		assert.Equal(t, "add_email", migrations[1].Description)
		assert.Equal(t, "002_add_email.sql", migrations[1].Filename)
		assert.False(t, migrations[1].Applied)
		assert.Nil(t, migrations[1].AppliedAt)
	})

	t.Run("returns_migrations_sorted_by_version", func(t *testing.T) {
		fakeRows := &FakeRows{
			NextFake:  func() bool { return false },
			CloseFake: func() error { return nil },
			ErrFake:   func() error { return nil },
		}
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			QueryFake: func(ctx context.Context, query string, args ...any) (Rows, error) {
				return fakeRows, nil
			},
		}

		migrator := NewMigrator()
		migrations, err := migrator.ListMigrations(fakeDB, "testdata")

		assert.NoError(t, err)
		assert.Len(t, migrations, 2)
		for i := 0; i < len(migrations)-1; i++ {
			assert.Less(t, migrations[i].Version, migrations[i+1].Version)
		}
	})

	t.Run("returns_error_when_database_connection_is_nil", func(t *testing.T) {
		migrator := NewMigrator()
		migrations, err := migrator.ListMigrations(nil, "testdata")

		assert.Nil(t, migrations)
		assert.EqualError(t, err, "database connection cannot be nil")
	})

	t.Run("returns_error_when_directory_path_is_empty", func(t *testing.T) {
		fakeDB := &FakeDB{}

		migrator := NewMigrator()
		migrations, err := migrator.ListMigrations(fakeDB, "")

		assert.Nil(t, migrations)
		assert.EqualError(t, err, "directory path cannot be empty")
	})

	t.Run("returns_error_when_creating_migrations_table_fails", func(t *testing.T) {
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, assert.AnError
			},
		}

		migrator := NewMigrator()
		migrations, err := migrator.ListMigrations(fakeDB, "testdata")

		assert.Nil(t, migrations)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create pgkit_migrations table")
	})

	t.Run("returns_error_when_directory_does_not_exist", func(t *testing.T) {
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
		}

		migrator := NewMigrator()
		migrations, err := migrator.ListMigrations(fakeDB, "nonexistent")

		assert.Nil(t, migrations)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read migration directory")
	})

	t.Run("returns_error_when_querying_applied_migrations_fails", func(t *testing.T) {
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			QueryFake: func(ctx context.Context, query string, args ...any) (Rows, error) {
				return nil, assert.AnError
			},
		}

		migrator := NewMigrator()
		migrations, err := migrator.ListMigrations(fakeDB, "testdata")

		assert.Nil(t, migrations)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query applied migrations")
	})

	t.Run("returns_error_when_scanning_migration_rows_fails", func(t *testing.T) {
		fakeRows := &FakeRows{
			NextFake:  func() bool { return true },
			ScanFake:  func(dest ...any) error { return assert.AnError },
			CloseFake: func() error { return nil },
		}
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			QueryFake: func(ctx context.Context, query string, args ...any) (Rows, error) {
				return fakeRows, nil
			},
		}

		migrator := NewMigrator()
		migrations, err := migrator.ListMigrations(fakeDB, "testdata")

		assert.Nil(t, migrations)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan migration row")
	})

	t.Run("returns_error_when_rows_err_returns_error", func(t *testing.T) {
		fakeRows := &FakeRows{
			NextFake:  func() bool { return false },
			CloseFake: func() error { return nil },
			ErrFake:   func() error { return assert.AnError },
		}
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			QueryFake: func(ctx context.Context, query string, args ...any) (Rows, error) {
				return fakeRows, nil
			},
		}

		migrator := NewMigrator()
		migrations, err := migrator.ListMigrations(fakeDB, "testdata")

		assert.Nil(t, migrations)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error iterating migration rows")
	})

	t.Run("returns_empty_list_when_directory_has_no_sql_files", func(t *testing.T) {
		fakeRows := &FakeRows{
			NextFake:  func() bool { return false },
			CloseFake: func() error { return nil },
			ErrFake:   func() error { return nil },
		}
		fakeDB := &FakeDB{
			ExecFake: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, nil
			},
			QueryFake: func(ctx context.Context, query string, args ...any) (Rows, error) {
				return fakeRows, nil
			},
		}

		migrator := NewMigrator()
		migrations, err := migrator.ListMigrations(fakeDB, ".")

		assert.NoError(t, err)
		assert.Empty(t, migrations)
	})
}

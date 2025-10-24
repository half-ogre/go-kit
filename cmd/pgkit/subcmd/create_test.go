package subcmd

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/stretchr/testify/assert"
)

func TestRunCreate(t *testing.T) {
	t.Run("checks_that_database_does_not_exist_and_executes_query_to_create_the_database", func(t *testing.T) {
		actualQueryRowQuery := ""
		actualDBName := ""
		actualExecQuery := ""
		fakeDB := &pgkit.FakeDB{
			QueryRowFake: func(query string, args ...any) pgkit.Row {
				actualQueryRowQuery = query
				if len(args) > 0 {
					actualDBName = args[0].(string)
				}
				return &pgkit.FakeRow{
					ScanFake: func(dest ...any) error {
						if boolPtr, ok := dest[0].(*bool); ok {
							*boolPtr = false
						}
						return nil
					},
				}
			},
			ExecFake: func(query string, args ...any) (sql.Result, error) {
				actualExecQuery = query
				return nil, nil
			},
		}

		err := runCreate(fakeDB, "theDatabase")

		assert.NoError(t, err)
		assert.Equal(t, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", actualQueryRowQuery)
		assert.Equal(t, "theDatabase", actualDBName)
		assert.Equal(t, `CREATE DATABASE "theDatabase"`, actualExecQuery)
	})

	t.Run("returns_error_when_database_already_exists", func(t *testing.T) {
		fakeDB := &pgkit.FakeDB{
			QueryRowFake: func(query string, args ...any) pgkit.Row {
				return &pgkit.FakeRow{
					ScanFake: func(dest ...any) error {
						if boolPtr, ok := dest[0].(*bool); ok {
							*boolPtr = true
						}
						return nil
					},
				}
			},
		}

		err := runCreate(fakeDB, "existing")

		assert.EqualError(t, err, "database 'existing' already exists")
	})

	t.Run("returns_error_when_existence_check_fails", func(t *testing.T) {
		fakeDB := &pgkit.FakeDB{
			QueryRowFake: func(query string, args ...any) pgkit.Row {
				return &pgkit.FakeRow{
					ScanFake: func(dest ...any) error {
						return errors.New("the query error")
					},
				}
			},
		}

		err := runCreate(fakeDB, "aDatabase")

		assert.EqualError(t, err, "failed to check if database exists: the query error")
	})

	t.Run("returns_error_when_create_execution_fails", func(t *testing.T) {
		fakeDB := &pgkit.FakeDB{
			QueryRowFake: func(query string, args ...any) pgkit.Row {
				return &pgkit.FakeRow{
					ScanFake: func(dest ...any) error {
						if boolPtr, ok := dest[0].(*bool); ok {
							*boolPtr = false
						}
						return nil
					},
				}
			},
			ExecFake: func(query string, args ...any) (sql.Result, error) {
				return nil, errors.New("the exec error")
			},
		}

		err := runCreate(fakeDB, "aDatabase")

		assert.EqualError(t, err, "failed to create database: the exec error")
	})

	t.Run("quotes_database_name_with_special_characters", func(t *testing.T) {
		actualExecQuery := ""
		fakeDB := &pgkit.FakeDB{
			QueryRowFake: func(query string, args ...any) pgkit.Row {
				return &pgkit.FakeRow{
					ScanFake: func(dest ...any) error {
						if boolPtr, ok := dest[0].(*bool); ok {
							*boolPtr = false
						}
						return nil
					},
				}
			},
			ExecFake: func(query string, args ...any) (sql.Result, error) {
				actualExecQuery = query
				return nil, nil
			},
		}

		err := runCreate(fakeDB, `my"db`)

		assert.NoError(t, err)
		assert.Equal(t, `CREATE DATABASE "my""db"`, actualExecQuery)
	})
}

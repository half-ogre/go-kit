package subcmd

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/stretchr/testify/assert"
)

func TestRunDrop(t *testing.T) {
	t.Run("checks_that_database_exists_and_executes_query_to_drop_the_database", func(t *testing.T) {
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
							*boolPtr = true
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

		err := runDrop(fakeDB, "theDatabase", true)

		assert.NoError(t, err)
		assert.Equal(t, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", actualQueryRowQuery)
		assert.Equal(t, "theDatabase", actualDBName)
		assert.Equal(t, `DROP DATABASE "theDatabase"`, actualExecQuery)
	})

	t.Run("returns_error_when_database_does_not_exist", func(t *testing.T) {
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
		}

		err := runDrop(fakeDB, "nonexistent", true)

		assert.EqualError(t, err, "database 'nonexistent' does not exist")
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

		err := runDrop(fakeDB, "aDatabase", true)

		assert.EqualError(t, err, "failed to check if database exists: the query error")
	})

	t.Run("returns_error_when_drop_execution_fails", func(t *testing.T) {
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
			ExecFake: func(query string, args ...any) (sql.Result, error) {
				return nil, errors.New("the exec error")
			},
		}

		err := runDrop(fakeDB, "aDatabase", true)

		assert.EqualError(t, err, "failed to drop database: the exec error")
	})

	t.Run("quotes_database_name_with_special_characters", func(t *testing.T) {
		actualExecQuery := ""
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
			ExecFake: func(query string, args ...any) (sql.Result, error) {
				actualExecQuery = query
				return nil, nil
			},
		}

		err := runDrop(fakeDB, `my"db`, true)

		assert.NoError(t, err)
		assert.Equal(t, `DROP DATABASE "my""db"`, actualExecQuery)
	})
}

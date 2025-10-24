package subcmd

import (
	"errors"
	"testing"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/stretchr/testify/assert"
)

func TestRunStatus(t *testing.T) {
	t.Run("successfully_displays_all_migrations", func(t *testing.T) {
		actualQuery := ""
		nextCallCount := 0
		scanCallCount := 0
		var actualFilenames []string
		var actualAppliedAts []string
		fakeRows := &pgkit.FakeRows{
			NextFake: func() bool {
				nextCallCount++
				return nextCallCount <= 2
			},
			ScanFake: func(dest ...any) error {
				scanCallCount++
				if nextCallCount == 1 {
					*dest[0].(*string) = "001_theMigration.sql"
					*dest[1].(*string) = "2025-01-01 12:00:00"
				} else if nextCallCount == 2 {
					*dest[0].(*string) = "002_anotherMigration.sql"
					*dest[1].(*string) = "2025-01-02 13:30:00"
				}
				actualFilenames = append(actualFilenames, *dest[0].(*string))
				actualAppliedAts = append(actualAppliedAts, *dest[1].(*string))
				return nil
			},
			CloseFake: func() error { return nil },
			ErrFake:   func() error { return nil },
		}
		fakeDB := &pgkit.FakeDB{
			QueryFake: func(query string, args ...any) (pgkit.Rows, error) {
				actualQuery = query
				return fakeRows, nil
			},
		}

		err := runStatus(fakeDB)

		assert.NoError(t, err)
		assert.Equal(t, "SELECT filename, applied_at FROM pgkit_migrations ORDER BY applied_at", actualQuery)
		assert.Equal(t, 2, scanCallCount)
		assert.Equal(t, []string{"001_theMigration.sql", "002_anotherMigration.sql"}, actualFilenames)
		assert.Equal(t, []string{"2025-01-01 12:00:00", "2025-01-02 13:30:00"}, actualAppliedAts)
	})

	t.Run("displays_no_migrations_message_when_no_rows", func(t *testing.T) {
		fakeRows := &pgkit.FakeRows{
			NextFake:  func() bool { return false },
			CloseFake: func() error { return nil },
			ErrFake:   func() error { return nil },
		}
		fakeDB := &pgkit.FakeDB{
			QueryFake: func(query string, args ...any) (pgkit.Rows, error) {
				return fakeRows, nil
			},
		}

		err := runStatus(fakeDB)

		assert.NoError(t, err)
	})

	t.Run("returns_error_when_query_fails", func(t *testing.T) {
		fakeDB := &pgkit.FakeDB{
			QueryFake: func(query string, args ...any) (pgkit.Rows, error) {
				return nil, errors.New("the query error")
			},
		}

		err := runStatus(fakeDB)

		assert.EqualError(t, err, "failed to query pgkit_migrations: the query error")
	})

	t.Run("returns_error_when_scan_fails", func(t *testing.T) {
		fakeRows := &pgkit.FakeRows{
			NextFake:  func() bool { return true },
			ScanFake:  func(dest ...any) error { return errors.New("the scan error") },
			CloseFake: func() error { return nil },
		}
		fakeDB := &pgkit.FakeDB{
			QueryFake: func(query string, args ...any) (pgkit.Rows, error) {
				return fakeRows, nil
			},
		}

		err := runStatus(fakeDB)

		assert.EqualError(t, err, "failed to scan row: the scan error")
	})

	t.Run("returns_error_when_rows_err_returns_error", func(t *testing.T) {
		fakeRows := &pgkit.FakeRows{
			NextFake:  func() bool { return false },
			CloseFake: func() error { return nil },
			ErrFake:   func() error { return errors.New("the rows error") },
		}
		fakeDB := &pgkit.FakeDB{
			QueryFake: func(query string, args ...any) (pgkit.Rows, error) {
				return fakeRows, nil
			},
		}

		err := runStatus(fakeDB)

		assert.EqualError(t, err, "the rows error")
	})
}

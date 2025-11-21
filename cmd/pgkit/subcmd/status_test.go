package subcmd

import (
	"errors"
	"testing"
	"time"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/stretchr/testify/assert"
)

func TestRunStatus(t *testing.T) {
	t.Run("successfully_displays_migrations_with_applied_status", func(t *testing.T) {
		fakeDB := &pgkit.FakeDB{}
		time1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		fakeMigrator := &pgkit.FakeMigrator{
			ListMigrationsFake: func(db pgkit.DB, dirPath string) ([]pgkit.Migration, error) {
				return []pgkit.Migration{
					{Version: 1, Description: "initial", Filename: "001_initial.sql", Applied: true, AppliedAt: &time1},
					{Version: 2, Description: "add_email", Filename: "002_add_email.sql", Applied: false, AppliedAt: nil},
				}, nil
			},
		}

		err := runStatus(fakeDB, "aMigrationsDir", fakeMigrator)

		assert.NoError(t, err)
	})

	t.Run("succeeds_when_no_migrations_found", func(t *testing.T) {
		fakeDB := &pgkit.FakeDB{}
		fakeMigrator := &pgkit.FakeMigrator{
			ListMigrationsFake: func(db pgkit.DB, dirPath string) ([]pgkit.Migration, error) {
				return []pgkit.Migration{}, nil
			},
		}

		err := runStatus(fakeDB, "aMigrationsDir", fakeMigrator)

		assert.NoError(t, err)
	})

	t.Run("returns_error_when_list_migrations_fails", func(t *testing.T) {
		fakeDB := &pgkit.FakeDB{}
		fakeMigrator := &pgkit.FakeMigrator{
			ListMigrationsFake: func(db pgkit.DB, dirPath string) ([]pgkit.Migration, error) {
				return nil, errors.New("the list error")
			},
		}

		err := runStatus(fakeDB, "aMigrationsDir", fakeMigrator)

		assert.EqualError(t, err, "failed to list migrations: the list error")
	})
}

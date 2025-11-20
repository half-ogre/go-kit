package subcmd

import (
	"errors"
	"testing"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/stretchr/testify/assert"
)

func TestRunMigrate(t *testing.T) {
	t.Run("successfully_runs_migrations_from_directory", func(t *testing.T) {
		fakeDB := &pgkit.FakeDB{}
		actualDir := ""
		fakeMigrator := &pgkit.FakeMigrator{
			RunMigrationsFake: func(db pgkit.DB, dir string) error {
				actualDir = dir
				return nil
			},
		}

		err := runMigrate(fakeDB, "theMigrationsDir", 0, fakeMigrator)

		assert.NoError(t, err)
		assert.Equal(t, "theMigrationsDir", actualDir)
	})

	t.Run("returns_error_when_migrator_returns_error", func(t *testing.T) {
		fakeDB := &pgkit.FakeDB{}
		fakeMigrator := &pgkit.FakeMigrator{
			RunMigrationsFake: func(db pgkit.DB, dir string) error {
				return errors.New("the migration error")
			},
		}

		err := runMigrate(fakeDB, "aMigrationsDir", 0, fakeMigrator)

		assert.EqualError(t, err, "migration failed: the migration error")
	})

	t.Run("successfully_runs_migrations_to_version", func(t *testing.T) {
		fakeDB := &pgkit.FakeDB{}
		actualDir := ""
		actualVersion := 0
		fakeMigrator := &pgkit.FakeMigrator{
			RunMigrationsToVersionFake: func(db pgkit.DB, dir string, toVersion int) error {
				actualDir = dir
				actualVersion = toVersion
				return nil
			},
		}

		err := runMigrate(fakeDB, "theMigrationsDir", 2, fakeMigrator)

		assert.NoError(t, err)
		assert.Equal(t, "theMigrationsDir", actualDir)
		assert.Equal(t, 2, actualVersion)
	})

	t.Run("returns_error_when_migrator_to_version_returns_error", func(t *testing.T) {
		fakeDB := &pgkit.FakeDB{}
		fakeMigrator := &pgkit.FakeMigrator{
			RunMigrationsToVersionFake: func(db pgkit.DB, dir string, toVersion int) error {
				return errors.New("the migration error")
			},
		}

		err := runMigrate(fakeDB, "aMigrationsDir", 2, fakeMigrator)

		assert.EqualError(t, err, "migration failed: the migration error")
	})
}

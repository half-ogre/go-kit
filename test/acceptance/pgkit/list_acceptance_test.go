//go:build acceptance

package pgkit_test

import (
	"testing"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListMigrations(t *testing.T) {
	t.Run("lists_all_migrations_from_testdata_directory", func(t *testing.T) {
		migrations, err := pgkit.ListMigrations("testdata")

		require.NoError(t, err)
		assert.Len(t, migrations, 4)
		assert.Equal(t, 1, migrations[0].Version)
		assert.Equal(t, "create_users", migrations[0].Description)
		assert.Equal(t, "001_create_users.sql", migrations[0].Filename)
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
		migrations, err := pgkit.ListMigrations("testdata")

		require.NoError(t, err)
		for i := 0; i < len(migrations)-1; i++ {
			assert.Less(t, migrations[i].Version, migrations[i+1].Version, "migrations should be sorted by version")
		}
	})

	t.Run("returns_error_when_directory_does_not_exist", func(t *testing.T) {
		migrations, err := pgkit.ListMigrations("nonexistent")

		assert.Nil(t, migrations)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read migration directory")
	})

	t.Run("returns_error_when_directory_path_is_empty", func(t *testing.T) {
		migrations, err := pgkit.ListMigrations("")

		assert.Nil(t, migrations)
		assert.EqualError(t, err, "directory path cannot be empty")
	})
}

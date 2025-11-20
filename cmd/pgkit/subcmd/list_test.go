package subcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunList(t *testing.T) {
	t.Run("successfully_lists_migrations_from_directory", func(t *testing.T) {
		err := runList("../../../pgkit/testdata")

		assert.NoError(t, err)
	})

	t.Run("returns_error_when_directory_is_empty", func(t *testing.T) {
		err := runList("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list migrations")
	})

	t.Run("returns_error_when_directory_does_not_exist", func(t *testing.T) {
		err := runList("nonexistent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list migrations")
	})

	t.Run("succeeds_when_directory_has_no_migrations", func(t *testing.T) {
		err := runList(".")

		assert.NoError(t, err)
	})
}

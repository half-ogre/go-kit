package subcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDBURL(t *testing.T) {
	t.Run("returns_the_flag_value_when_flag_is_set", func(t *testing.T) {
		dbURL = "theDBURL"
		t.Cleanup(func() { dbURL = "" })

		url, err := getDBURL()

		assert.NoError(t, err)
		assert.Equal(t, "theDBURL", url)
	})

	t.Run("returns_environment_variable_when_flag_is_not_set", func(t *testing.T) {
		dbURL = ""
		t.Cleanup(func() { dbURL = "" })
		t.Setenv("DATABASE_URL", "theEnvDBURL")

		url, err := getDBURL()

		assert.NoError(t, err)
		assert.Equal(t, "theEnvDBURL", url)
	})

	t.Run("returns_error_when_neither_flag_nor_environment_variable_is_set", func(t *testing.T) {
		dbURL = ""
		t.Cleanup(func() { dbURL = "" })

		url, err := getDBURL()

		assert.Equal(t, "", url)
		assert.EqualError(t, err, "database URL not provided (use --db flag or DATABASE_URL environment variable)")
	})

	t.Run("prefers_flag_over_environment_variable_when_both_are_set", func(t *testing.T) {
		dbURL = "theFlagDBURL"
		t.Cleanup(func() { dbURL = "" })
		t.Setenv("DATABASE_URL", "anEnvDBURL")

		url, err := getDBURL()

		assert.NoError(t, err)
		assert.Equal(t, "theFlagDBURL", url)
	})
}

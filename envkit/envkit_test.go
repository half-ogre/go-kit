package envkit

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetenvBoolWithDefault(t *testing.T) {
	key := "TEST_BOOL_ENV_VAR"

	t.Run("environment_variable_not_set_returns_default_true", func(t *testing.T) {
		theDefaultValue := true
		os.Unsetenv(key)
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvBoolWithDefault(key, theDefaultValue)

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("environment_variable_not_set_returns_default_false", func(t *testing.T) {
		theDefaultValue := false
		os.Unsetenv(key)
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvBoolWithDefault(key, theDefaultValue)

		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("environment_variable_set_to_true", func(t *testing.T) {
		theDefaultValue := false
		os.Setenv(key, "true")
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvBoolWithDefault(key, theDefaultValue)

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("environment_variable_set_to_false", func(t *testing.T) {
		theDefaultValue := true
		os.Setenv(key, "false")
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvBoolWithDefault(key, theDefaultValue)

		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("environment_variable_set_to_1", func(t *testing.T) {
		theDefaultValue := false
		os.Setenv(key, "1")
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvBoolWithDefault(key, theDefaultValue)

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("environment_variable_set_to_0", func(t *testing.T) {
		theDefaultValue := true
		os.Setenv(key, "0")
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvBoolWithDefault(key, theDefaultValue)

		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("environment_variable_set_to_invalid_value", func(t *testing.T) {
		theDefaultValue := true
		os.Setenv(key, "invalid")
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvBoolWithDefault(key, theDefaultValue)

		assert.Error(t, err)
		assert.False(t, result)
	})
}

func TestGetenvIntWithDefault(t *testing.T) {
	key := "TEST_INT_ENV_VAR"

	t.Run("environment_variable_not_set_returns_default", func(t *testing.T) {
		theDefaultValue := 42
		os.Unsetenv(key)
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvIntWithDefault(key, theDefaultValue)

		assert.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("environment_variable_set_to_positive_integer", func(t *testing.T) {
		theDefaultValue := 42
		os.Setenv(key, "123")
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvIntWithDefault(key, theDefaultValue)

		assert.NoError(t, err)
		assert.Equal(t, 123, result)
	})

	t.Run("environment_variable_set_to_negative_integer", func(t *testing.T) {
		theDefaultValue := 42
		os.Setenv(key, "-456")
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvIntWithDefault(key, theDefaultValue)

		assert.NoError(t, err)
		assert.Equal(t, -456, result)
	})

	t.Run("environment_variable_set_to_zero", func(t *testing.T) {
		theDefaultValue := 42
		os.Setenv(key, "0")
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvIntWithDefault(key, theDefaultValue)

		assert.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("environment_variable_set_to_invalid_value", func(t *testing.T) {
		theDefaultValue := 42
		os.Setenv(key, "invalid")
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvIntWithDefault(key, theDefaultValue)

		assert.Error(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("environment_variable_set_to_float_value", func(t *testing.T) {
		theDefaultValue := 42
		os.Setenv(key, "3.14")
		t.Cleanup(func() { os.Unsetenv(key) })

		result, err := GetenvIntWithDefault(key, theDefaultValue)

		assert.Error(t, err)
		assert.Equal(t, 0, result)
	})
}

func TestGetenvWithDefault(t *testing.T) {
	key := "TEST_STRING_ENV_VAR"

	t.Run("environment_variable_not_set_returns_default", func(t *testing.T) {
		theDefaultValue := "theDefaultValue"
		os.Unsetenv(key)
		t.Cleanup(func() { os.Unsetenv(key) })

		result := GetenvWithDefault(key, theDefaultValue)

		assert.Equal(t, theDefaultValue, result)
	})

	t.Run("environment_variable_set_to_value", func(t *testing.T) {
		theDefaultValue := "theDefaultValue"
		theEnvironmentValue := "theEnvironmentValue"
		os.Setenv(key, theEnvironmentValue)
		t.Cleanup(func() { os.Unsetenv(key) })

		result := GetenvWithDefault(key, theDefaultValue)

		assert.Equal(t, theEnvironmentValue, result)
	})

	t.Run("environment_variable_set_to_empty_string_returns_default", func(t *testing.T) {
		theDefaultValue := "theDefaultValue"
		os.Setenv(key, "")
		t.Cleanup(func() { os.Unsetenv(key) })

		result := GetenvWithDefault(key, theDefaultValue)

		assert.Equal(t, theDefaultValue, result)
	})

	t.Run("environment_variable_with_spaces", func(t *testing.T) {
		theDefaultValue := "theDefaultValue"
		theValueWithSpaces := "value with spaces"
		os.Setenv(key, theValueWithSpaces)
		t.Cleanup(func() { os.Unsetenv(key) })

		result := GetenvWithDefault(key, theDefaultValue)

		assert.Equal(t, theValueWithSpaces, result)
	})

	t.Run("environment_variable_with_special_characters", func(t *testing.T) {
		theDefaultValue := "theDefaultValue"
		theSpecialValue := "value!@#$%^&*()"
		os.Setenv(key, theSpecialValue)
		t.Cleanup(func() { os.Unsetenv(key) })

		result := GetenvWithDefault(key, theDefaultValue)

		assert.Equal(t, theSpecialValue, result)
	})
}

func TestMustGetenv(t *testing.T) {
	t.Run("environment_variable_set_returns_value", func(t *testing.T) {
		key := "TEST_MUST_ENV_VAR"
		theRequiredValue := "theRequiredValue"
		os.Setenv(key, theRequiredValue)
		t.Cleanup(func() { os.Unsetenv(key) })

		result := MustGetenv(key)

		assert.Equal(t, theRequiredValue, result)
	})

	t.Run("environment_variable_not_set_panics", func(t *testing.T) {
		key := "TEST_MUST_ENV_VAR_NOT_SET"
		os.Unsetenv(key)
		t.Cleanup(func() { os.Unsetenv(key) })

		assert.Panics(t, func() {
			MustGetenv(key)
		})
	})

	t.Run("environment_variable_set_to_empty_string_panics", func(t *testing.T) {
		key := "TEST_MUST_ENV_VAR_EMPTY"
		os.Setenv(key, "")
		t.Cleanup(func() { os.Unsetenv(key) })

		assert.Panics(t, func() {
			MustGetenv(key)
		})
	})
}

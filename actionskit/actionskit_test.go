package actionskit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInput(t *testing.T) {
	t.Run("simple_input_name", func(t *testing.T) {
		inputName := "token"
		theTokenValue := "test-token-value"
		envName := "INPUT_" + strings.ToUpper(strings.ReplaceAll(inputName, "-", "_"))
		os.Setenv(envName, theTokenValue)
		t.Cleanup(func() { os.Unsetenv(envName) })

		result := GetInput(inputName)

		assert.Equal(t, theTokenValue, result)
	})

	t.Run("input_name_with_dashes", func(t *testing.T) {
		inputName := "github-token"
		theTokenValue := "ghp_test123"
		envName := "INPUT_GITHUB_TOKEN"
		os.Setenv(envName, theTokenValue)
		t.Cleanup(func() { os.Unsetenv(envName) })

		result := GetInput(inputName)

		assert.Equal(t, theTokenValue, result)
	})

	t.Run("input_name_with_multiple_dashes", func(t *testing.T) {
		inputName := "my-long-input-name"
		theValue := "some-value"
		envName := "INPUT_MY_LONG_INPUT_NAME"
		os.Setenv(envName, theValue)
		t.Cleanup(func() { os.Unsetenv(envName) })

		result := GetInput(inputName)

		assert.Equal(t, theValue, result)
	})

	t.Run("empty_input_value", func(t *testing.T) {
		inputName := "empty-input"
		envName := "INPUT_EMPTY_INPUT"
		os.Setenv(envName, "")
		t.Cleanup(func() { os.Unsetenv(envName) })

		result := GetInput(inputName)

		assert.Empty(t, result)
	})

	t.Run("input_not_set", func(t *testing.T) {
		inputName := "nonexistent-input"

		result := GetInput(inputName)

		assert.Empty(t, result)
	})

	t.Run("input_with_special_characters", func(t *testing.T) {
		inputName := "special-chars"
		theSpecialValue := "value with spaces & symbols!"
		envName := "INPUT_SPECIAL_CHARS"
		os.Setenv(envName, theSpecialValue)
		t.Cleanup(func() { os.Unsetenv(envName) })

		result := GetInput(inputName)

		assert.Equal(t, theSpecialValue, result)
	})
}

func TestGetInputRequired(t *testing.T) {
	t.Run("valid_required_input", func(t *testing.T) {
		inputName := "required-token"
		theTokenValue := "valid-token"
		envName := "INPUT_REQUIRED_TOKEN"
		os.Setenv(envName, theTokenValue)
		t.Cleanup(func() { os.Unsetenv(envName) })

		result, err := GetInputRequired(inputName)

		assert.NoError(t, err)
		assert.Equal(t, theTokenValue, result)
	})

	t.Run("empty_required_input", func(t *testing.T) {
		inputName := "empty-required"
		envName := "INPUT_EMPTY_REQUIRED"
		os.Setenv(envName, "")
		t.Cleanup(func() { os.Unsetenv(envName) })

		result, err := GetInputRequired(inputName)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, "input required and not supplied: "+inputName, err.Error())
	})

	t.Run("missing_required_input", func(t *testing.T) {
		inputName := "missing-required"

		result, err := GetInputRequired(inputName)

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, "input required and not supplied: "+inputName, err.Error())
	})
}

func TestSetOutput(t *testing.T) {
	t.Run("simple_output", func(t *testing.T) {
		outputName := "result"
		theOutputValue := "success"
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "output")
		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
		assert.NoError(t, err)
		file.Close()
		os.Setenv("GITHUB_OUTPUT", outputFile)
		t.Cleanup(func() { os.Unsetenv("GITHUB_OUTPUT") })

		err = SetOutput(outputName, theOutputValue)

		assert.NoError(t, err)
		content, err := os.ReadFile(outputFile)
		assert.NoError(t, err)
		assert.Equal(t, "result=success\n", string(content))
	})

	t.Run("output_with_spaces", func(t *testing.T) {
		outputName := "message"
		theOutputValue := "hello world"
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "output")
		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
		assert.NoError(t, err)
		file.Close()
		os.Setenv("GITHUB_OUTPUT", outputFile)
		t.Cleanup(func() { os.Unsetenv("GITHUB_OUTPUT") })

		err = SetOutput(outputName, theOutputValue)

		assert.NoError(t, err)
		content, err := os.ReadFile(outputFile)
		assert.NoError(t, err)
		assert.Equal(t, "message=hello world\n", string(content))
	})

	t.Run("output_with_special_characters", func(t *testing.T) {
		outputName := "data"
		theOutputValue := "value=with&special!chars"
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "output")
		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
		assert.NoError(t, err)
		file.Close()
		os.Setenv("GITHUB_OUTPUT", outputFile)
		t.Cleanup(func() { os.Unsetenv("GITHUB_OUTPUT") })

		err = SetOutput(outputName, theOutputValue)

		assert.NoError(t, err)
		content, err := os.ReadFile(outputFile)
		assert.NoError(t, err)
		assert.Equal(t, "data=value=with&special!chars\n", string(content))
	})

	t.Run("empty_output_value", func(t *testing.T) {
		outputName := "empty"
		theOutputValue := ""
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "output")
		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
		assert.NoError(t, err)
		file.Close()
		os.Setenv("GITHUB_OUTPUT", outputFile)
		t.Cleanup(func() { os.Unsetenv("GITHUB_OUTPUT") })

		err = SetOutput(outputName, theOutputValue)

		assert.NoError(t, err)
		content, err := os.ReadFile(outputFile)
		assert.NoError(t, err)
		assert.Equal(t, "empty=\n", string(content))
	})

	t.Run("no_output_file_set", func(t *testing.T) {
		outputName := "test"
		theOutputValue := "value"
		os.Unsetenv("GITHUB_OUTPUT")

		err := SetOutput(outputName, theOutputValue)

		assert.NoError(t, err)
	})
}

func TestSetOutputAppend(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output")
	file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
	assert.NoError(t, err)
	file.Close()
	os.Setenv("GITHUB_OUTPUT", outputFile)
	t.Cleanup(func() { os.Unsetenv("GITHUB_OUTPUT") })

	err = SetOutput("first", "value1")
	assert.NoError(t, err)
	err = SetOutput("second", "value2")
	assert.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Equal(t, "first=value1\nsecond=value2\n", string(content))
}

func TestIsDebug(t *testing.T) {
	t.Run("debug_enabled", func(t *testing.T) {
		os.Setenv("RUNNER_DEBUG", "1")
		t.Cleanup(func() { os.Unsetenv("RUNNER_DEBUG") })

		result := IsDebug()

		assert.True(t, result)
	})

	t.Run("debug_disabled_with_0", func(t *testing.T) {
		os.Setenv("RUNNER_DEBUG", "0")
		t.Cleanup(func() { os.Unsetenv("RUNNER_DEBUG") })

		result := IsDebug()

		assert.False(t, result)
	})

	t.Run("debug_disabled_with_false", func(t *testing.T) {
		os.Setenv("RUNNER_DEBUG", "false")
		t.Cleanup(func() { os.Unsetenv("RUNNER_DEBUG") })

		result := IsDebug()

		assert.False(t, result)
	})

	t.Run("debug_disabled_when_not_set", func(t *testing.T) {
		os.Unsetenv("RUNNER_DEBUG")

		result := IsDebug()

		assert.False(t, result)
	})

	t.Run("debug_disabled_with_other_value", func(t *testing.T) {
		os.Setenv("RUNNER_DEBUG", "true")
		t.Cleanup(func() { os.Unsetenv("RUNNER_DEBUG") })

		result := IsDebug()

		assert.False(t, result)
	})
}

func TestLoggingFunctions(t *testing.T) {
	t.Run("Info_doesn't_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Info() panicked: %v", r)
			}
		}()

		Info("test info message")
	})

	t.Run("Warning_doesn't_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Warning() panicked: %v", r)
			}
		}()

		Warning("test warning message")
	})

	t.Run("Error_doesn't_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Error() panicked: %v", r)
			}
		}()

		Error("test error message")
	})

	t.Run("Debug_doesn't_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Debug() panicked: %v", r)
			}
		}()

		os.Unsetenv("RUNNER_DEBUG")
		Debug("test debug message when disabled")
		os.Setenv("RUNNER_DEBUG", "1")
		Debug("test debug message when enabled")
		os.Unsetenv("RUNNER_DEBUG")
	})
}

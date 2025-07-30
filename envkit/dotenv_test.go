package envkit

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("basic_key-value_pairs", func(t *testing.T) {
		input := "KEY1=value1\nKEY2=value2"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "value1", result["KEY1"])
		assert.Equal(t, "value2", result["KEY2"])
	})

	t.Run("key-value_with_spaces", func(t *testing.T) {
		input := "KEY1 = value1\nKEY2=value2"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "value1", result["KEY1"])
		assert.Equal(t, "value2", result["KEY2"])
	})

	t.Run("quoted_values", func(t *testing.T) {
		input := "KEY1=\"value with spaces\"\nKEY2='single quoted'"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "value with spaces", result["KEY1"])
		assert.Equal(t, "single quoted", result["KEY2"])
	})

	t.Run("empty_value", func(t *testing.T) {
		input := "KEY1=\nKEY2=value2"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "", result["KEY1"])
		assert.Equal(t, "value2", result["KEY2"])
	})

	t.Run("comments", func(t *testing.T) {
		input := "# This is a comment\nKEY1=value1\n# Another comment\nKEY2=value2"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "value1", result["KEY1"])
		assert.Equal(t, "value2", result["KEY2"])
	})

	t.Run("inline_comments", func(t *testing.T) {
		input := "KEY1=value1 # inline comment\nKEY2=value2"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "value1", result["KEY1"])
		assert.Equal(t, "value2", result["KEY2"])
	})

	t.Run("export_prefix", func(t *testing.T) {
		input := "export KEY1=value1\nexport KEY2=value2"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "value1", result["KEY1"])
		assert.Equal(t, "value2", result["KEY2"])
	})

	t.Run("multiline_with_escapes", func(t *testing.T) {
		input := "KEY1=\"line1\\nline2\"\nKEY2=\"test\\rtest\""
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "line1\nline2", result["KEY1"])
		assert.Equal(t, "test\rtest", result["KEY2"])
	})

	t.Run("escaped_quotes", func(t *testing.T) {
		input := "KEY1=\"value with \\\" quote\"\nKEY2='value with \\' quote'"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "value with \" quote", result["KEY1"])
		assert.Equal(t, "value with \\' quote", result["KEY2"])
	})

	t.Run("variable_expansion", func(t *testing.T) {
		input := "KEY1=value1\nKEY2=${KEY1}_expanded"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "value1", result["KEY1"])
		assert.Equal(t, "value1_expanded", result["KEY2"])
	})

	t.Run("windows_line_endings", func(t *testing.T) {
		input := "KEY1=value1\r\nKEY2=value2\r\n"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "value1", result["KEY1"])
		assert.Equal(t, "value2", result["KEY2"])
	})

	t.Run("empty_lines", func(t *testing.T) {
		input := "\n\nKEY1=value1\n\n\nKEY2=value2\n\n"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "value1", result["KEY1"])
		assert.Equal(t, "value2", result["KEY2"])
	})

	t.Run("yaml_style", func(t *testing.T) {
		input := "KEY1: value1\nKEY2: value2"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "value1", result["KEY1"])
		assert.Equal(t, "value2", result["KEY2"])
	})

	t.Run("special_characters_in_key", func(t *testing.T) {
		input := "KEY_1=value1\nKEY.2=value2\nKEY3=value3"
		reader := strings.NewReader(input)

		result, err := parse(reader)

		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "value1", result["KEY_1"])
		assert.Equal(t, "value2", result["KEY.2"])
		assert.Equal(t, "value3", result["KEY3"])
	})
}

func TestParseErrors(t *testing.T) {
	t.Run("invalid_character_in_key", func(t *testing.T) {
		input := "KEY-INVALID=value"
		reader := strings.NewReader(input)

		_, err := parse(reader)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected character")
	})

	t.Run("unterminated_double_quote", func(t *testing.T) {
		input := `KEY="unterminated value`
		reader := strings.NewReader(input)

		_, err := parse(reader)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unterminated quoted value")
	})

	t.Run("unterminated_single_quote", func(t *testing.T) {
		input := `KEY='unterminated value`
		reader := strings.NewReader(input)

		_, err := parse(reader)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unterminated quoted value")
	})
}

func TestLoadEnv(t *testing.T) {
	content := "TEST_KEY1=test_value1\nTEST_KEY2=test_value2"
	tmpfile, err := os.CreateTemp("", "test_*.env")
	assert.NoError(t, err)
	t.Cleanup(func() { os.Remove(tmpfile.Name()) })

	_, err = tmpfile.Write([]byte(content))
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	os.Unsetenv("TEST_KEY1")
	os.Unsetenv("TEST_KEY2")
	t.Cleanup(func() {
		os.Unsetenv("TEST_KEY1")
		os.Unsetenv("TEST_KEY2")
	})

	err = LoadEnv(tmpfile.Name())

	assert.NoError(t, err)
	assert.Equal(t, "test_value1", os.Getenv("TEST_KEY1"))
	assert.Equal(t, "test_value2", os.Getenv("TEST_KEY2"))

	// Test that existing environment variables are not overwritten
	os.Setenv("TEST_KEY1", "existing_value")
	err = LoadEnv(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "existing_value", os.Getenv("TEST_KEY1"))
}

func TestLoadEnvFileNotFound(t *testing.T) {
	err := LoadEnv("/non/existent/file.env")

	assert.Error(t, err)
}

func TestExpandVariables(t *testing.T) {
	t.Run("basic_expansion", func(t *testing.T) {
		input := "${KEY1}"
		vars := map[string]string{"KEY1": "value1"}

		result := expandVariables(input, vars)

		assert.Equal(t, "value1", result)
	})

	t.Run("expansion_without_braces", func(t *testing.T) {
		input := "$KEY1"
		vars := map[string]string{"KEY1": "value1"}

		result := expandVariables(input, vars)

		assert.Equal(t, "value1", result)
	})

	t.Run("escaped_variable", func(t *testing.T) {
		input := "\\$KEY1"
		vars := map[string]string{"KEY1": "value1"}

		result := expandVariables(input, vars)

		assert.Equal(t, "$KEY1", result)
	})

	t.Run("multiple_expansions", func(t *testing.T) {
		input := "${KEY1} and ${KEY2}"
		vars := map[string]string{"KEY1": "value1", "KEY2": "value2"}

		result := expandVariables(input, vars)

		assert.Equal(t, "value1 and value2", result)
	})

	t.Run("non-existent_variable", func(t *testing.T) {
		input := "${NONEXISTENT}"
		vars := map[string]string{}

		result := expandVariables(input, vars)

		assert.Equal(t, "", result)
	})

	t.Run("mixed_text_and_variables", func(t *testing.T) {
		input := "prefix_${KEY1}_suffix"
		vars := map[string]string{"KEY1": "value1"}

		result := expandVariables(input, vars)

		assert.Equal(t, "prefix_value1_suffix", result)
	})
}

func TestHasQuotePrefix(t *testing.T) {
	t.Run("double_quote", func(t *testing.T) {
		input := []byte(`"value"`)

		prefix, isQuoted := hasQuotePrefix(input)

		assert.Equal(t, byte('"'), prefix)
		assert.True(t, isQuoted)
	})

	t.Run("single_quote", func(t *testing.T) {
		input := []byte(`'value'`)

		prefix, isQuoted := hasQuotePrefix(input)

		assert.Equal(t, byte('\''), prefix)
		assert.True(t, isQuoted)
	})

	t.Run("no_quote", func(t *testing.T) {
		input := []byte(`value`)

		prefix, isQuoted := hasQuotePrefix(input)

		assert.Equal(t, byte(0), prefix)
		assert.False(t, isQuoted)
	})

	t.Run("empty_input", func(t *testing.T) {
		input := []byte{}

		prefix, isQuoted := hasQuotePrefix(input)

		assert.Equal(t, byte(0), prefix)
		assert.False(t, isQuoted)
	})
}

func TestExpandEscapes(t *testing.T) {
	t.Run("newline_escape", func(t *testing.T) {
		input := `line1\nline2`

		result := expandEscapes(input)

		assert.Equal(t, "line1\nline2", result)
	})

	t.Run("carriage_return_escape", func(t *testing.T) {
		input := `line1\rline2`

		result := expandEscapes(input)

		assert.Equal(t, "line1\rline2", result)
	})

	t.Run("escaped_backslash", func(t *testing.T) {
		input := `path\\to\\file`

		result := expandEscapes(input)

		assert.Equal(t, `path\to\file`, result)
	})

	t.Run("no_escapes", func(t *testing.T) {
		input := `plain text`

		result := expandEscapes(input)

		assert.Equal(t, `plain text`, result)
	})

	t.Run("mixed_escapes", func(t *testing.T) {
		input := `line1\nline2\rreturn`

		result := expandEscapes(input)

		assert.Equal(t, "line1\nline2\rreturn", result)
	})
}

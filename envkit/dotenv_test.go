package envkit

import (
	"os"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "basic key-value pairs",
			input:    "KEY1=value1\nKEY2=value2",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "key-value with spaces",
			input:    "KEY1 = value1\nKEY2=value2",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "quoted values",
			input:    "KEY1=\"value with spaces\"\nKEY2='single quoted'",
			expected: map[string]string{"KEY1": "value with spaces", "KEY2": "single quoted"},
		},
		{
			name:     "empty value",
			input:    "KEY1=\nKEY2=value2",
			expected: map[string]string{"KEY1": "", "KEY2": "value2"},
		},
		{
			name:     "comments",
			input:    "# This is a comment\nKEY1=value1\n# Another comment\nKEY2=value2",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "inline comments",
			input:    "KEY1=value1 # inline comment\nKEY2=value2",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "export prefix",
			input:    "export KEY1=value1\nexport KEY2=value2",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "multiline with escapes",
			input:    "KEY1=\"line1\\nline2\"\nKEY2=\"test\\rtest\"",
			expected: map[string]string{"KEY1": "line1\nline2", "KEY2": "test\rtest"},
		},
		{
			name:     "escaped quotes",
			input:    "KEY1=\"value with \\\" quote\"\nKEY2='value with \\' quote'",
			expected: map[string]string{"KEY1": "value with \" quote", "KEY2": "value with \\' quote"},
		},
		{
			name:     "variable expansion",
			input:    "KEY1=value1\nKEY2=${KEY1}_expanded",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value1_expanded"},
		},
		{
			name:     "windows line endings",
			input:    "KEY1=value1\r\nKEY2=value2\r\n",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "empty lines",
			input:    "\n\nKEY1=value1\n\n\nKEY2=value2\n\n",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "yaml style",
			input:    "KEY1: value1\nKEY2: value2",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "special characters in key",
			input:    "KEY_1=value1\nKEY.2=value2\nKEY3=value3",
			expected: map[string]string{"KEY_1": "value1", "KEY.2": "value2", "KEY3": "value3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := parse(reader)
			if err != nil {
				t.Fatalf("parse() error = %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("parse() returned %d values, want %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if value, ok := result[key]; !ok {
					t.Errorf("parse() missing key %q", key)
				} else if value != expectedValue {
					t.Errorf("parse() key %q = %q, want %q", key, value, expectedValue)
				}
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "invalid character in key",
			input:       "KEY-INVALID=value",
			expectedErr: "unexpected character",
		},
		{
			name:        "unterminated double quote",
			input:       `KEY="unterminated value`,
			expectedErr: "unterminated quoted value",
		},
		{
			name:        "unterminated single quote",
			input:       `KEY='unterminated value`,
			expectedErr: "unterminated quoted value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			_, err := parse(reader)
			if err == nil {
				t.Fatal("parse() expected error but got nil")
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("parse() error = %v, want error containing %q", err, tt.expectedErr)
			}
		})
	}
}

func TestLoadEnv(t *testing.T) {
	// Create a temporary .env file
	content := "TEST_KEY1=test_value1\nTEST_KEY2=test_value2"
	tmpfile, err := os.CreateTemp("", "test_*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Clear any existing TEST_KEY variables
	os.Unsetenv("TEST_KEY1")
	os.Unsetenv("TEST_KEY2")

	// Test loading the file
	err = LoadEnv(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadEnv() error = %v", err)
	}

	// Check that environment variables were set
	if value := os.Getenv("TEST_KEY1"); value != "test_value1" {
		t.Errorf("LoadEnv() TEST_KEY1 = %q, want %q", value, "test_value1")
	}
	if value := os.Getenv("TEST_KEY2"); value != "test_value2" {
		t.Errorf("LoadEnv() TEST_KEY2 = %q, want %q", value, "test_value2")
	}

	// Test that existing environment variables are not overwritten
	os.Setenv("TEST_KEY1", "existing_value")

	err = LoadEnv(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadEnv() error = %v", err)
	}

	if value := os.Getenv("TEST_KEY1"); value != "existing_value" {
		t.Errorf("LoadEnv() overwrote existing TEST_KEY1 = %q, want %q", value, "existing_value")
	}

	// Clean up
	os.Unsetenv("TEST_KEY1")
	os.Unsetenv("TEST_KEY2")
}

func TestLoadEnvFileNotFound(t *testing.T) {
	err := LoadEnv("/non/existent/file.env")
	if err == nil {
		t.Error("LoadEnv() expected error for non-existent file but got nil")
	}
}

func TestExpandVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		vars     map[string]string
		expected string
	}{
		{
			name:     "basic expansion",
			input:    "${KEY1}",
			vars:     map[string]string{"KEY1": "value1"},
			expected: "value1",
		},
		{
			name:     "expansion without braces",
			input:    "$KEY1",
			vars:     map[string]string{"KEY1": "value1"},
			expected: "value1",
		},
		{
			name:     "escaped variable",
			input:    "\\$KEY1",
			vars:     map[string]string{"KEY1": "value1"},
			expected: "$KEY1",
		},
		{
			name:     "multiple expansions",
			input:    "${KEY1} and ${KEY2}",
			vars:     map[string]string{"KEY1": "value1", "KEY2": "value2"},
			expected: "value1 and value2",
		},
		{
			name:     "non-existent variable",
			input:    "${NONEXISTENT}",
			vars:     map[string]string{},
			expected: "",
		},
		{
			name:     "mixed text and variables",
			input:    "prefix_${KEY1}_suffix",
			vars:     map[string]string{"KEY1": "value1"},
			expected: "prefix_value1_suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandVariables(tt.input, tt.vars)
			if result != tt.expected {
				t.Errorf("expandVariables() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestHasQuotePrefix(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		expectedPrefix byte
		expectedBool   bool
	}{
		{
			name:           "double quote",
			input:          []byte(`"value"`),
			expectedPrefix: '"',
			expectedBool:   true,
		},
		{
			name:           "single quote",
			input:          []byte(`'value'`),
			expectedPrefix: '\'',
			expectedBool:   true,
		},
		{
			name:           "no quote",
			input:          []byte(`value`),
			expectedPrefix: 0,
			expectedBool:   false,
		},
		{
			name:           "empty input",
			input:          []byte{},
			expectedPrefix: 0,
			expectedBool:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, isQuoted := hasQuotePrefix(tt.input)
			if prefix != tt.expectedPrefix {
				t.Errorf("hasQuotePrefix() prefix = %v, want %v", prefix, tt.expectedPrefix)
			}
			if isQuoted != tt.expectedBool {
				t.Errorf("hasQuotePrefix() isQuoted = %v, want %v", isQuoted, tt.expectedBool)
			}
		})
	}
}

func TestExpandEscapes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "newline escape",
			input:    `line1\nline2`,
			expected: "line1\nline2",
		},
		{
			name:     "carriage return escape",
			input:    `line1\rline2`,
			expected: "line1\rline2",
		},
		{
			name:     "escaped backslash",
			input:    `path\\to\\file`,
			expected: `path\to\file`,
		},
		{
			name:     "no escapes",
			input:    `plain text`,
			expected: `plain text`,
		},
		{
			name:     "mixed escapes",
			input:    `line1\nline2\rreturn`,
			expected: "line1\nline2\rreturn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEscapes(tt.input)
			if result != tt.expected {
				t.Errorf("expandEscapes() = %q, want %q", result, tt.expected)
			}
		})
	}
}

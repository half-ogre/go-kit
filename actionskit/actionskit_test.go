package actionskit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetInput(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		envValue    string
		expected    string
	}{
		{
			name:      "simple input name",
			inputName: "token",
			envValue:  "test-token-value",
			expected:  "test-token-value",
		},
		{
			name:      "input name with dashes",
			inputName: "github-token",
			envValue:  "ghp_test123",
			expected:  "ghp_test123",
		},
		{
			name:      "input name with multiple dashes",
			inputName: "my-long-input-name",
			envValue:  "some-value",
			expected:  "some-value",
		},
		{
			name:      "empty input value",
			inputName: "empty-input",
			envValue:  "",
			expected:  "",
		},
		{
			name:      "input not set",
			inputName: "nonexistent-input",
			envValue:  "", // Will not be set
			expected:  "",
		},
		{
			name:      "input with special characters",
			inputName: "special-chars",
			envValue:  "value with spaces & symbols!",
			expected:  "value with spaces & symbols!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable
			envName := "INPUT_" + strings.ToUpper(strings.ReplaceAll(tt.inputName, "-", "_"))
			if tt.name != "input not set" {
				os.Setenv(envName, tt.envValue)
				defer os.Unsetenv(envName)
			}

			// Test GetInput
			result := GetInput(tt.inputName)
			if result != tt.expected {
				t.Errorf("GetInput(%q) = %q, want %q", tt.inputName, result, tt.expected)
			}
		})
	}
}

func TestGetInputRequired(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		envValue    string
		shouldError bool
		expected    string
	}{
		{
			name:        "valid required input",
			inputName:   "required-token",
			envValue:    "valid-token",
			shouldError: false,
			expected:    "valid-token",
		},
		{
			name:        "empty required input",
			inputName:   "empty-required",
			envValue:    "",
			shouldError: true,
			expected:    "",
		},
		{
			name:        "missing required input",
			inputName:   "missing-required",
			envValue:    "", // Will not be set
			shouldError: true,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable
			envName := "INPUT_" + strings.ToUpper(strings.ReplaceAll(tt.inputName, "-", "_"))
			if tt.name != "missing required input" {
				os.Setenv(envName, tt.envValue)
				defer os.Unsetenv(envName)
			}

			// Test GetInputRequired
			result, err := GetInputRequired(tt.inputName)

			// Check error expectation
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check result
			if result != tt.expected {
				t.Errorf("GetInputRequired(%q) = %q, want %q", tt.inputName, result, tt.expected)
			}

			// Check error message format
			if tt.shouldError && err != nil {
				expectedErrorMsg := "input required and not supplied: " + tt.inputName
				if err.Error() != expectedErrorMsg {
					t.Errorf("Error message = %q, want %q", err.Error(), expectedErrorMsg)
				}
			}
		})
	}
}

func TestSetOutput(t *testing.T) {
	tests := []struct {
		name           string
		outputName     string
		outputValue    string
		useOutputFile  bool
		expectedInFile string
	}{
		{
			name:           "simple output",
			outputName:     "result",
			outputValue:    "success",
			useOutputFile:  true,
			expectedInFile: "result=success\n",
		},
		{
			name:           "output with spaces",
			outputName:     "message",
			outputValue:    "hello world",
			useOutputFile:  true,
			expectedInFile: "message=hello world\n",
		},
		{
			name:           "output with special characters",
			outputName:     "data",
			outputValue:    "value=with&special!chars",
			useOutputFile:  true,
			expectedInFile: "data=value=with&special!chars\n",
		},
		{
			name:           "empty output value",
			outputName:     "empty",
			outputValue:    "",
			useOutputFile:  true,
			expectedInFile: "empty=\n",
		},
		{
			name:           "no output file set",
			outputName:     "test",
			outputValue:    "value",
			useOutputFile:  false,
			expectedInFile: "", // No file to check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.useOutputFile {
				// Create temporary file for GitHub Actions output
				tempDir := t.TempDir()
				outputFile := filepath.Join(tempDir, "output")
				
				// Create the file with proper permissions
				file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					t.Fatalf("Failed to create output file: %v", err)
				}
				file.Close()

				os.Setenv("GITHUB_OUTPUT", outputFile)
				defer os.Unsetenv("GITHUB_OUTPUT")

				// Test SetOutput
				err = SetOutput(tt.outputName, tt.outputValue)
				if err != nil {
					t.Errorf("SetOutput(%q, %q) returned error: %v", tt.outputName, tt.outputValue, err)
				}

				// Verify file contents
				content, err := os.ReadFile(outputFile)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				if string(content) != tt.expectedInFile {
					t.Errorf("Output file content = %q, want %q", string(content), tt.expectedInFile)
				}
			} else {
				// Test without GITHUB_OUTPUT set (should not error)
				os.Unsetenv("GITHUB_OUTPUT")
				
				err := SetOutput(tt.outputName, tt.outputValue)
				if err != nil {
					t.Errorf("SetOutput(%q, %q) returned error: %v", tt.outputName, tt.outputValue, err)
				}
			}
		})
	}
}

func TestSetOutputAppend(t *testing.T) {
	// Test that multiple SetOutput calls append to the file
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output")
	
	// Create the file
	file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	file.Close()

	os.Setenv("GITHUB_OUTPUT", outputFile)
	defer os.Unsetenv("GITHUB_OUTPUT")

	// Make multiple SetOutput calls
	err = SetOutput("first", "value1")
	if err != nil {
		t.Errorf("First SetOutput failed: %v", err)
	}

	err = SetOutput("second", "value2")
	if err != nil {
		t.Errorf("Second SetOutput failed: %v", err)
	}

	// Verify both outputs are in the file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expected := "first=value1\nsecond=value2\n"
	if string(content) != expected {
		t.Errorf("Output file content = %q, want %q", string(content), expected)
	}
}

func TestIsDebug(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "debug enabled",
			envValue: "1",
			expected: true,
		},
		{
			name:     "debug disabled with 0",
			envValue: "0",
			expected: false,
		},
		{
			name:     "debug disabled with false",
			envValue: "false",
			expected: false,
		},
		{
			name:     "debug disabled when not set",
			envValue: "", // Will not be set
			expected: false,
		},
		{
			name:     "debug disabled with other value",
			envValue: "true",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name != "debug disabled when not set" {
				os.Setenv("RUNNER_DEBUG", tt.envValue)
				defer os.Unsetenv("RUNNER_DEBUG")
			} else {
				os.Unsetenv("RUNNER_DEBUG")
			}

			result := IsDebug()
			if result != tt.expected {
				t.Errorf("IsDebug() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Note: Testing Info, Debug, Warning, and Error functions would require
// capturing stdout/stderr, which is more complex. In a real-world scenario,
// you might want to make these functions accept io.Writer parameters for easier testing.
// For now, we'll test that they don't panic.

func TestLoggingFunctions(t *testing.T) {
	// Test that logging functions don't panic
	t.Run("Info doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Info() panicked: %v", r)
			}
		}()
		Info("test info message")
	})

	t.Run("Warning doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Warning() panicked: %v", r)
			}
		}()
		Warning("test warning message")
	})

	t.Run("Error doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Error() panicked: %v", r)
			}
		}()
		Error("test error message")
	})

	t.Run("Debug doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Debug() panicked: %v", r)
			}
		}()
		
		// Test with debug disabled
		os.Unsetenv("RUNNER_DEBUG")
		Debug("test debug message when disabled")
		
		// Test with debug enabled
		os.Setenv("RUNNER_DEBUG", "1")
		Debug("test debug message when enabled")
		os.Unsetenv("RUNNER_DEBUG")
	})
}
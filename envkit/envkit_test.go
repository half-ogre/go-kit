package envkit

import (
	"os"
	"testing"
)

func TestGetenvBoolWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expectedBool bool
		expectError  bool
	}{
		{
			name:         "environment variable not set returns default true",
			envValue:     "",
			defaultValue: true,
			expectedBool: true,
			expectError:  false,
		},
		{
			name:         "environment variable not set returns default false",
			envValue:     "",
			defaultValue: false,
			expectedBool: false,
			expectError:  false,
		},
		{
			name:         "environment variable set to true",
			envValue:     "true",
			defaultValue: false,
			expectedBool: true,
			expectError:  false,
		},
		{
			name:         "environment variable set to false",
			envValue:     "false",
			defaultValue: true,
			expectedBool: false,
			expectError:  false,
		},
		{
			name:         "environment variable set to 1",
			envValue:     "1",
			defaultValue: false,
			expectedBool: true,
			expectError:  false,
		},
		{
			name:         "environment variable set to 0",
			envValue:     "0",
			defaultValue: true,
			expectedBool: false,
			expectError:  false,
		},
		{
			name:         "environment variable set to invalid value",
			envValue:     "invalid",
			defaultValue: true,
			expectedBool: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variable before and after test
			key := "TEST_BOOL_ENV_VAR"
			defer os.Unsetenv(key)
			
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			} else {
				os.Unsetenv(key)
			}

			result, err := GetenvBoolWithDefault(key, tt.defaultValue)

			if tt.expectError {
				if err == nil {
					t.Error("GetenvBoolWithDefault() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetenvBoolWithDefault() unexpected error: %v", err)
				}
				if result != tt.expectedBool {
					t.Errorf("GetenvBoolWithDefault() = %v, want %v", result, tt.expectedBool)
				}
			}
		})
	}
}

func TestGetenvIntWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		expectedInt  int
		expectError  bool
	}{
		{
			name:         "environment variable not set returns default",
			envValue:     "",
			defaultValue: 42,
			expectedInt:  42,
			expectError:  false,
		},
		{
			name:         "environment variable set to positive integer",
			envValue:     "123",
			defaultValue: 42,
			expectedInt:  123,
			expectError:  false,
		},
		{
			name:         "environment variable set to negative integer",
			envValue:     "-456",
			defaultValue: 42,
			expectedInt:  -456,
			expectError:  false,
		},
		{
			name:         "environment variable set to zero",
			envValue:     "0",
			defaultValue: 42,
			expectedInt:  0,
			expectError:  false,
		},
		{
			name:         "environment variable set to invalid value",
			envValue:     "invalid",
			defaultValue: 42,
			expectedInt:  0,
			expectError:  true,
		},
		{
			name:         "environment variable set to float value",
			envValue:     "3.14",
			defaultValue: 42,
			expectedInt:  0,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variable before and after test
			key := "TEST_INT_ENV_VAR"
			defer os.Unsetenv(key)
			
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			} else {
				os.Unsetenv(key)
			}

			result, err := GetenvIntWithDefault(key, tt.defaultValue)

			if tt.expectError {
				if err == nil {
					t.Error("GetenvIntWithDefault() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetenvIntWithDefault() unexpected error: %v", err)
				}
				if result != tt.expectedInt {
					t.Errorf("GetenvIntWithDefault() = %v, want %v", result, tt.expectedInt)
				}
			}
		})
	}
}

func TestGetenvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable not set returns default",
			envValue:     "",
			defaultValue: "theDefaultValue",
			expected:     "theDefaultValue",
		},
		{
			name:         "environment variable set to value",
			envValue:     "theEnvironmentValue",
			defaultValue: "theDefaultValue",
			expected:     "theEnvironmentValue",
		},
		{
			name:         "environment variable set to empty string returns default",
			envValue:     "",
			defaultValue: "theDefaultValue",
			expected:     "theDefaultValue",
		},
		{
			name:         "environment variable with spaces",
			envValue:     "value with spaces",
			defaultValue: "theDefaultValue",
			expected:     "value with spaces",
		},
		{
			name:         "environment variable with special characters",
			envValue:     "value!@#$%^&*()",
			defaultValue: "theDefaultValue",
			expected:     "value!@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variable before and after test
			key := "TEST_STRING_ENV_VAR"
			defer os.Unsetenv(key)
			
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			} else {
				os.Unsetenv(key)
			}

			result := GetenvWithDefault(key, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("GetenvWithDefault() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMustGetenv(t *testing.T) {
	t.Run("environment variable set returns value", func(t *testing.T) {
		key := "TEST_MUST_ENV_VAR"
		expectedValue := "theRequiredValue"
		
		defer os.Unsetenv(key)
		os.Setenv(key, expectedValue)

		result := MustGetenv(key)

		if result != expectedValue {
			t.Errorf("MustGetenv() = %q, want %q", result, expectedValue)
		}
	})

	t.Run("environment variable not set panics", func(t *testing.T) {
		key := "TEST_MUST_ENV_VAR_NOT_SET"
		
		defer os.Unsetenv(key)
		os.Unsetenv(key) // Ensure it's not set

		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGetenv() expected panic but got none")
			} else {
				expectedMessage := "environment variable TEST_MUST_ENV_VAR_NOT_SET not set"
				if r != expectedMessage {
					t.Errorf("MustGetenv() panic message = %q, want %q", r, expectedMessage)
				}
			}
		}()

		MustGetenv(key)
	})

	t.Run("environment variable set to empty string panics", func(t *testing.T) {
		key := "TEST_MUST_ENV_VAR_EMPTY"
		
		defer os.Unsetenv(key)
		os.Setenv(key, "")

		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGetenv() expected panic but got none")
			}
		}()

		MustGetenv(key)
	})
}
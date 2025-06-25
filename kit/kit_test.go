package kit

import (
	"errors"
	"fmt"
	"testing"
)

func TestMapValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected []string
	}{
		{
			name:     "an empty map",
			input:    map[string]string{},
			expected: []string{},
		},
		{
			name: "a single element",
			input: map[string]string{
				"aFirstKey": "theFirstValue",
			},
			expected: []string{"theFirstValue"},
		},
		{
			name: "multiple elements",
			input: map[string]string{
				"aFirstKey":  "theFirstValue",
				"aSecondKey": "theSecondValue",
				"aThirdKey":  "theThirdValue",
			},
			expected: []string{"theFirstValue", "theSecondValue", "theThirdValue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapValues(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("MapValues() returned %d values, want %d", len(result), len(tt.expected))
				return
			}

			// Create a map to count occurrences since order is not guaranteed
			resultCount := make(map[string]int)
			for _, v := range result {
				resultCount[v]++
			}

			expectedCount := make(map[string]int)
			for _, v := range tt.expected {
				expectedCount[v]++
			}

			for k, v := range expectedCount {
				if resultCount[k] != v {
					t.Errorf("MapValues() value %s appears %d times, want %d times", k, resultCount[k], v)
				}
			}
		})
	}
}

func TestMapValuesWithDifferentTypes(t *testing.T) {
	// Test with string values
	stringMap := map[int]string{
		1: "theFirstValue",
		2: "theSecondValue",
		3: "theThirdValue",
	}

	stringResult := MapValues(stringMap)
	if len(stringResult) != 3 {
		t.Errorf("MapValues() with strings returned %d values, want 3", len(stringResult))
	}

	// Test with struct values
	type person struct {
		name string
		age  int
	}

	personMap := map[string]person{
		"aFirstKey":  {name: "aFirstName", age: 1},
		"aSecondKey": {name: "aSecondName", age: 3},
	}

	personResult := MapValues(personMap)
	if len(personResult) != 2 {
		t.Errorf("MapValues() with structs returned %d values, want 2", len(personResult))
	}
}

func TestWrapError(t *testing.T) {
	baseErr := errors.New("theError")

	tests := []struct {
		name           string
		err            error
		format         string
		args           []any
		expectedPrefix string
	}{
		{
			name:           "wrap without formatting",
			err:            baseErr,
			format:         "theMessageWithoutFormatting",
			args:           []any{},
			expectedPrefix: "theMessageWithoutFormatting: theError",
		},
		{
			name:           "wrap with formatting",
			err:            baseErr,
			format:         "theMessageWithFormatting of %d",
			args:           []any{42},
			expectedPrefix: "theMessageWithFormatting of 42: theError",
		},
		{
			name:           "wrap with multiple format args",
			err:            baseErr,
			format:         "theMessageWithFormatting of %d and %s",
			args:           []any{42, "theValue"},
			expectedPrefix: "theMessageWithFormatting of 42 and theValue: theError",
		},
		{
			name:           "wrap with nil error",
			err:            nil,
			format:         "theMessage",
			args:           []any{},
			expectedPrefix: "theMessage: %!w(<nil>)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.err, tt.format, tt.args...)

			if result == nil {
				t.Error("WrapError() returned nil")
				return
			}

			resultStr := result.Error()
			if resultStr != tt.expectedPrefix {
				t.Errorf("WrapError() = %q, want %q", resultStr, tt.expectedPrefix)
			}

			// Test error unwrapping for non-nil base errors
			if tt.err != nil {
				if !errors.Is(result, tt.err) {
					t.Errorf("WrapError() result does not wrap the original error")
				}
			}
		})
	}
}

func TestWrapErrorChaining(t *testing.T) {
	err1 := errors.New("original error")
	err2 := WrapError(err1, "first wrap")
	err3 := WrapError(err2, "second wrap")

	expected := "second wrap: first wrap: original error"
	if err3.Error() != expected {
		t.Errorf("Chained WrapError() = %q, want %q", err3.Error(), expected)
	}

	// Verify error chain
	if !errors.Is(err3, err1) {
		t.Error("Chained WrapError() does not properly wrap the original error")
	}
}

func ExampleMapValues() {
	ages := map[string]int{
		"aFirstName":  1,
		"aSecondName": 3,
		"aThirdName":  5,
	}

	values := MapValues(ages)
	fmt.Println(len(values)) // Order not guaranteed, just print length
	// Output: 3
}

func ExampleWrapError() {
	err := errors.New("file not found")
	wrapped := WrapError(err, "failed to load config")
	fmt.Println(wrapped.Error())
	// Output: failed to load config: file not found
}

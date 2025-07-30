package kit

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapValues(t *testing.T) {
	t.Run("an_empty_map", func(t *testing.T) {
		input := map[string]string{}

		result := MapValues(input)

		assert.Empty(t, result)
	})

	t.Run("a_single_element", func(t *testing.T) {
		input := map[string]string{"aFirstKey": "theFirstValue"}

		result := MapValues(input)

		assert.Len(t, result, 1)
		assert.Contains(t, result, "theFirstValue")
	})

	t.Run("multiple_elements", func(t *testing.T) {
		input := map[string]string{
			"aFirstKey":  "theFirstValue",
			"aSecondKey": "theSecondValue",
			"aThirdKey":  "theThirdValue",
		}

		result := MapValues(input)

		assert.Len(t, result, 3)
		assert.Contains(t, result, "theFirstValue")
		assert.Contains(t, result, "theSecondValue")
		assert.Contains(t, result, "theThirdValue")
	})
}

func TestMapValuesWithDifferentTypes(t *testing.T) {
	// Test with string values
	stringMap := map[int]string{
		1: "theFirstValue",
		2: "theSecondValue",
		3: "theThirdValue",
	}

	stringResult := MapValues(stringMap)
	assert.Len(t, stringResult, 3)

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
	assert.Len(t, personResult, 2)
}

func TestWrapError(t *testing.T) {
	t.Run("wrap_without_formatting", func(t *testing.T) {
		theError := errors.New("theError")

		result := WrapError(theError, "theMessageWithoutFormatting")

		assert.NotNil(t, result)
		assert.Equal(t, "theMessageWithoutFormatting: theError", result.Error())
		assert.True(t, errors.Is(result, theError))
	})

	t.Run("wrap_with_formatting", func(t *testing.T) {
		theError := errors.New("theError")
		format := "theMessageWithFormatting of %d"

		result := WrapError(theError, format, 42)

		assert.NotNil(t, result)
		assert.Equal(t, "theMessageWithFormatting of 42: theError", result.Error())
		assert.True(t, errors.Is(result, theError))
	})

	t.Run("wrap_with_multiple_format_args", func(t *testing.T) {
		theError := errors.New("theError")
		format := "theMessageWithFormatting of %d and %s"

		result := WrapError(theError, format, 42, "theValue")

		assert.NotNil(t, result)
		assert.Equal(t, "theMessageWithFormatting of 42 and theValue: theError", result.Error())
		assert.True(t, errors.Is(result, theError))
	})

	t.Run("wrap_with_nil_error", func(t *testing.T) {

		result := WrapError(nil, "theMessage")

		assert.NotNil(t, result)
		assert.Equal(t, "theMessage: %!w(<nil>)", result.Error())
	})
}

func TestWrapErrorChaining(t *testing.T) {
	err1 := errors.New("original error")
	err2 := WrapError(err1, "first wrap")

	err3 := WrapError(err2, "second wrap")

	assert.Equal(t, "second wrap: first wrap: original error", err3.Error())
	assert.True(t, errors.Is(err3, err1))
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

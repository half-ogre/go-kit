package echokit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	t.Run("returns_a_validator_instance", func(t *testing.T) {
		validator := NewValidator()

		assert.NotNil(t, validator)
		assert.NotNil(t, validator.validator)
	})
}

func TestValidate(t *testing.T) {
	validator := NewValidator()

	t.Run("returns_no_error_when_all_required_fields_are_present", func(t *testing.T) {
		type TestStruct struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
			Age   int    `validate:"required,min=0"`
		}
		data := TestStruct{
			Name:  "the name",
			Email: "the@email.com",
			Age:   25,
		}

		err := validator.Validate(data)

		assert.NoError(t, err)
	})

	t.Run("returns_error_when_required_field_is_missing", func(t *testing.T) {
		type TestStruct struct {
			Name string `validate:"required"`
		}
		data := TestStruct{
			Name: "",
		}

		err := validator.Validate(data)

		assert.Error(t, err)
	})

	t.Run("returns_error_when_email_format_is_invalid", func(t *testing.T) {
		type TestStruct struct {
			Email string `validate:"email"`
		}
		data := TestStruct{
			Email: "invalid-email",
		}

		err := validator.Validate(data)

		assert.Error(t, err)
	})

	t.Run("returns_error_when_min_validation_fails", func(t *testing.T) {
		type TestStruct struct {
			Age int `validate:"min=0"`
		}
		data := TestStruct{
			Age: -1,
		}

		err := validator.Validate(data)

		assert.Error(t, err)
	})

	t.Run("returns_error_when_max_validation_fails", func(t *testing.T) {
		type TestStruct struct {
			Age int `validate:"max=100"`
		}
		data := TestStruct{
			Age: 101,
		}

		err := validator.Validate(data)

		assert.Error(t, err)
	})
}

func TestCustomISODateValidator(t *testing.T) {
	validator := NewValidator()

	t.Run("returns_no_error_when_date_format_is_valid", func(t *testing.T) {
		type TestStruct struct {
			Date string `validate:"isodate"`
		}
		data := TestStruct{
			Date: "2025-11-07",
		}

		err := validator.Validate(data)

		assert.NoError(t, err)
	})

	t.Run("returns_error_when_date_format_is_invalid", func(t *testing.T) {
		type TestStruct struct {
			Date string `validate:"isodate"`
		}
		data := TestStruct{
			Date: "invalid-date",
		}

		err := validator.Validate(data)

		assert.Error(t, err)
	})

	t.Run("returns_error_when_date_has_wrong_separator", func(t *testing.T) {
		type TestStruct struct {
			Date string `validate:"isodate"`
		}
		data := TestStruct{
			Date: "2025/11/07",
		}

		err := validator.Validate(data)

		assert.Error(t, err)
	})

	t.Run("returns_error_when_date_includes_time", func(t *testing.T) {
		type TestStruct struct {
			Date string `validate:"isodate"`
		}
		data := TestStruct{
			Date: "2025-11-07T10:30:00Z",
		}

		err := validator.Validate(data)

		assert.Error(t, err)
	})

	t.Run("returns_error_when_month_is_invalid", func(t *testing.T) {
		type TestStruct struct {
			Date string `validate:"isodate"`
		}
		data := TestStruct{
			Date: "2025-13-07",
		}

		err := validator.Validate(data)

		assert.Error(t, err)
	})

	t.Run("returns_error_when_day_is_invalid", func(t *testing.T) {
		type TestStruct struct {
			Date string `validate:"isodate"`
		}
		data := TestStruct{
			Date: "2025-11-32",
		}

		err := validator.Validate(data)

		assert.Error(t, err)
	})

	t.Run("handles_leap_year_correctly", func(t *testing.T) {
		type TestStruct struct {
			Date string `validate:"isodate"`
		}
		data := TestStruct{
			Date: "2024-02-29",
		}

		err := validator.Validate(data)

		assert.NoError(t, err)
	})

	t.Run("returns_error_for_invalid_leap_year_date", func(t *testing.T) {
		type TestStruct struct {
			Date string `validate:"isodate"`
		}
		data := TestStruct{
			Date: "2025-02-29",
		}

		err := validator.Validate(data)

		assert.Error(t, err)
	})

	t.Run("returns_no_error_when_date_field_is_empty_and_not_required", func(t *testing.T) {
		type TestStruct struct {
			Date string `validate:"omitempty,isodate"`
		}
		data := TestStruct{
			Date: "",
		}

		err := validator.Validate(data)

		assert.NoError(t, err)
	})

	t.Run("returns_error_when_date_is_empty_and_required", func(t *testing.T) {
		type TestStruct struct {
			Date string `validate:"required,isodate"`
		}
		data := TestStruct{
			Date: "",
		}

		err := validator.Validate(data)

		assert.Error(t, err)
	})
}

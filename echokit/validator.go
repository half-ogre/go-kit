package echokit

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the validator for Echo
type CustomValidator struct {
	validator *validator.Validate
}

// Validate implements Echo's Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// NewValidator creates a new validator instance
func NewValidator() *CustomValidator {
	v := validator.New()

	// Register custom isodate validator for YYYY-MM-DD format
	v.RegisterValidation("isodate", func(fl validator.FieldLevel) bool {
		dateStr := fl.Field().String()
		_, err := time.Parse("2006-01-02", dateStr)
		return err == nil
	})

	return &CustomValidator{
		validator: v,
	}
}

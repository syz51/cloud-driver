package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator with custom functionality
type Validator struct {
	validator *validator.Validate
}

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface for ValidationErrors
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// New creates a new validator instance with custom configurations
func New() *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())

	// Register custom tag name function to use json tags for field names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	registerCustomValidators(v)

	return &Validator{validator: v}
}

// registerCustomValidators adds custom validation rules
func registerCustomValidators(v *validator.Validate) {
	// Custom validator for 115drive credentials format
	v.RegisterValidation("drive115_id", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return len(value) > 0 && !strings.Contains(value, " ")
	})

	// Custom validator for URL arrays
	v.RegisterValidation("urls", func(fl validator.FieldLevel) bool {
		urls := fl.Field().Interface().([]string)
		for _, url := range urls {
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "magnet:") {
				return false
			}
		}
		return true
	})
}

// ValidateStruct validates a struct and returns detailed error information
func (v *Validator) ValidateStruct(s interface{}) error {
	err := v.validator.Struct(s)
	if err != nil {
		return v.formatValidationErrors(err)
	}
	return nil
}

// formatValidationErrors converts validator errors to our custom format
func (v *Validator) formatValidationErrors(err error) ValidationErrors {
	var validationErrors []ValidationError

	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			validationError := ValidationError{
				Field:   fieldError.Field(),
				Tag:     fieldError.Tag(),
				Value:   fmt.Sprintf("%v", fieldError.Value()),
				Message: v.getErrorMessage(fieldError),
			}
			validationErrors = append(validationErrors, validationError)
		}
	}

	return ValidationErrors{Errors: validationErrors}
}

// getErrorMessage returns a human-readable error message for a validation error
func (v *Validator) getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "drive115_id":
		return fmt.Sprintf("%s must be a valid 115drive ID (no spaces allowed)", fe.Field())
	case "urls":
		return fmt.Sprintf("%s must contain valid URLs (http/https/magnet)", fe.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// Global validator instance
var GlobalValidator = New()

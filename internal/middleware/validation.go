package middleware

import (
	"cloud-driver/internal/validation"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ValidationMiddleware creates a middleware that validates request bodies
func ValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Store the validator in the context for use in handlers
			c.Set("validator", validation.GlobalValidator)
			return next(c)
		}
	}
}

// ValidateRequest is a helper function to validate request bodies in handlers
func ValidateRequest(c echo.Context, req interface{}) error {
	// Bind the request body to the struct
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
	}

	// Get the validator from context
	validator, ok := c.Get("validator").(*validation.Validator)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Validator not available")
	}

	// Validate the struct
	if err := validator.ValidateStruct(req); err != nil {
		if validationErr, ok := err.(validation.ValidationErrors); ok {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
				"error":   "Validation failed",
				"details": validationErr.Errors,
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	return nil
}

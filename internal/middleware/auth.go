package middleware

import (
	"net/http"
	"strings"

	"cloud-driver/internal/services"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(authService *services.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization header")
			}

			// Check for Bearer token format
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization header format")
			}

			token := authHeader[len(bearerPrefix):]
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing session token")
			}

			// Validate session token
			user, err := authService.ValidateSession(c.Request().Context(), token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid session token")
			}

			// Set user in context for handlers to access
			c.Set("user", user)

			return next(c)
		}
	}
}

// OptionalAuthMiddleware creates an optional authentication middleware
// that doesn't fail if no token is provided, but sets user context if valid token is present
func OptionalAuthMiddleware(authService *services.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				// Check for Bearer token format
				const bearerPrefix = "Bearer "
				if strings.HasPrefix(authHeader, bearerPrefix) {
					token := authHeader[len(bearerPrefix):]
					if token != "" {
						// Validate session token
						user, err := authService.ValidateSession(c.Request().Context(), token)
						if err == nil {
							// Set user in context if valid
							c.Set("user", user)
						}
					}
				}
			}

			return next(c)
		}
	}
}

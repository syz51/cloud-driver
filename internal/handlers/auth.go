package handlers

import (
	"net/http"

	"cloud-driver/internal/models"
	"cloud-driver/internal/services"

	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	var req models.UserRegistrationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	user, err := h.authService.RegisterUser(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login handles user login
func (h *AuthHandler) Login(c echo.Context) error {
	var req models.UserLoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	authResponse, err := h.authService.LoginUser(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, authResponse)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	token := extractTokenFromHeader(c)
	if token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "No session token provided")
	}

	if err := h.authService.LogoutUser(c.Request().Context(), token); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to logout: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	return c.JSON(http.StatusOK, user)
}

// extractTokenFromHeader extracts the session token from Authorization header
func extractTokenFromHeader(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expect format: "Bearer <token>"
	const bearerPrefix = "Bearer "
	if len(authHeader) > len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
		return authHeader[len(bearerPrefix):]
	}

	return ""
}

// getUserFromContext extracts user from echo context (set by auth middleware)
func getUserFromContext(c echo.Context) *models.User {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return nil
	}
	return user
}

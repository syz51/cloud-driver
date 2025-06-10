package handlers

import (
	"net/http"
	"strconv"

	"cloud-driver/internal/models"
	"cloud-driver/internal/services"

	"github.com/labstack/echo/v4"
)

// CredentialsHandler handles Drive115 credentials related requests
type CredentialsHandler struct {
	credentialsService *services.CredentialsService
}

// NewCredentialsHandler creates a new credentials handler
func NewCredentialsHandler(credentialsService *services.CredentialsService) *CredentialsHandler {
	return &CredentialsHandler{
		credentialsService: credentialsService,
	}
}

// AddCredentials adds new Drive115 credentials for the authenticated user
func (h *CredentialsHandler) AddCredentials(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	var req models.Drive115CredentialsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	response, err := h.credentialsService.AddCredentials(c.Request().Context(), user.ID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add credentials: "+err.Error())
	}

	return c.JSON(http.StatusCreated, response)
}

// GetCredentials returns all credentials for the authenticated user
func (h *CredentialsHandler) GetCredentials(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	credentials, err := h.credentialsService.GetUserCredentials(c.Request().Context(), user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get credentials: "+err.Error())
	}

	return c.JSON(http.StatusOK, credentials)
}

// GetActiveCredentials returns only active credentials for the authenticated user
func (h *CredentialsHandler) GetActiveCredentials(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	credentials, err := h.credentialsService.GetActiveUserCredentials(c.Request().Context(), user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get active credentials: "+err.Error())
	}

	return c.JSON(http.StatusOK, credentials)
}

// GetCredentialsByID returns specific credentials by ID
func (h *CredentialsHandler) GetCredentialsByID(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	credentialsIDStr := c.Param("id")
	credentialsID, err := strconv.ParseInt(credentialsIDStr, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid credentials ID")
	}

	credentials, err := h.credentialsService.GetCredentialsByID(c.Request().Context(), user.ID, int32(credentialsID))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, credentials)
}

// UpdateCredentials updates existing credentials
func (h *CredentialsHandler) UpdateCredentials(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	credentialsIDStr := c.Param("id")
	credentialsID, err := strconv.ParseInt(credentialsIDStr, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid credentials ID")
	}

	var req models.Drive115CredentialsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	response, err := h.credentialsService.UpdateCredentials(c.Request().Context(), user.ID, int32(credentialsID), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update credentials: "+err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

// SetCredentialsActive sets the active status of credentials
func (h *CredentialsHandler) SetCredentialsActive(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	credentialsIDStr := c.Param("id")
	credentialsID, err := strconv.ParseInt(credentialsIDStr, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid credentials ID")
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if err := h.credentialsService.SetCredentialsActive(c.Request().Context(), user.ID, int32(credentialsID), req.IsActive); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update credentials status: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":   "Credentials status updated successfully",
		"is_active": req.IsActive,
	})
}

// DeleteCredentials deletes credentials
func (h *CredentialsHandler) DeleteCredentials(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	credentialsIDStr := c.Param("id")
	credentialsID, err := strconv.ParseInt(credentialsIDStr, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid credentials ID")
	}

	if err := h.credentialsService.DeleteCredentials(c.Request().Context(), user.ID, int32(credentialsID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete credentials: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Credentials deleted successfully",
	})
}

package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check returns the health status of the service
func (h *HealthHandler) Check(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "cloud-driver-server",
		"time":    time.Now().Format(time.RFC3339),
	})
}

package handlers

import (
	"net/http"
	"strconv"

	"cloud-driver/internal/models"
	"cloud-driver/internal/services"

	"github.com/labstack/echo/v4"
)

// Drive115Handler handles 115drive related requests
type Drive115Handler struct {
	service *services.Drive115Service
}

// NewDrive115Handler creates a new 115drive handler
func NewDrive115Handler(service *services.Drive115Service) *Drive115Handler {
	return &Drive115Handler{
		service: service,
	}
}

// GetUser returns the current user information
func (h *Drive115Handler) GetUser(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	userInfo, err := h.service.GetUser(c.Request().Context(), user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user info: "+err.Error())
	}

	return c.JSON(http.StatusOK, userInfo)
}

// ListOfflineTasks returns the list of offline download tasks
func (h *Drive115Handler) ListOfflineTasks(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	pageStr := c.QueryParam("page")
	page := int64(1)

	if pageStr != "" {
		if p, err := strconv.ParseInt(pageStr, 10, 64); err == nil {
			page = p
		}
	}

	tasks, err := h.service.ListOfflineTasks(c.Request().Context(), user.ID, page)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list offline tasks: "+err.Error())
	}

	return c.JSON(http.StatusOK, tasks)
}

// AddOfflineTask adds new offline download tasks
func (h *Drive115Handler) AddOfflineTask(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	var req models.OfflineDownloadRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if len(req.URLs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "URLs are required")
	}

	hashes, err := h.service.AddOfflineTaskURIs(c.Request().Context(), user.ID, req.URLs, req.SaveDirID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add offline task: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Offline download tasks added successfully",
		"hashes":  hashes,
		"count":   len(hashes),
	})
}

// DeleteOfflineTasks deletes offline tasks
func (h *Drive115Handler) DeleteOfflineTasks(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	var req models.DeleteTasksRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if len(req.Hashes) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Task hashes are required")
	}

	err := h.service.DeleteOfflineTasks(c.Request().Context(), user.ID, req.Hashes, req.DeleteFiles)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete offline tasks: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Offline tasks deleted successfully",
		"deleted_count": len(req.Hashes),
		"files_deleted": req.DeleteFiles,
	})
}

// ClearOfflineTasks clears offline tasks
func (h *Drive115Handler) ClearOfflineTasks(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	var req models.ClearTasksRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	err := h.service.ClearOfflineTasks(c.Request().Context(), user.ID, req.ClearFlag)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to clear offline tasks: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Offline tasks cleared successfully",
	})
}

// ListFiles lists files and directories
func (h *Drive115Handler) ListFiles(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	dirIDStr := c.QueryParam("dir_id")
	dirID := int64(0) // Root directory

	if dirIDStr != "" {
		if id, err := strconv.ParseInt(dirIDStr, 10, 64); err == nil {
			dirID = id
		}
	}

	files, err := h.service.ListFiles(c.Request().Context(), user.ID, dirID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list files: "+err.Error())
	}

	return c.JSON(http.StatusOK, files)
}

// GetFileInfo returns information about a specific file
func (h *Drive115Handler) GetFileInfo(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file ID")
	}

	fileInfo, err := h.service.GetFileInfo(c.Request().Context(), user.ID, fileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get file info: "+err.Error())
	}

	return c.JSON(http.StatusOK, fileInfo)
}

// DownloadFile returns download information for a file
func (h *Drive115Handler) DownloadFile(c echo.Context) error {
	user := getUserFromContext(c)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file ID")
	}

	downloadInfo, err := h.service.GetDownloadInfo(c.Request().Context(), user.ID, fileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get download info: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Download info retrieved successfully",
		"download_info": downloadInfo,
	})
}

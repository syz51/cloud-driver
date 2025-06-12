package handlers

import (
	"net/http"
	"strconv"

	"cloud-driver/internal/middleware"
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
	var req models.GetUserRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	userInfo, err := h.service.GetUser(c.Request().Context(), req.Credentials)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user info: "+err.Error())
	}

	return c.JSON(http.StatusOK, userInfo)
}

// ListOfflineTasks returns the list of offline download tasks
func (h *Drive115Handler) ListOfflineTasks(c echo.Context) error {
	var req models.TaskListRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Handle page parameter from query string if not in body
	if req.Page == 0 {
		pageStr := c.QueryParam("page")
		if pageStr != "" {
			if p, err := strconv.ParseInt(pageStr, 10, 64); err == nil {
				req.Page = p
			}
		}
		if req.Page == 0 {
			req.Page = 1
		}
	}

	tasks, err := h.service.ListOfflineTasks(c.Request().Context(), req.Credentials, req.Page)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list offline tasks: "+err.Error())
	}

	return c.JSON(http.StatusOK, tasks)
}

// AddOfflineTask adds new offline download tasks
func (h *Drive115Handler) AddOfflineTask(c echo.Context) error {
	var req models.OfflineDownloadRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	hashes, err := h.service.AddOfflineTaskURIs(c.Request().Context(), req.Credentials, req.URLs, req.SaveDirID)
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
	var req models.DeleteTasksRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	err := h.service.DeleteOfflineTasks(c.Request().Context(), req.Credentials, req.Hashes, req.DeleteFiles)
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
	var req models.ClearTasksRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	err := h.service.ClearOfflineTasks(c.Request().Context(), req.Credentials, req.ClearFlag)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to clear offline tasks: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Offline tasks cleared successfully",
	})
}

// ListFiles lists files and directories
func (h *Drive115Handler) ListFiles(c echo.Context) error {
	var req models.ListFilesRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Handle dir_id parameter from query string if not in body
	if req.DirID == 0 {
		dirIDStr := c.QueryParam("dir_id")
		if dirIDStr != "" {
			if id, err := strconv.ParseInt(dirIDStr, 10, 64); err == nil {
				req.DirID = id
			}
		}
	}

	files, err := h.service.ListFiles(c.Request().Context(), req.Credentials, req.DirID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list files: "+err.Error())
	}

	return c.JSON(http.StatusOK, files)
}

// GetFileInfo returns information about a specific file
func (h *Drive115Handler) GetFileInfo(c echo.Context) error {
	var req models.FileInfoRequest

	// Get file ID from URL path parameter
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file ID")
	}
	req.FileID = fileID

	// Get credentials from request body and validate
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	fileInfo, err := h.service.GetFileInfo(c.Request().Context(), req.Credentials, req.FileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get file info: "+err.Error())
	}

	return c.JSON(http.StatusOK, fileInfo)
}

// DownloadFile returns download information for a file
func (h *Drive115Handler) DownloadFile(c echo.Context) error {
	var req models.DownloadRequest

	// Get file ID from URL path parameter
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file ID")
	}
	req.FileID = fileID

	// Get credentials from request body and validate
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	downloadInfo, err := h.service.GetDownloadInfo(c.Request().Context(), req.Credentials, req.FileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get download info: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Download info retrieved successfully",
		"download_info": downloadInfo,
	})
}

// QRCodeStart starts a new QR code login session
func (h *Drive115Handler) QRCodeStart(c echo.Context) error {
	var req models.QRCodeStartRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	response, err := h.service.QRCodeStart(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start QR code session: "+err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

// QRCodeImage generates and returns QR code image data
func (h *Drive115Handler) QRCodeImage(c echo.Context) error {
	var req models.QRCodeImageRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	imageData, err := h.service.QRCodeGetImage(c.Request().Context(), req.UID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate QR code image: "+err.Error())
	}

	// Set proper headers for PNG image
	c.Response().Header().Set("Content-Type", "image/png")
	c.Response().Header().Set("Content-Length", strconv.Itoa(len(imageData)))
	c.Response().Header().Set("Cache-Control", "no-cache")

	return c.Blob(http.StatusOK, "image/png", imageData)
}

// QRCodeStatus checks the current status of a QR code scan
func (h *Drive115Handler) QRCodeStatus(c echo.Context) error {
	var req models.QRCodeStatusRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	response, err := h.service.QRCodeCheckStatus(c.Request().Context(), req.UID, req.Sign, req.Time)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check QR code status: "+err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

// QRCodeLogin completes the QR code login process and returns credentials
func (h *Drive115Handler) QRCodeLogin(c echo.Context) error {
	var req models.QRCodeLoginRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	response, err := h.service.QRCodeLogin(c.Request().Context(), req.UID, req.Sign, req.Time, req.App)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to complete QR code login: "+err.Error())
	}

	// Return appropriate status code based on success
	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	return c.JSON(statusCode, response)
}

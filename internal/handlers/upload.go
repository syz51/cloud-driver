package handlers

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"cloud-driver/internal/middleware"
	"cloud-driver/internal/models"
	"cloud-driver/internal/services"

	"github.com/labstack/echo/v4"
)

const uploadSessionLifetime = 30 * 24 * time.Hour

func (h *Drive115Handler) InitUpload(c echo.Context) error {
	var req models.UploadInitRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}
	fileName, err := validUploadFileName(req.FileName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if (req.SignKey == "") != (req.SignValue == "") {
		return echo.NewHTTPError(http.StatusBadRequest, "sign_key and sign_value must be provided together")
	}
	if req.DirID == "" {
		req.DirID = "0"
	}
	req.FileName = fileName
	req.SHA1 = strings.ToUpper(req.SHA1)
	req.PreSHA1 = strings.ToUpper(req.PreSHA1)
	req.SignValue = strings.ToUpper(req.SignValue)
	expiresAt := time.Now().Add(uploadSessionLifetime).Unix()
	result, err := h.uploads.InitUpload(c.Request().Context(), req, expiresAt)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "Failed to initialize upload: "+err.Error())
	}

	switch result.State {
	case "instant":
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "instant", "message": "File uploaded successfully", "name": req.FileName, "dir_id": req.DirID, "size": req.FileSize,
		})
	case "sign_check":
		if result.SignKey == "" || result.SignCheck == "" {
			return echo.NewHTTPError(http.StatusBadGateway, "115 returned an invalid sign check")
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "sign_check", "sign_key": result.SignKey, "sign_check": result.SignCheck,
		})
	case "upload":
		if result.Session == nil {
			return echo.NewHTTPError(http.StatusBadGateway, "115 returned an invalid upload session")
		}
		token, err := h.uploadCodec.encode(*result.Session)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create upload session")
		}
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"status": "upload", "session_token": token, "part_size": result.Session.PartSize, "expires_at": result.Session.ExpiresAt,
		})
	default:
		return echo.NewHTTPError(http.StatusBadGateway, "115 returned an invalid upload state")
	}
}

func (h *Drive115Handler) UploadStatus(c echo.Context) error {
	session, err := h.sessionFromRequest(c, false)
	if err != nil {
		return err
	}
	progress, serviceErr := h.uploads.UploadStatus(c.Request().Context(), session)
	if serviceErr != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "Failed to read upload status: "+serviceErr.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"next_part": progress.NextPart, "uploaded_bytes": progress.UploadedBytes, "complete": progress.Complete,
		"part_size": session.PartSize, "file_size": session.FileSize, "expires_at": session.ExpiresAt,
	})
}

func (h *Drive115Handler) UploadPart(c echo.Context) error {
	session, err := h.sessionFromRequest(c, false)
	if err != nil {
		return err
	}
	partNumber, err := strconv.Atoi(c.Request().Header.Get("X-Part-Number"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Part-Number must be a positive integer")
	}
	expected, err := uploadPartSize(session, partNumber)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if c.Request().ContentLength != expected {
		return echo.NewHTTPError(http.StatusBadRequest, "part Content-Length does not match expected size")
	}
	if err := h.uploads.UploadPart(c.Request().Context(), session, partNumber, c.Request().Body); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "Failed to upload part: "+err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Drive115Handler) CompleteUpload(c echo.Context) error {
	session, err := h.sessionFromRequest(c, false)
	if err != nil {
		return err
	}
	if err := h.uploads.CompleteUpload(c.Request().Context(), session); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "Failed to complete upload: "+err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "File uploaded successfully", "dir_id": session.DirID, "name": session.FileName, "size": session.FileSize,
	})
}

func (h *Drive115Handler) AbortUpload(c echo.Context) error {
	session, err := h.sessionFromRequest(c, true)
	if err != nil {
		return err
	}
	if err := h.uploads.AbortUpload(c.Request().Context(), session); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "Failed to abort upload: "+err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Drive115Handler) sessionFromRequest(c echo.Context, allowExpired bool) (services.UploadSession, error) {
	authorization := c.Request().Header.Get(echo.HeaderAuthorization)
	if !strings.HasPrefix(authorization, "Bearer ") {
		return services.UploadSession{}, echo.NewHTTPError(http.StatusUnauthorized, "Missing upload session")
	}
	session, err := h.uploadCodec.decode(strings.TrimPrefix(authorization, "Bearer "), allowExpired)
	if err != nil {
		return services.UploadSession{}, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	return session, nil
}

func validUploadFileName(raw string) (string, error) {
	name := path.Base(strings.ReplaceAll(raw, "\\", "/"))
	invalidControl := strings.IndexFunc(name, func(r rune) bool { return r < 0x20 || r == 0x7f }) >= 0
	if name == "." || name == ".." || name == "/" || len(name) == 0 || len(name) > 255 || invalidControl {
		return "", fmt.Errorf("file name must be valid and between 1 and 255 bytes")
	}
	return name, nil
}

func uploadPartSize(session services.UploadSession, partNumber int) (int64, error) {
	if partNumber < 1 {
		return 0, fmt.Errorf("part number must be positive")
	}
	offset := int64(partNumber-1) * session.PartSize
	if offset >= session.FileSize {
		return 0, fmt.Errorf("part number exceeds file size")
	}
	remaining := session.FileSize - offset
	if remaining < session.PartSize {
		return remaining, nil
	}
	return session.PartSize, nil
}

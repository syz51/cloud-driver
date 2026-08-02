package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"cloud-driver/internal/middleware"
	"cloud-driver/internal/models"
	"cloud-driver/internal/services"

	"github.com/labstack/echo/v4"
)

func TestUploadFileIntegration(t *testing.T) {
	if os.Getenv("CLOUD_DRIVER_INTEGRATION") != "1" {
		t.Skip("set CLOUD_DRIVER_INTEGRATION=1 to run integration tests")
	}

	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	for name, value := range map[string]string{
		"uid": "uid", "cid": "cid", "seid": "seid", "kid": "kid", "dir_id": "42",
	} {
		if err := form.WriteField(name, value); err != nil {
			t.Fatal(err)
		}
	}
	file, err := form.CreateFormFile("file", `..\video.mp4`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("video")); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}

	handler := NewDrive115Handler(services.NewDrive115Service())
	handler.uploadFile = func(_ context.Context, credentials models.Drive115Credentials, dirID, fileName string, fileSize int64, file io.ReadSeeker) error {
		content, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		if credentials.UID != "uid" || dirID != "42" || fileName != "video.mp4" || fileSize != 5 || string(content) != "video" {
			t.Fatalf("unexpected upload: credentials=%+v dir=%q name=%q size=%d content=%q", credentials, dirID, fileName, fileSize, content)
		}
		return nil
	}

	e := echo.New()
	e.Use(middleware.ValidationMiddleware())
	e.POST("/api/v1/115/files/upload", handler.UploadFile)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/115/files/upload", &body)
	req.Header.Set(echo.HeaderContentType, form.FormDataContentType())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var response struct {
		Name  string `json:"name"`
		Size  int64  `json:"size"`
		DirID string `json:"dir_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Name != "video.mp4" || response.Size != 5 || response.DirID != "42" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

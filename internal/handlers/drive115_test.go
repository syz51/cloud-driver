package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cloud-driver/internal/middleware"
	"cloud-driver/internal/models"
	"cloud-driver/internal/services"

	"github.com/labstack/echo/v4"
)

type fakeUploadService struct {
	initResult *services.UploadInitResult
	partCalled bool
}

func (f *fakeUploadService) InitUpload(context.Context, models.UploadInitRequest, int64) (*services.UploadInitResult, error) {
	return f.initResult, nil
}

func (f *fakeUploadService) UploadStatus(context.Context, services.UploadSession) (*services.UploadProgress, error) {
	return &services.UploadProgress{NextPart: 1}, nil
}

func (f *fakeUploadService) UploadPart(_ context.Context, _ services.UploadSession, _ int, source io.Reader) error {
	f.partCalled = true
	_, err := io.Copy(io.Discard, source)
	return err
}

func (f *fakeUploadService) CompleteUpload(context.Context, services.UploadSession) error { return nil }
func (f *fakeUploadService) AbortUpload(context.Context, services.UploadSession) error    { return nil }

func TestUploadProtocolIntegration(t *testing.T) {
	if os.Getenv("CLOUD_DRIVER_INTEGRATION") != "1" {
		t.Skip("set CLOUD_DRIVER_INTEGRATION=1 to run integration tests")
	}

	handler, err := NewDrive115Handler(services.NewDrive115Service(), "test-upload-session-secret-at-least-32-characters")
	if err != nil {
		t.Fatal(err)
	}
	handler.uploads = &fakeUploadService{initResult: &services.UploadInitResult{State: "instant"}}
	e := echo.New()
	e.Use(middleware.ValidationMiddleware())
	e.POST("/api/v1/115/uploads/init", handler.InitUpload)
	body := `{"credentials":{"uid":"uid","cid":"cid","seid":"seid","kid":"kid"},"dir_id":"42","file_name":"video.mp4","file_size":5,"sha1":"0123456789abcdef0123456789abcdef01234567","pre_sha1":"0123456789abcdef0123456789abcdef01234567"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/115/uploads/init", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"instant"`) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

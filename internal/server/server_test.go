package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cloud-driver/internal/config"
)

func TestUploadRouteIntegration(t *testing.T) {
	if os.Getenv("CLOUD_DRIVER_INTEGRATION") != "1" {
		t.Skip("set CLOUD_DRIVER_INTEGRATION=1 to run integration tests")
	}

	readLog, writeLog, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	originalStdout := os.Stdout
	os.Stdout = writeLog
	defer func() {
		os.Stdout = originalStdout
		readLog.Close()
		writeLog.Close()
	}()

	server, err := New(&config.Config{UploadBodyLimit: "4B"})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/115/files/upload", strings.NewReader("12345"))
	rec := httptest.NewRecorder()
	server.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	if err := writeLog.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = originalStdout
	logData, err := io.ReadAll(readLog)
	if err != nil {
		t.Fatal(err)
	}
	var logEntry map[string]interface{}
	if err := json.Unmarshal(logData, &logEntry); err != nil {
		t.Fatalf("invalid request log %q: %v", logData, err)
	}
	if len(logEntry) != 13 {
		t.Fatalf("unexpected request log fields: %s", logData)
	}
	for _, key := range []string{"time", "id", "remote_ip", "host", "method", "uri", "user_agent", "status", "error", "latency", "latency_human", "bytes_in", "bytes_out"} {
		if _, ok := logEntry[key]; !ok {
			t.Fatalf("request log missing %q: %s", key, logData)
		}
	}
}

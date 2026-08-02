package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-driver/internal/config"
)

func TestServerSupportsH2C(t *testing.T) {
	server, err := New(&config.Config{UploadSessionSecret: "test-upload-session-secret-at-least-32-characters"})
	if err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server.echo.Listener = listener

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Start() }()

	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	client := &http.Client{Transport: &http.Transport{Protocols: protocols}}
	response, err := client.Get("http://" + listener.Addr().String() + "/health")
	if err != nil {
		t.Fatal(err)
	}

	if response.ProtoMajor != 2 {
		t.Fatalf("protocol = %s, want HTTP/2", response.Proto)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	response.Body.Close()
	client.CloseIdleConnections()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		t.Fatal(err)
	}
}

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

	server, err := New(&config.Config{UploadPartBodyLimit: "4B", UploadSessionSecret: "test-upload-session-secret-at-least-32-characters"})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/115/uploads/part", strings.NewReader("12345"))
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

func TestUploadCORSIntegration(t *testing.T) {
	if os.Getenv("CLOUD_DRIVER_INTEGRATION") != "1" {
		t.Skip("set CLOUD_DRIVER_INTEGRATION=1 to run integration tests")
	}
	server, err := New(&config.Config{
		UploadSessionSecret: "test-upload-session-secret-at-least-32-characters",
		AllowedOrigins:      []string{"https://drive.example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		origin string
		allow  bool
	}{
		{origin: "https://drive.example.com", allow: true},
		{origin: "https://evil.example.com", allow: false},
	} {
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/115/uploads/part", nil)
		req.Header.Set("Origin", test.origin)
		req.Header.Set("Access-Control-Request-Method", http.MethodPut)
		req.Header.Set("Access-Control-Request-Headers", "authorization,x-part-number,content-type")
		rec := httptest.NewRecorder()
		server.echo.ServeHTTP(rec, req)
		got := rec.Header().Get("Access-Control-Allow-Origin")
		if test.allow && got != test.origin {
			t.Fatalf("origin %q: allow header = %q", test.origin, got)
		}
		if !test.allow && got != "" {
			t.Fatalf("origin %q unexpectedly allowed", test.origin)
		}
	}
}

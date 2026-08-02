package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud-driver/internal/config"
	"cloud-driver/internal/handlers"
	"cloud-driver/internal/middleware"
	"cloud-driver/internal/services"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	echo   *echo.Echo
}

// New creates a new server instance
func New(cfg *config.Config) (*Server, error) {
	// Initialize 115drive service (no database needed)
	drive115Service := services.NewDrive115Service()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	drive115Handler := handlers.NewDrive115Handler(drive115Service)

	// Setup Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogLatency:      true,
		LogRemoteIP:     true,
		LogHost:         true,
		LogMethod:       true,
		LogURI:          true,
		LogRequestID:    true,
		LogUserAgent:    true,
		LogStatus:       true,
		LogError:        true,
		LogResponseSize: true,
		HandleError:     true,
		LogValuesFunc: func(c echo.Context, values echomiddleware.RequestLoggerValues) error {
			errorMessage := ""
			if values.Error != nil {
				errorMessage = values.Error.Error()
			}
			bytesIn := c.Request().ContentLength
			if bytesIn < 0 {
				bytesIn = 0
			}

			entry, err := json.Marshal(struct {
				Time         string `json:"time"`
				ID           string `json:"id"`
				RemoteIP     string `json:"remote_ip"`
				Host         string `json:"host"`
				Method       string `json:"method"`
				URI          string `json:"uri"`
				UserAgent    string `json:"user_agent"`
				Status       int    `json:"status"`
				Error        string `json:"error"`
				Latency      int64  `json:"latency"`
				LatencyHuman string `json:"latency_human"`
				BytesIn      int64  `json:"bytes_in"`
				BytesOut     int64  `json:"bytes_out"`
			}{
				Time:         time.Now().Format(time.RFC3339Nano),
				ID:           values.RequestID,
				RemoteIP:     values.RemoteIP,
				Host:         values.Host,
				Method:       values.Method,
				URI:          values.URI,
				UserAgent:    values.UserAgent,
				Status:       values.Status,
				Error:        errorMessage,
				Latency:      int64(values.Latency),
				LatencyHuman: values.Latency.String(),
				BytesIn:      bytesIn,
				BytesOut:     values.ResponseSize,
			})
			if err != nil {
				return err
			}
			entry = append(entry, '\n')
			_, err = os.Stdout.Write(entry)
			return err
		},
	}))
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	e.Use(middleware.ValidationMiddleware())

	// Setup routes
	uploadBodyLimit := cfg.UploadBodyLimit
	if uploadBodyLimit == "" {
		uploadBodyLimit = "20G"
	}
	setupRoutes(e, healthHandler, drive115Handler, uploadBodyLimit)

	return &Server{
		config: cfg,
		echo:   e,
	}, nil
}

// setupRoutes configures all the application routes
func setupRoutes(e *echo.Echo, healthHandler *handlers.HealthHandler, drive115Handler *handlers.Drive115Handler, uploadBodyLimit string) {
	// Health check
	e.GET("/health", healthHandler.Check)

	// API routes
	api := e.Group("/api/v1")

	// 115drive routes
	drive115 := api.Group("/115")
	{
		drive115.POST("/user", drive115Handler.GetUser)
		drive115.POST("/tasks", drive115Handler.ListOfflineTasks)
		drive115.POST("/tasks/add", drive115Handler.AddOfflineTask)
		drive115.POST("/tasks/delete", drive115Handler.DeleteOfflineTasks)
		drive115.POST("/tasks/clear", drive115Handler.ClearOfflineTasks)
		drive115.POST("/files", drive115Handler.ListFiles)
		drive115.POST("/files/upload", drive115Handler.UploadFile, echomiddleware.BodyLimit(uploadBodyLimit))
		drive115.POST("/files/video-check", drive115Handler.CheckFolderVideos)
		drive115.POST("/files/:id", drive115Handler.GetFileInfo)
		drive115.POST("/files/:id/download", drive115Handler.DownloadFile)

		// QR Code login routes
		drive115.POST("/qrcode/start", drive115Handler.QRCodeStart)
		drive115.POST("/qrcode/image", drive115Handler.QRCodeImage)
		drive115.POST("/qrcode/status", drive115Handler.QRCodeStatus)
		drive115.POST("/qrcode/login", drive115Handler.QRCodeLogin)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	address := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	s.echo.Server.Addr = address
	s.echo.Server.Protocols = protocols
	return s.echo.StartServer(s.echo.Server)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

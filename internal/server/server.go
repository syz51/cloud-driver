package server

import (
	"context"
	"fmt"

	"cloud-driver/internal/config"
	"cloud-driver/internal/database"
	"cloud-driver/internal/handlers"
	"cloud-driver/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	echo   *echo.Echo
	db     *database.Database
}

// New creates a new server instance
func New(cfg *config.Config) (*Server, error) {
	// Initialize database connection
	db, err := database.New(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize 115drive service with database pool
	drive115Service := services.NewDrive115Service(db.GetPool())

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	drive115Handler := handlers.NewDrive115Handler(drive115Service)

	// Setup Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Setup routes
	setupRoutes(e, healthHandler, drive115Handler)

	return &Server{
		config: cfg,
		echo:   e,
		db:     db,
	}, nil
}

// setupRoutes configures all the application routes
func setupRoutes(e *echo.Echo, healthHandler *handlers.HealthHandler, drive115Handler *handlers.Drive115Handler) {
	// Health check
	e.GET("/health", healthHandler.Check)

	// API routes
	api := e.Group("/api/v1")

	// 115drive routes
	drive115 := api.Group("/115")
	{
		drive115.GET("/user", drive115Handler.GetUser)
		drive115.GET("/tasks", drive115Handler.ListOfflineTasks)
		drive115.POST("/tasks/add", drive115Handler.AddOfflineTask)
		drive115.DELETE("/tasks", drive115Handler.DeleteOfflineTasks)
		drive115.DELETE("/tasks/clear", drive115Handler.ClearOfflineTasks)
		drive115.GET("/files", drive115Handler.ListFiles)
		drive115.GET("/files/:id", drive115Handler.GetFileInfo)
		drive115.POST("/files/:id/download", drive115Handler.DownloadFile)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	address := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	return s.echo.Start(address)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Close database connection
	if s.db != nil {
		s.db.Close()
	}
	return s.echo.Shutdown(ctx)
}

package api

import (
	"context"
	"go-template/internal/api/router"
	"go-template/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ServerConfig holds server related configuration
type Config struct {
	Addr            string `mapstructure:"addr"`
	ReadTimeout     int    `mapstructure:"read_timeout"`
	WriteTimeout    int    `mapstructure:"write_timeout"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
	Debug           bool
}

// Server represents the HTTP server
type Server struct {
	server *http.Server
	router *gin.Engine
	config Config
}

// NewServer creates and configures a new server instance
func NewServer(cfg Config) *Server {
	// Set Gin mode based on debug flag
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin engine
	r := gin.Default()

	// Add middleware
	if cfg.Debug {
		r.Use(logger.GinMiddleware())
	}
	r.Use(gin.Recovery())

	// Setup routes
	router.SetupRoutes(r)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}

	return &Server{
		server: srv,
		router: r,
		config: cfg,
	}
}

// Start begins listening for requests
func (s *Server) Start() error {
	logger.Infof("Server listening on %s", s.config.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down server...")
	return s.server.Shutdown(ctx)
}

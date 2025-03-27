// @title           Go Template API
// @version         1.0
// @description     A RESTful API for Go Template

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @BasePath  /api/v1
package api

import (
	"context"
	"go-template/internal/api/middleware"
	"go-template/internal/api/router"
	"go-template/internal/config"
	"go-template/internal/database"
	"go-template/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	router *gin.Engine
	config *config.Config
	db     *database.Client
}

// NewServer creates and configures a new server instance
func NewServer(cfg *config.Config, db *database.Client) *Server {
	// Set Gin mode based on debug flag
	if cfg.Server.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin engine
	r := gin.Default()

	// Add middleware
	if cfg.Server.Debug {
		r.Use(middleware.RequestLog())
	}
	r.Use(gin.Recovery())

	// Setup routes
	router.SetupRoutes(r, db, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	return &Server{
		server: srv,
		router: r,
		config: cfg,
	}
}

// Start begins listening for requests
func (s *Server) ListenAndServe() error {
	logger.Infof("Server listening on %s", s.config.Server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down server...")
	return s.server.Shutdown(ctx)
}

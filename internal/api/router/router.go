package router

import (
	"go-template/internal/api/handler"
	"go-template/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the routes for the server
func SetupRoutes(r *gin.Engine) {
	r.Use(middleware.RequestID())

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/", handler.Welcome)

	}
}

package router

import (
	"go-template/internal/api/handler"
	"go-template/internal/api/middleware"
	"go-template/internal/database"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the routes for the server
func SetupRoutes(r *gin.Engine, db *database.Client) {
	r.Use(middleware.RequestID())

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/", handler.Welcome)
		// User routes
		userHandler := handler.NewUserHandler(db)
		users := v1.Group("/users")
		{
			users.GET("/", userHandler.List)
			users.GET("/:id", userHandler.Get)
			users.POST("/", userHandler.Create)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		// Role routes
		roleHandler := handler.NewRoleHandler(db)
		roles := v1.Group("/roles")
		{
			roles.GET("/", roleHandler.List)
			roles.GET("/:id", roleHandler.Get)
			roles.POST("/", roleHandler.Create)
			roles.PUT("/:id", roleHandler.Update)
			roles.DELETE("/:id", roleHandler.Delete)
			roles.GET("/:id/users", roleHandler.GetUsers) // Get users with this role
		}

	}
}

package router

import (
	"go-template/internal/api/handler"
	"go-template/internal/api/middleware"
	"go-template/internal/config"
	"go-template/internal/database"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-template/docs"
)

// SetupRoutes configures all the routes for the server
func SetupRoutes(r *gin.Engine, db *database.Client, cfg *config.Config) {
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	// Swagger routes
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/", handler.Welcome)
		// auth routes
		authHandler := handler.NewAuthHandler(db, cfg.JWT)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)

			authRequired := auth.Group("")
			authRequired.Use(middleware.JWTAuthMiddleware(cfg.JWT))
			{
				authRequired.GET("/me", authHandler.GetUserInfo)
			}
		}

		protected := v1.Group("")
		protected.Use(middleware.JWTAuthMiddleware(cfg.JWT))
		// User routes
		userHandler := handler.NewUserHandler(db)
		users := protected.Group("/users")
		{
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.Get)
			adminOnly := users.Group("")
			adminOnly.Use(middleware.RequireRole("admin"))
			{
				adminOnly.POST("", userHandler.Create)
				adminOnly.PUT("/:id", userHandler.Update)
				adminOnly.DELETE("/:id", userHandler.Delete)
			}
		}

		// Role routes
		roleHandler := handler.NewRoleHandler(db)
		roles := protected.Group("/roles")
		roles.Use(middleware.RequireRole("admin"))
		{
			roles.GET("", roleHandler.List)
			roles.GET("/:id", roleHandler.Get)
			roles.POST("", roleHandler.Create)
			roles.PUT("/:id", roleHandler.Update)
			roles.DELETE("/:id", roleHandler.Delete)
			roles.GET("/:id/users", roleHandler.GetUsers) // Get users with this role
		}

		// Raw Query routes
		setupRawQueryRoutes(v1, db)
	}
}

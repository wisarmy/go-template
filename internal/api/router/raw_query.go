package router

import (
	"go-template/internal/api/handler"
	"go-template/internal/database"

	"github.com/gin-gonic/gin"
)

// setupRawQueryRoutes
func setupRawQueryRoutes(rg *gin.RouterGroup, db *database.Client) {
	// Raw Query routes
	rawQueryHandler := handler.NewRawQueryHandler(db)
	rawQueries := rg.Group("/raw")
	{
		rawQueries.GET("/user-stats", rawQueryHandler.GetUserStats)
		rawQueries.GET("/role-user-counts", rawQueryHandler.GetRoleUserCounts)
		rawQueries.POST("/execute-update", rawQueryHandler.ExecuteUpdateExample)
		rawQueries.GET("/data-structures", rawQueryHandler.RawDataStructsExample)
	}
}

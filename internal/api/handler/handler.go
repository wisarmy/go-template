package handler

import (
	"go-template/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Welcome returns a welcome message
func Welcome(c *gin.Context) {
	logger.Info(c.GetString("RequestID"))
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to the API service",
		"version": "1.0.0",
	})
}

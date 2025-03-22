package cmd

import (
	"context"
	"fmt"
	"go-template/internal/config"
	"go-template/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the daemon service",
	Long:  `Start the daemon service`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := startDaemon(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}

func startDaemon() error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("Failed to load config: %w", err)
	}
	// Initialize logger
	if err := logger.Init(&cfg.Log); err != nil {
		return fmt.Errorf("Failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin engine
	r := gin.Default()

	// Configure routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to the service",
		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Listening on %s", cfg.Server.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Listen failed: %s", err)
		}
	}()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Wait for signal
	sig := <-sigChan
	logger.Infof("received signal: %v, shutting down service...", sig)

	// Create a context with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	// Gracefully shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("server shutdown error: %v", err)
	}
	return nil
}

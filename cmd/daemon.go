package cmd

import (
	"context"
	"fmt"
	"go-template/internal/api"
	"go-template/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	cfg.Server.Debug = cfg.Log.Level == "debug"
	server := api.NewServer(cfg.Server)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Listen failed: %s", err)
		}
	}()

	// Set up signal handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// Wait for interrupt signal
	sig := <-quit
	logger.Infof("received signal: %v, shutting down service...", sig)

	// Create a context with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	// Gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("server shutdown error: %v", err)
	}
	return nil
}

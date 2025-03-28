package cmd

import (
	"fmt"
	"go-template/internal/config"
	"go-template/internal/database"
	"go-template/pkg/logger"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	cfg      *config.Config
	dbClient *database.Client
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "go-template",
	Short:        "A go template application",
	Long:         `A go template application containing a set of utilities`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for help command
		if cmd.Name() == "help" {
			return nil
		}
		// Load config
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Initialize logger
		if err := logger.Init(&cfg.Log); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		logger.Infof("Logger initialized with level: %s", logger.Level())

		// Initialize database connection
		if cmd.Name() == "daemon" {
			dbClient, err = database.New(&cfg.Database)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Close database connection if it exists
		if dbClient != nil {
			if err := dbClient.Close(); err != nil {
				logger.Errorf("Error closing database connection: %v", err)
			}
		}
		logger.Sync()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is configs/app.toml)")

}

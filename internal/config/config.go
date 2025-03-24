package config

import (
	"fmt"
	"go-template/internal/api"
	"go-template/internal/database"
	"go-template/pkg/logger"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   api.Config      `mapstructure:"server"`
	Log      logger.Config   `mapstructure:"log"`
	Database database.Config `mapstructure:"database"`
}

// Load loads configuration from file
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	if cfgFile != "" {
		// Use config file from the flag
		v.SetConfigFile(cfgFile)
	} else {
		// Find config directory
		configDir := "./configs"
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return nil, fmt.Errorf("creating config directory: %w", err)
			}
		}

		// Search for config in configs directory with name "app.toml"
		v.AddConfigPath(configDir)
		v.SetConfigName("app")
		v.SetConfigType("toml")

		// Create default config if it doesn't exist
		configPath := filepath.Join(configDir, "app.toml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Println("Config file not found, creating default config:", configPath)
			// Copy example file if it exists
			examplePath := filepath.Join(configDir, "app.toml.example")
			if _, err := os.Stat(examplePath); err == nil {
				if exampleData, err := os.ReadFile(examplePath); err == nil {
					if err := os.WriteFile(configPath, exampleData, 0644); err != nil {
						fmt.Printf("Failed to create default config: %v\n", err)
					} else {
						fmt.Println("Default config created successfully from example file")
					}
				}
			}
		}
	}

	// Set defaults
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("server.read_timeout", 10)
	v.SetDefault("server.write_timeout", 10)
	v.SetDefault("server.shutdown_timeout", 5)

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.output", "console")
	v.SetDefault("log.file", "logs/app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 3)
	v.SetDefault("log.max_age", 28)
	v.SetDefault("log.compress", true)

	// Read config
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	fmt.Println("Using config file:", v.ConfigFileUsed())

	// Unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &config, nil
}

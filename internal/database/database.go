package database

import (
	"context"
	"fmt"
	"go-template/ent"
	"go-template/ent/migrate"
	"go-template/pkg/logger"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	Driver string `mapstructure:"driver"` // Database driver type
	DSN    string `mapstructure:"dsn"`    // Database connection string
	Debug  bool   `mapstructure:"debug"`  // Enable SQL statement logging
}

// Client represents the database client
type Client struct {
	Ent *ent.Client
}

// New creates a new database client
func New(cfg *Config) (*Client, error) {
	RegisterPgxDriver()
	logger.Info("Connecting to database...")

	// options for the client
	opts := []ent.Option{
		// ent.Debug(), // Enable SQL statement logging
	}

	// Connect to the database using the specified driver and DSN
	client, err := ent.Open(cfg.Driver, cfg.DSN, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to database: %w", err)
	}
	ent.NewClient(opts...)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Schema.Create(ctx,
		migrate.WithForeignKeys(false),
		migrate.WithDropColumn(false),
	); err != nil {
		return nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	logger.Info("Database connection established")
	return &Client{Ent: client}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	if c.Ent != nil {
		logger.Info("Closing database connection")
		return c.Ent.Close()
	}
	return nil
}

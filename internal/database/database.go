package database

import (
	"context"
	"database/sql"
	"fmt"
	"go-template/ent"
	"go-template/ent/migrate"
	"go-template/pkg/logger"
	"os"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
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
	db  *sql.DB
}

// New creates a new database client
func New(cfg *Config) (*Client, error) {
	RegisterPgxDriver()
	logger.Info("Connecting to database...")

	// Connect to the database using the specified driver and DSN
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to database: %w", err)
	}

	// options for the client
	var opts []ent.Option
	if cfg.Debug {
		opts = append(opts, ent.Debug())
		opts = append(opts, ent.Log(func(a ...any) {
			message := fmt.Sprint(a...)
			// filter initialization queries
			if strings.Contains(message, "INFORMATION_SCHEMA") ||
				strings.Contains(message, "pg_catalog") ||
				strings.Contains(message, "server_version_num") ||
				strings.Contains(message, "pg_namespace") ||
				strings.Contains(message, "pg_constraint") ||
				strings.Contains(message, "pg_class") {
				return
			}
			logger.Info(message)
		}))
	}
	// Create an ent.Driver from `db`
	drv := entsql.OpenDB(dialect.Postgres, db)
	opts = append(opts, ent.Driver(drv))

	client := ent.NewClient(opts...)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	unsafeMigrate := strings.ToLower(os.Getenv("DB_UNSAFE_MIGRATE")) == "true"
	// Configure migration options
	migrateOpts := []schema.MigrateOption{migrate.WithForeignKeys(false)}

	if unsafeMigrate {
		logger.Warn("UNSAFE DATABASE MIGRATE ENABLED - This should only be used in development")
		migrateOpts = append(migrateOpts, migrate.WithDropColumn(true), migrate.WithDropIndex(true))
	}
	if err := client.Schema.Create(ctx,
		migrateOpts...,
	); err != nil {
		return nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	logger.Info("Database connection established")
	return &Client{Ent: client, db: db}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	if c.Ent != nil {
		logger.Info("Closing database connection")
		return c.Ent.Close()
	}
	return nil
}

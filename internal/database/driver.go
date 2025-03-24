package database

import (
	"database/sql"

	"go-template/pkg/logger"

	"github.com/jackc/pgx/v5/stdlib"
)

// RegisterPgxDriver registers the pgx driver as "postgres" so Ent can use it.
func RegisterPgxDriver() {
	driverName := "postgres"

	// Create pgx driver
	pgxDriver := stdlib.GetDefaultDriver()

	// Register the driver
	sql.Register(driverName, pgxDriver)

	logger.Info("Registered pgx driver as 'postgres'")
}

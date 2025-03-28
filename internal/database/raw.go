package database

import (
	"context"
	"database/sql"
	"fmt"

	"go-template/pkg/logger"
)

// RawDB returns the underlying sql.DB instance
func (c *Client) RawDB() *sql.DB {
	return c.db
}

// ExecContext executes a raw SQL query
func (c *Client) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	logger.Debugf("Executing raw SQL: %s", query)
	return c.db.ExecContext(ctx, query, args...)
}

// QueryContext executes a raw SQL query and returns the rows
func (c *Client) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	logger.Debugf("Querying raw SQL: %s", query)
	return c.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a raw SQL query that returns a single row
func (c *Client) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	logger.Debugf("Querying raw SQL row: %s", query)
	return c.db.QueryRowContext(ctx, query, args...)
}

// Transaction executes the given function within a transaction
func (c *Client) Transaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

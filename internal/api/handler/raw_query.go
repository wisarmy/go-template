package handler

import (
	"database/sql"
	"encoding/json"
	"go-template/internal/api/response"
	"go-template/internal/database"
	"go-template/pkg/errcode"
	"go-template/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

// RawQueryHandler handles raw SQL query operations
type RawQueryHandler struct {
	db *database.Client
}

// NewRawQueryHandler creates a new raw query handler
func NewRawQueryHandler(db *database.Client) *RawQueryHandler {
	return &RawQueryHandler{db: db}
}

// UserStatsDTO represents user statistics
type UserStatsDTO struct {
	TotalUsers       int       `json:"total_users"`
	ActiveUsers      int       `json:"active_users"`
	DisabledUsers    int       `json:"disabled_users"`
	NewestUserDate   time.Time `json:"newest_user_date"`
	UsersPerRoleJSON string    `json:"users_per_role_json"`
}

// GetUserStats godoc
// @Summary      Get user statistics
// @Description  Get user statistics using raw SQL queries
// @Tags         raw-queries
// @Accept       json
// @Produce      json
// @Success      200  {object}   response.Response{data=UserStatsDTO} "ok"
// @Failure      500  {object}   response.Response "server.error"
// @Router       /raw/user-stats [get]
// @Security     BearerAuth
func (h *RawQueryHandler) GetUserStats(c *gin.Context) {
	ctx := c.Request.Context()

	var stats UserStatsDTO

	// Example of a raw SQL query with multiple result columns
	err := h.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) as total_users,
			SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) as active_users,
			SUM(CASE WHEN status = 'disabled' THEN 1 ELSE 0 END) as disabled_users,
			MAX(created_at) as newest_user_date
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		GROUP BY r.id
	`).Scan(
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.DisabledUsers,
		&stats.NewestUserDate,
	)

	// Handle potential error cases
	if err != nil {
		if err == sql.ErrNoRows {
			// No users in database
			stats = UserStatsDTO{
				TotalUsers:    0,
				ActiveUsers:   0,
				DisabledUsers: 0,
			}
		} else {
			logger.Errorf("Failed to execute raw query: %v", err)
			response.Err(c, errcode.ServerError, "Failed to fetch user statistics")
			return
		}
	}

	err = h.db.QueryRowContext(ctx, `
        SELECT json_agg(role_counts)::text
        FROM (
            SELECT
                r.name as role,
                COUNT(*) as count
            FROM
                users u
            LEFT JOIN
                roles r ON u.role_id = r.id
            GROUP BY
                r.id, r.name
        ) role_counts
    `).Scan(&stats.UsersPerRoleJSON)

	response.Ok(c, stats)
}

// RoleUserCountDTO represents a role with user count
type RoleUserCountDTO struct {
	RoleID      int    `json:"role_id"`
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
	UserCount   int    `json:"user_count"`
}

// GetRoleUserCounts godoc
// @Summary      Get role user counts
// @Description  Get counts of users per role using raw SQL
// @Tags         raw-queries
// @Accept       json
// @Produce      json
// @Success      200  {object}   response.Response{data=[]RoleUserCountDTO} "ok"
// @Failure      500  {object}   response.Response "server.error"
// @Router       /raw/role-user-counts [get]
// @Security     BearerAuth
func (h *RawQueryHandler) GetRoleUserCounts(c *gin.Context) {
	ctx := c.Request.Context()

	// Query database
	rows, err := h.db.QueryContext(ctx, `
		SELECT
			r.id as role_id,
			r.name as role_name,
			r.description,
			COUNT(u.id) as user_count
		FROM roles r
		LEFT JOIN users u ON r.id = u.role_id
		GROUP BY r.id
		ORDER BY user_count DESC
	`)

	if err != nil {
		logger.Errorf("Failed to execute raw query: %v", err)
		response.Err(c, errcode.ServerError, "Failed to fetch role user counts")
		return
	}
	defer rows.Close()

	// Manually scan results
	var results []RoleUserCountDTO
	for rows.Next() {
		var dto RoleUserCountDTO
		if err := rows.Scan(&dto.RoleID, &dto.RoleName, &dto.Description, &dto.UserCount); err != nil {
			logger.Errorf("Failed to scan row data: %v", err)
			response.Err(c, errcode.ServerError, "Failed to process query results")
			return
		}
		results = append(results, dto)
	}

	if err := rows.Err(); err != nil {
		logger.Errorf("Failed to read result set: %v", err)
		response.Err(c, errcode.ServerError, "Failed to read query results")
		return
	}

	response.Ok(c, results)
}

// ExecuteUpdateExample godoc
// @Summary      Execute update example
// @Description  Example of executing a raw SQL update query
// @Tags         raw-queries
// @Accept       json
// @Produce      json
// @Param        role_name query string true "Role name to update users for"
// @Success      200  {object}   response.Response{data=map[string]int} "ok"
// @Failure      500  {object}   response.Response "server.error | invalid.params"
// @Router       /raw/execute-update [post]
// @Security     BearerAuth
func (h *RawQueryHandler) ExecuteUpdateExample(c *gin.Context) {
	roleName := c.Query("role_name")
	if roleName == "" {
		response.Err(c, errcode.InvalidParams, "role_name query parameter is required")
		return
	}

	ctx := c.Request.Context()

	// Example of a transaction with raw SQL
	err := h.db.Transaction(ctx, func(tx *sql.Tx) error {
		// First get the role ID
		var roleID int
		err := tx.QueryRowContext(ctx, "SELECT id FROM roles WHERE name = $1", roleName).Scan(&roleID)
		if err != nil {
			if err == sql.ErrNoRows {
				return errcode.New(errcode.RoleNotFound)
			}
			return err
		}

		// Then update users with that role
		res, err := tx.ExecContext(ctx,
			"UPDATE users SET status = 'active', updated_at = NOW() WHERE role_id = $1 AND status = 'disabled'",
			roleID)
		if err != nil {
			return err
		}

		// Return success even if no rows were affected
		_, err = res.RowsAffected()
		return err
	})

	if err != nil {
		if e, ok := err.(*errcode.Error); ok {
			response.Err(c, e.Code, e.Message)
			return
		}
		logger.Errorf("Failed to execute update: %v", err)
		response.Err(c, errcode.ServerError, "Failed to update users")
		return
	}

	// Get active count to return in response
	var activeCount int
	err = h.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users u JOIN roles r ON u.role_id = r.id WHERE r.name = $1 AND u.status = 'active'",
		roleName).Scan(&activeCount)

	if err != nil {
		logger.Errorf("Failed to get updated count: %v", err)
		// Still return success for the update operation
		response.OkWithMessage(c, "Users updated successfully", nil)
		return
	}

	response.Ok(c, map[string]int{
		"active_users_count": activeCount,
	})
}

// RawDataStructsExample godoc
// @Summary      Complex JSON data example
// @Description  Shows how to handle complex JSON data returned from raw SQL
// @Tags         raw-queries
// @Accept       json
// @Produce      json
// @Success      200  {object}   response.Response{data=map[string]interface{}} "ok"
// @Failure      500  {object}   response.Response "server.error"
// @Router       /raw/data-structures [get]
// @Security     BearerAuth
func (h *RawQueryHandler) RawDataStructsExample(c *gin.Context) {
	ctx := c.Request.Context()

	// Example: Using PostgreSQL's JSON capabilities
	var jsonData string
	err := h.db.QueryRowContext(ctx, `
		SELECT json_build_object(
		    'id', u.id,
		    'name', u.name,
		    'email', u.email,
		    'created_at', u.created_at
		) AS user_data
		FROM users u
		LIMIT 1
	`).Scan(&jsonData)

	if err != nil {
		logger.Errorf("Failed to execute complex JSON query: %v", err)
		response.Err(c, errcode.ServerError, "Failed to fetch data structures")
		return
	}

	// Parse JSON string to Go structure
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		logger.Errorf("Failed to parse JSON: %v", err)
		response.Err(c, errcode.ServerError, "Failed to process query result")
		return
	}

	response.Ok(c, result)
}

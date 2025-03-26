package handler

import (
	"go-template/ent"
	"go-template/ent/role"
	"go-template/ent/user"
	"go-template/internal/api/response"
	"go-template/internal/database"
	"go-template/pkg/errcode"
	"go-template/pkg/logger"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RoleHandler handles role-related HTTP requests
type RoleHandler struct {
	db *database.Client
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(db *database.Client) *RoleHandler {
	return &RoleHandler{db: db}
}

// List returns a list of roles
func (h *RoleHandler) List(c *gin.Context) {
	// Check if we should include users information
	withUsers := c.Query("with_users") == "true"

	// Build the query
	query := h.db.Ent.Role.Query()

	// Include users information if requested
	if withUsers {
		query = query.WithUsers()
	}

	// Execute the query
	roles, err := query.All(c.Request.Context())
	if err != nil {
		logger.Errorf("Failed to fetch roles: %v", err)
		response.Err(c, errcode.ServerError, "Failed to fetch roles")
		return
	}

	response.Ok(c, roles)
}

// Get returns a specific role by ID
func (h *RoleHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Err(c, errcode.InvalidParams, "Invalid role ID")
		return
	}

	// Check if we should include users information
	withUsers := c.Query("with_users") == "true"

	// Build the query
	query := h.db.Ent.Role.Query().Where(role.ID(id))

	// Include users information if requested
	if withUsers {
		query = query.WithUsers()
	}

	// Execute the query
	r, err := query.Only(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			response.Err(c, errcode.RoleNotFound)
			return
		}
		logger.Errorf("Failed to fetch role: %v", err)
		response.Err(c, errcode.ServerError, "Failed to fetch role")
		return
	}

	response.Ok(c, r)
}

// Create creates a new role
func (h *RoleHandler) Create(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, errcode.InvalidParams, err.Error())
		return
	}

	// Create role
	r, err := h.db.Ent.Role.Create().
		SetName(input.Name).
		SetDescription(input.Description).
		Save(c.Request.Context())

	if err != nil {
		logger.Errorf("Failed to create role: %v", err)
		response.Err(c, errcode.ServerError, "Failed to create role")
		return
	}

	response.Ok(c, r)
}

// Update updates an existing role
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Err(c, errcode.InvalidParams, "Invalid role ID")
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, errcode.InvalidParams, err.Error())
		return
	}

	// Start building the update query
	update := h.db.Ent.Role.UpdateOneID(id)

	// Only set fields that were provided
	if input.Name != "" {
		update = update.SetName(input.Name)
	}
	if input.Description != "" {
		update = update.SetDescription(input.Description)
	}

	// Execute update
	r, err := update.Save(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			response.Err(c, errcode.RoleNotFound)
			return
		}
		logger.Errorf("Failed to update role: %v", err)
		response.Err(c, errcode.ServerError, "Failed to update role")
		return
	}

	response.Ok(c, r)
}

// Delete deletes a role by ID
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Err(c, errcode.InvalidParams, "Invalid role ID")
		return
	}

	// Check if role is assigned to any users
	count, err := h.db.Ent.User.Query().
		Where(user.RoleID(id)).
		Count(c.Request.Context())

	if err != nil {
		logger.Errorf("Failed to check role usage: %v", err)
		response.Err(c, errcode.ServerError, "Failed to check role usage")
		return
	}

	// Prevent deletion if role is in use
	if count > 0 {
		response.ErrWithData(c, errcode.RoleInUse, map[string]interface{}{
			"count": count,
		}, "Cannot delete role that is assigned to users")
		return
	}

	// Delete the role
	err = h.db.Ent.Role.DeleteOneID(id).Exec(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			response.Err(c, errcode.RoleNotFound)
			return
		}
		logger.Errorf("Failed to delete role: %v", err)
		response.Err(c, errcode.ServerError, "Failed to delete role")
		return
	}

	response.OkWithMessage(c, "Role deleted successfully", nil)
}

// GetUsers returns all users with a specific role
func (h *RoleHandler) GetUsers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Err(c, errcode.InvalidParams, "Invalid role ID")
		return
	}

	// Query users with this role
	users, err := h.db.Ent.Role.Query().
		Where(role.ID(id)).
		QueryUsers().
		All(c.Request.Context())

	if err != nil {
		if ent.IsNotFound(err) {
			response.Err(c, errcode.RoleNotFound)
			return
		}
		logger.Errorf("Failed to fetch users for role: %v", err)
		response.Err(c, errcode.ServerError, "Failed to fetch users")
		return
	}

	response.Ok(c, users)
}

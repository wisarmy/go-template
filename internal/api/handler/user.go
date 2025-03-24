package handler

import (
	"go-template/ent"
	"go-template/ent/role"
	"go-template/ent/user"
	"go-template/internal/database"
	"go-template/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	db *database.Client
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *database.Client) *UserHandler {
	return &UserHandler{db: db}
}

// List returns a list of users
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.db.Ent.User.Query().WithRole().All(c.Request.Context())
	if err != nil {
		logger.Errorf("Failed to fetch users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Get returns a specific user by ID
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.db.Ent.User.Query().WithRole().Where(user.ID(id)).Only(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		logger.Errorf("Failed to fetch user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Create creates a new user
func (h *UserHandler) Create(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		RoleID   int    `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.db.Ent.User.Create().
		SetName(input.Name).
		SetEmail(input.Email).
		SetPassword(input.Password).
		SetRoleID(input.RoleID).
		Save(c.Request.Context())

	if err != nil {
		logger.Errorf("Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Update updates an existing user
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		RoleID   *int   `json:"role_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start building the update query
	update := h.db.Ent.User.UpdateOneID(id)

	// Only set fields that were provided
	if input.Name != "" {
		update = update.SetName(input.Name)
	}
	if input.Email != "" {
		update = update.SetEmail(input.Email)
	}
	if input.Password != "" {
		update = update.SetPassword(input.Password)
	}
	// Handle role relationship
	if input.RoleID != nil {
		if *input.RoleID > 0 {
			// Check if the role exists
			exists, err := h.db.Ent.Role.Query().
				Where(role.ID(*input.RoleID)).
				Exist(c.Request.Context())

			if err != nil {
				logger.Errorf("Failed to check role existence: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate role"})
				return
			}

			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Specified role does not exist"})
				return
			}

			// Set role ID
			update = update.SetRoleID(*input.RoleID)
		} else {
			// If role ID is 0 or negative, clear the role relationship
			update = update.ClearRole()
		}
	}
	// Execute update
	user, err := update.Save(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		logger.Errorf("Failed to update user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete deletes a user by ID
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.db.Ent.User.DeleteOneID(id).Exec(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		logger.Errorf("Failed to delete user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

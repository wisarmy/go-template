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
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	db *database.Client
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *database.Client) *UserHandler {
	return &UserHandler{db: db}
}

// List godoc
// @Summary      List users
// @Description  get user list
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   response.Response{data=[]ent.User} "ok"
// @Failure      500  {object}   response.Response "server.error"
// @Router       /users [get]
// @Security     BearerAuth
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.db.Ent.User.Query().WithRole().All(c.Request.Context())
	if err != nil {
		logger.Errorf("Failed to fetch users: %v", err)
		response.Err(c, errcode.ServerError, "Failed to fetch users")
		return
	}
	response.Ok(c, users)
}

// Get godoc
// @Summary      Get a user
// @Description  get user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}   response.Response{data=ent.User} "ok"
// @Failure      500  {object}   response.Response "server.error ｜ invalid.params ｜ user.not_found"
// @Router       /users/{id} [get]
// @Security     BearerAuth
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Err(c, errcode.InvalidParams, "Invalid user ID")
		return
	}

	user, err := h.db.Ent.User.Query().WithRole().Where(user.ID(id)).Only(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			response.Err(c, errcode.UserNotFound)
			return
		}
		logger.Errorf("Failed to fetch user: %v", err)
		response.Err(c, errcode.ServerError, "Failed to fetch user")
		return
	}

	response.Ok(c, user)
}

// Create godoc
// @Summary      Create a user
// @Description  create a new user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      UserCreateInput  true  "User Info"
// @Success      200  {object}   response.Response{data=ent.User} "ok"
// @Failure      500  {object}   response.Response "server.error ｜ invalid.params"
// @Router       /users [post]
// @Security     BearerAuth
func (h *UserHandler) Create(c *gin.Context) {
	var input UserCreateInput

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, errcode.InvalidParams, err.Error())
		return
	}

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		logger.Errorf("Failed to hash password: %v", err)
		response.Err(c, errcode.ServerError, "Failed to process user data")
		return
	}
	user, err := h.db.Ent.User.Create().
		SetName(input.Name).
		SetEmail(input.Email).
		SetPassword(hashedPassword).
		SetRoleID(input.RoleID).
		Save(c.Request.Context())

	if err != nil {
		logger.Errorf("Failed to create user: %v", err)
		response.Err(c, errcode.ServerError, "Failed to create user")
		return
	}

	response.Ok(c, user)
}

// UserCreateInput represents the input for creating a user
type UserCreateInput struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"secret123"`
	RoleID   int    `json:"role_id" binding:"required" example:"1"`
}

// UserUpdateInput represents the input for updating a user
type UserUpdateInput struct {
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"newsecret123"`
	RoleID   *int   `json:"role_id" example:"2"`
}

// Update godoc
// @Summary      Update a user
// @Description  update an existing user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "User ID"
// @Param        user  body      UserUpdateInput  true  "User Info"
// @Success      200  {object}   response.Response{data=ent.User} "ok"
// @Failure      500  {object}   response.Response "server.error ｜ invalid.params | user.not_found | role.not_found"
// @Router       /users/{id} [put]
// @Security     BearerAuth
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Err(c, errcode.InvalidParams, "Invalid user ID")
		return
	}

	var input UserUpdateInput

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, errcode.InvalidParams, err.Error())
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
		hashedPassword, err := hashPassword(input.Password)
		if err != nil {
			logger.Errorf("Failed to hash password: %v", err)
			response.Err(c, errcode.ServerError, "Failed to process user data")
			return
		}
		update = update.SetPassword(hashedPassword)
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
				response.Err(c, errcode.ServerError, "Failed to validate role")
				return
			}

			if !exists {
				response.Err(c, errcode.RoleNotFound)
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
			response.Err(c, errcode.UserNotFound)
			return
		}
		logger.Errorf("Failed to update user: %v", err)
		response.Err(c, errcode.ServerError, "Failed to update user")
		return
	}

	response.Ok(c, user)
}

// Delete godoc
// @Summary      Delete a user
// @Description  delete a user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}   response.Response "ok"
// @Failure      500  {object}   response.Response "server.error ｜ invalid.params | user.not_found"
// @Router       /users/{id} [delete]
// @Security     BearerAuth
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Err(c, errcode.InvalidParams, "Invalid user ID")
		return
	}

	err = h.db.Ent.User.DeleteOneID(id).Exec(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			response.Err(c, errcode.UserNotFound)
			return
		}
		logger.Errorf("Failed to delete user: %v", err)
		response.Err(c, errcode.ServerError, "Failed to delete user")
		return
	}

	response.OkWithMessage(c, "User deleted successfully", nil)
}

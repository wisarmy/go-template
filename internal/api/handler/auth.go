package handler

import (
	"go-template/ent"
	"go-template/ent/role"
	"go-template/ent/user"
	"go-template/internal/api/response"
	"go-template/internal/database"
	"go-template/pkg/auth"
	"go-template/pkg/errcode"
	"go-template/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	db     *database.Client
	config auth.JWTConfig
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *database.Client, config auth.JWTConfig) *AuthHandler {
	return &AuthHandler{db: db, config: config}
}

// RegisterInput represents the input for user registration
type RegisterInput struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// Register godoc
// @Summary      Register new user
// @Description  Register a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterInput  true  "User registration data"
// @Success      200  {object}   response.Response{data=UserInfo} "ok"
// @Failure      500  {object}   response.Response "server.error ｜ invalid.params ｜ user.register.error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, errcode.InvalidParams, err.Error())
		return
	}

	// Check if user already exists
	exists, err := h.db.Ent.User.Query().
		Where(user.EmailEQ(input.Email)).
		Exist(c.Request.Context())

	if err != nil {
		logger.Errorf("Failed to check user existence: %v", err)
		response.Err(c, errcode.ServerError, "Failed to process registration")
		return
	}

	if exists {
		response.Err(c, errcode.UserRegisterError, "Email already registered")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("Failed to hash password: %v", err)
		response.Err(c, errcode.ServerError, "Failed to process registration")
		return
	}

	// Find default user role
	defaultRole, err := h.db.Ent.Role.Query().
		Where(role.NameEQ("user")).
		Only(c.Request.Context())

	if err != nil {
		if ent.IsNotFound(err) {
			// Create default role if it doesn't exist
			defaultRole, err = h.db.Ent.Role.Create().
				SetName("user").
				SetDescription("Regular user with standard permissions").
				Save(c.Request.Context())

			if err != nil {
				logger.Errorf("Failed to create default role: %v", err)
				response.Err(c, errcode.ServerError, "Failed to create user role")
				return
			}
		} else {
			logger.Errorf("Failed to fetch default role: %v", err)
			response.Err(c, errcode.ServerError, "Failed to process registration")
			return
		}
	}

	// Create user
	u, err := h.db.Ent.User.Create().
		SetName(input.Name).
		SetEmail(input.Email).
		SetPassword(string(hashedPassword)).
		SetRoleID(defaultRole.ID).
		Save(c.Request.Context())

	if err != nil {
		logger.Errorf("Failed to create user: %v", err)
		response.Err(c, errcode.UserRegisterError, "Failed to register user")
		return
	}

	// Return user info
	userInfo := UserInfo{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Role:  defaultRole.Name,
	}

	response.Ok(c, userInfo)
}

// LoginInput represents the input for user login
type LoginInput struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse represents the response from a successful login
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Login godoc
// @Summary      User login
// @Description  Authenticate a user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginInput  true  "Login credentials"
// @Success      200  {object}   response.Response{data=LoginResponse} "ok"
// @Failure      500  {object}   response.Response "server.error ｜ invalid.params ｜ user.login.error ｜ user.disabled"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, errcode.InvalidParams, err.Error())
		return
	}

	// Find user by email
	u, err := h.db.Ent.User.Query().
		Where(user.EmailEQ(input.Email)).
		WithRole().
		Only(c.Request.Context())

	if err != nil {
		if ent.IsNotFound(err) {
			response.Err(c, errcode.UserLoginError, "Invalid email or password")
			return
		}
		logger.Errorf("Failed to fetch user: %v", err)
		response.Err(c, errcode.ServerError, "Failed to authenticate user")
		return
	}

	// Check if user is disabled
	if u.Status == user.StatusDisabled {
		response.Err(c, errcode.UserDisabled)
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(input.Password))
	if err != nil {
		response.Err(c, errcode.UserLoginError, "Invalid email or password")
		return
	}

	// Generate JWT token
	roleName := "user"
	if u.Edges.Role != nil {
		roleName = u.Edges.Role.Name
	}

	token, err := auth.GenerateToken(u.ID, u.Name, roleName, h.config)
	if err != nil {
		logger.Errorf("Failed to generate token: %v", err)
		response.Err(c, errcode.ServerError, "Failed to generate authentication token")
		return
	}

	// Generate refresh token
	refreshToken, err := auth.GenerateRefreshToken(u.ID, u.Name, h.config)
	if err != nil {
		logger.Errorf("Failed to generate refresh token: %v", err)
		response.Err(c, errcode.ServerError, "Failed to generate refresh token")
		return
	}

	// Create response
	resp := LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(h.config.Expiration),
		User: UserInfo{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Role:  roleName,
		},
	}

	response.Ok(c, resp)
}

// RefreshInput represents the input for token refresh
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshToken godoc
// @Summary      Refresh token
// @Description  Refresh JWT token using a refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        refresh  body      RefreshInput  true  "Refresh token"
// @Success      200  {object}   response.Response{data=LoginResponse} "ok"
// @Failure      500  {object}   response.Response "server.error ｜ invalid.params ｜ auth.token.expired | auth.token.invalid | user.not_found | user.disabled"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input RefreshInput

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Err(c, errcode.InvalidParams, err.Error())
		return
	}

	// Parse refresh token
	userID, err := auth.ParseRefreshToken(input.RefreshToken, h.config.Secret)
	if err != nil {
		if err == auth.ErrExpiredToken {
			response.Err(c, errcode.AuthTokenExpired)
			return
		}
		response.Err(c, errcode.AuthTokenInvalid)
		return
	}

	// Get user information
	u, err := h.db.Ent.User.Query().
		Where(user.ID(userID)).
		WithRole().
		Only(c.Request.Context())

	if err != nil {
		if ent.IsNotFound(err) {
			response.Err(c, errcode.UserNotFound)
			return
		}
		logger.Errorf("Failed to fetch user: %v", err)
		response.Err(c, errcode.ServerError, "Failed to refresh token")
		return
	}

	// Check if user is disabled
	if u.Status == user.StatusDisabled {
		response.Err(c, errcode.UserDisabled)
		return
	}

	// Generate new JWT token
	roleName := "user"
	if u.Edges.Role != nil {
		roleName = u.Edges.Role.Name
	}

	token, err := auth.GenerateToken(u.ID, u.Name, roleName, h.config)
	if err != nil {
		logger.Errorf("Failed to generate token: %v", err)
		response.Err(c, errcode.ServerError, "Failed to generate authentication token")
		return
	}

	// Generate new refresh token
	refreshToken, err := auth.GenerateRefreshToken(u.ID, u.Name, h.config)
	if err != nil {
		logger.Errorf("Failed to generate refresh token: %v", err)
		response.Err(c, errcode.ServerError, "Failed to generate refresh token")
		return
	}

	// Create response
	resp := LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(h.config.Expiration),
		User: UserInfo{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Role:  roleName,
		},
	}

	response.Ok(c, resp)
}

// GetUserInfo godoc
// @Summary      Get current user info
// @Description  Returns information about the currently authenticated user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}   response.Response{data=UserInfo} "ok"
// @Failure      500  {object}   response.Response "server.error ｜ invalid.params ｜ user.unauthorized | user.not_found"
// @Router       /auth/me [get]
// @Security     BearerAuth
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	// Get user ID from context (set by JWTAuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.Err(c, errcode.UserUnauthorized)
		return
	}

	// Fetch user details
	u, err := h.db.Ent.User.Query().
		Where(user.ID(userID.(int))).
		WithRole().
		Only(c.Request.Context())

	if err != nil {
		if ent.IsNotFound(err) {
			response.Err(c, errcode.UserNotFound)
			return
		}
		logger.Errorf("Failed to fetch user: %v", err)
		response.Err(c, errcode.ServerError, "Failed to get user information")
		return
	}

	// Get role name
	roleName := "user"
	if u.Edges.Role != nil {
		roleName = u.Edges.Role.Name
	}

	// Return user info
	info := UserInfo{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Role:  roleName,
	}

	response.Ok(c, info)
}

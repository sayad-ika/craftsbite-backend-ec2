package handlers

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/utils"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// LoginRequest represents login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents registration request body
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService services.AuthService
	userService services.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService services.AuthService, userService services.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	response, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		utils.ErrorResponse(c, 401, "INVALID_CREDENTIALS", err.Error())
		return
	}

	setAuthCookie(c, response.Token, response.ExpiresAt)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user":       response.User,
			"expires_at": response.ExpiresAt,
		},
	})
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	// Validate role
	validRoles := map[string]bool{
		"employee":  true,
		"team_lead": true,
		"admin":     true,
		"logistics": true,
	}
	if !validRoles[req.Role] {
		utils.ErrorResponse(c, 400, "INVALID_ROLE", "Role must be one of: employee, team_lead, admin, logistics")
		return
	}

	// Create user using UserService
	userInput := services.CreateUserInput{
		Email:                 req.Email,
		Name:                  req.Name,
		Password:              req.Password,
		Role:                  models.Role(req.Role),
		DefaultMealPreference: "opt_in",
	}

	user, err := h.userService.CreateUser(userInput)
	if err != nil {
		utils.ErrorResponse(c, 400, "REGISTRATION_FAILED", err.Error())
		return
	}

	// Auto-login: generate token for the newly created user
	loginResponse, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		// User created but login failed - still return success with user data
		utils.SuccessResponse(c, 201, user, "User registered successfully. Please login.")
		return
	}

	utils.SuccessResponse(c, 201, loginResponse, "User registered and logged in successfully")
}

// Logout handles user logout (placeholder)
func (h *AuthHandler) Logout(c *gin.Context) {
	expireAuthCookie(c)
	utils.SuccessResponse(c, 200, nil, "Logout successful")
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	user, err := h.authService.GetCurrentUser(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, 404, "USER_NOT_FOUND", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, user, "User retrieved successfully")
}

func setAuthCookie(c *gin.Context, token string, expiresAt time.Time) {
	isProd := os.Getenv("ENV") == "production"

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true, // JS cannot read this
		Secure:   isProd,
		SameSite: http.SameSiteStrictMode,
	})
}

func expireAuthCookie(c *gin.Context) {
	isProd := os.Getenv("ENV") == "production"

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1, // browser deletes it immediately
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteStrictMode,
	})
}

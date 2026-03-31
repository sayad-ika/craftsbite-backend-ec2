package handlers

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// ListUsers lists all users (Admin only)
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.userService.ListUsers(nil)
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, users, "Users retrieved successfully")
}

// GetUser gets a user by ID (Admin or Self)
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	currentUserID, _ := c.Get("user_id")
	currentUserRole, _ := c.Get("role")

	// Check if user is accessing their own data or is admin
	if userID != currentUserID.(string) && currentUserRole.(string) != models.RoleAdmin.String() {
		utils.ErrorResponse(c, 403, "FORBIDDEN", "You can only access your own data")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		utils.ErrorResponse(c, 404, "USER_NOT_FOUND", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, user, "User retrieved successfully")
}

// CreateUser creates a new user (Admin only)
func (h *UserHandler) CreateUser(c *gin.Context) {
	var input services.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	user, err := h.userService.CreateUser(input)
	if err != nil {
		utils.ErrorResponse(c, 400, "CREATE_FAILED", err.Error())
		return
	}

	utils.SuccessResponse(c, 201, user, "User created successfully")
}

// UpdateUser updates a user (Admin or Self)
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	currentUserID, _ := c.Get("user_id")
	currentUserRole, _ := c.Get("role")

	// Check if user is updating their own data or is admin
	if userID != currentUserID.(string) && currentUserRole.(string) != models.RoleAdmin.String() {
		utils.ErrorResponse(c, 403, "FORBIDDEN", "You can only update your own data")
		return
	}

	var input services.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	// Non-admin users cannot change their role
	if currentUserRole.(string) != models.RoleAdmin.String() && input.Role != nil {
		utils.ErrorResponse(c, 403, "FORBIDDEN", "You cannot change your own role")
		return
	}

	user, err := h.userService.UpdateUser(userID, input)
	if err != nil {
		utils.ErrorResponse(c, 400, "UPDATE_FAILED", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, user, "User updated successfully")
}

// DeactivateUser deactivates a user (Admin only)
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	userID := c.Param("id")

	if err := h.userService.DeactivateUser(userID); err != nil {
		utils.ErrorResponse(c, 400, "DEACTIVATE_FAILED", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "User deactivated successfully")
}

// GetMyTeamMembers returns the team members for the authenticated team lead
// GET /api/v1/users/me/team-members
func (h *UserHandler) GetMyTeamMembers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	response, err := h.userService.GetMyTeamMembers(userID.(string))
	if err != nil {
		if err.Error() == "user is not a team lead" {
			utils.ErrorResponse(c, 403, "FORBIDDEN", "Only team leads can access this endpoint")
			return
		}
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, response, "Team members retrieved successfully")
}


func (h *UserHandler) GetMyTeam(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	team, err := h.userService.GetMyTeam(userID.(string))
	if err != nil {
		if err.Error() == "user is not an employee" {
			utils.ErrorResponse(c, 403, "FORBIDDEN", "Only employees can access this endpoint")
			return
		}
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, team, "My team retrieved successfully")
}

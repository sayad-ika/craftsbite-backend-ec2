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
// @Summary List all users
// @Description Get list of all users (Admin/Logistics only)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Users retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.userService.ListUsers(nil)
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, users, "Users retrieved successfully")
}

// GetUser gets a user by ID (Admin or Self)
// @Summary Get user by ID
// @Description Get user details (Admin or own profile)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "User retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /users/{id} [get]
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
// @Summary Create new user
// @Description Create a new user account (Admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateUserInput true "User details"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]interface{} "Validation error or creation failed"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /users [post]
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
// @Summary Update user
// @Description Update user details (Admin or own profile)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body services.UpdateUserInput true "Update details"
// @Success 200 {object} map[string]interface{} "User updated successfully"
// @Failure 400 {object} map[string]interface{} "Validation error or update failed"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /users/{id} [put]
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
// @Summary Deactivate user
// @Description Deactivate a user account (Admin only)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "User deactivated successfully"
// @Failure 400 {object} map[string]interface{} "Deactivation failed"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /users/{id} [delete]
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	userID := c.Param("id")

	if err := h.userService.DeactivateUser(userID); err != nil {
		utils.ErrorResponse(c, 400, "DEACTIVATE_FAILED", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "User deactivated successfully")
}

// GetMyTeamMembers returns the team members for the authenticated team lead
// @Summary Get my team members
// @Description Get team members for authenticated team lead
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Team members retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - only team leads"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /users/me/team-members [get]
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


// GetMyTeam returns the team for the authenticated employee
// @Summary Get my team
// @Description Get team details for authenticated employee
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Team retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - only employees"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /users/me/team [get]
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

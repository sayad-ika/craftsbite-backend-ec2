package handlers

import (
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// PreferenceHandler handles user preference endpoints
type PreferenceHandler struct {
	preferenceService services.PreferenceService
}

// NewPreferenceHandler creates a new preference handler
func NewPreferenceHandler(preferenceService services.PreferenceService) *PreferenceHandler {
	return &PreferenceHandler{
		preferenceService: preferenceService,
	}
}

// GetPreferences returns the current user's meal preferences
// @Summary Get user preferences
// @Description Get current user's meal preferences
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Preferences retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /users/me/preferences [get]
func (h *PreferenceHandler) GetPreferences(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	preferences, err := h.preferenceService.GetPreferences(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, preferences, "Preferences retrieved successfully")
}

// UpdatePreferencesRequest represents the request body for updating preferences
type UpdatePreferencesRequest struct {
	DefaultMealPreference string `json:"default_meal_preference" binding:"required"`
}

// UpdatePreferences updates the current user's meal preferences
// @Summary Update user preferences
// @Description Update current user's meal preferences
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdatePreferencesRequest true "Preference details"
// @Success 200 {object} map[string]interface{} "Preferences updated successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /users/me/preferences [put]
func (h *PreferenceHandler) UpdatePreferences(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse request body
	var req UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	// Update preference
	err := h.preferenceService.UpdateDefaultPreference(userID.(string), req.DefaultMealPreference)
	if err != nil {
		utils.ErrorResponse(c, 400, "UPDATE_PREFERENCE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "Preferences updated successfully")
}

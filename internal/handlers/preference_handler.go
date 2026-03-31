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
// GET /api/v1/users/me/preferences
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
// PUT /api/v1/users/me/preferences
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

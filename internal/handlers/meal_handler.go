package handlers

import (
	"craftsbite-backend/internal/repository"
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/sse"
	"craftsbite-backend/internal/utils"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
)

// MealHandler handles meal participation endpoints
type MealHandler struct {
	mealService      services.MealService
	teamRepo         repository.TeamRepository
	headcountService services.HeadcountService
	hub              *sse.Hub
}

// NewMealHandler creates a new meal handler
func NewMealHandler(mealService services.MealService, teamRepo repository.TeamRepository, headcountService services.HeadcountService, hub *sse.Hub) *MealHandler {
	return &MealHandler{
		mealService:      mealService,
		teamRepo:         teamRepo,
		headcountService: headcountService,
		hub:              hub,
	}
}

// GetTodayMeals returns today's meals and participation status
// @Summary Get today's meals
// @Description Get tomorrow's meal schedule and user participation status
// @Tags meals
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Meals retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /meals/today [get]
func (h *MealHandler) GetTodayMeals(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	response, err := h.mealService.GetTodayMeals(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, response, "Tomorrow's meals retrieved successfully")
}

// GetParticipationByDate returns participation status for a specific date
// @Summary Get participation by date
// @Description Get user's meal participation status for a specific date
// @Tags meals
// @Produce json
// @Security BearerAuth
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Participation retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /meals/participation/{date} [get]
func (h *MealHandler) GetParticipationByDate(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Get date from URL parameter
	date := c.Param("date")
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Date parameter is required")
		return
	}

	participations, err := h.mealService.GetParticipation(userID.(string), date)
	if err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, participations, "Participation retrieved successfully")
}

// SetParticipationRequest represents the request body for setting participation
type SetParticipationRequest struct {
	Date          string `json:"date" binding:"required"`
	MealType      string `json:"meal_type" binding:"required"`
	Participating *bool  `json:"participating" binding:"required"`
}

// SetParticipation sets or updates a user's participation
// @Summary Set meal participation
// @Description Set or update user's participation for a specific meal
// @Tags meals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SetParticipationRequest true "Participation details"
// @Success 200 {object} map[string]interface{} "Participation updated successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /meals/participation [post]
func (h *MealHandler) SetParticipation(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse request body
	var req SetParticipationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	// Set participation (dereference the pointer)
	err := h.mealService.SetParticipation(userID.(string), req.Date, req.MealType, *req.Participating)
	if err != nil {
		utils.ErrorResponse(c, 400, "SET_PARTICIPATION_ERROR", err.Error())
		return
	}

	if summary, broadcastErr := h.headcountService.GetHeadcountByDate(req.Date); broadcastErr == nil {
		if payload, marshalErr := json.Marshal(summary); marshalErr == nil {
			h.hub.Broadcast(req.Date, string(payload))
		}
	}

	utils.SuccessResponse(c, 200, nil, "Participation updated successfully")
}

// OverrideParticipationRequest represents the request body for admin override
type OverrideParticipationRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	Date          string `json:"date" binding:"required"`
	MealType      string `json:"meal_type" binding:"required"`
	Participating *bool  `json:"participating" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
}

// OverrideParticipation allows admins to override a user's participation
// @Summary Override participation (Admin/Team Lead)
// @Description Admin or team lead can override user's meal participation
// @Tags meals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body OverrideParticipationRequest true "Override details"
// @Success 200 {object} map[string]interface{} "Participation overridden successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /meals/participation/override [post]
func (h *MealHandler) OverrideParticipation(c *gin.Context) {
	// Get admin ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse request body
	var req OverrideParticipationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	// Override participation (dereference the pointer)
	err := h.mealService.OverrideParticipation(
		adminID.(string),
		req.UserID,
		req.Date,
		req.MealType,
		*req.Participating,
		req.Reason,
	)
	if err != nil {
		utils.ErrorResponse(c, 400, "OVERRIDE_ERROR", err.Error())
		return
	}

	if summary, broadcastErr := h.headcountService.GetHeadcountByDate(req.Date); broadcastErr == nil {
		if payload, marshalErr := json.Marshal(summary); marshalErr == nil {
			h.hub.Broadcast(req.Date, string(payload))
		}
	}

	utils.SuccessResponse(c, 200, nil, "Participation overridden successfully")
}

// GetTeamParticipation returns team participation for today
// @Summary Get team participation
// @Description Get today's meal participation for team lead's team
// @Tags meals
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Team participation retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /meals/team-participation [get]
func (h *MealHandler) GetTeamParticipation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	today := time.Now().Format("2006-01-02")

	response, err := h.mealService.GetTeamParticipation(userID.(string), today)

	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, response, "Team participation retrieved successfully")
}

// GetAllTeamsParticipation returns all teams participation for today
// @Summary Get all teams participation
// @Description Get today's meal participation for all teams (Admin/Logistics only)
// @Tags meals
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "All teams participation retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /meals/all-teams-participation [get]
func (h *MealHandler) GetAllTeamsParticipation(c *gin.Context) {
	today := time.Now().Format("2006-01-02")

	response, err := h.mealService.GetAllTeamsParticipation(today)
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, response, "All teams participation retrieved successfully")
}

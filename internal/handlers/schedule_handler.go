package handlers

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// ScheduleHandler handles day schedule endpoints
type ScheduleHandler struct {
	scheduleService services.ScheduleService
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(scheduleService services.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
	}
}

// GetSchedule returns a day schedule for a specific date
// @Summary Get schedule by date
// @Description Get meal schedule for a specific date
// @Tags schedules
// @Produce json
// @Security BearerAuth
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Schedule retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /schedules/{date} [get]
func (h *ScheduleHandler) GetSchedule(c *gin.Context) {
	// Get date from URL parameter
	date := c.Param("date")
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Date parameter is required")
		return
	}

	schedule, err := h.scheduleService.GetSchedule(date)
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	if schedule == nil {
		utils.SuccessResponse(c, 200, nil, "No schedule found for this date")
		return
	}

	utils.SuccessResponse(c, 200, schedule, "Schedule retrieved successfully")
}

// GetScheduleRange returns schedules within a date range
// @Summary Get schedule range
// @Description Get meal schedules within a date range
// @Tags schedules
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Schedules retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /schedules/range [get]
func (h *ScheduleHandler) GetScheduleRange(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "start_date and end_date query parameters are required")
		return
	}

	schedules, err := h.scheduleService.GetScheduleRange(startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, schedules, "Schedules retrieved successfully")
}

// CreateScheduleRequest represents the request body for creating a schedule
type CreateScheduleRequest struct {
	Date           string   `json:"date" binding:"required"`
	DayStatus      string   `json:"day_status" binding:"required"`
	Reason         string   `json:"reason"`
	AvailableMeals []string `json:"available_meals"`
}

// CreateSchedule creates a new day schedule (Admin and Logistics only)
// @Summary Create schedule
// @Description Create a new meal schedule (Admin/Logistics only)
// @Tags schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateScheduleRequest true "Schedule details"
// @Success 201 {object} map[string]interface{} "Schedule created successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /schedules [post]
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse request body
	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	// Convert to service input
	input := services.CreateScheduleInput{
		Date:      req.Date,
		DayStatus: models.DayStatus(req.DayStatus),
		Reason:    req.Reason,
	}

	// Convert meal types
	if len(req.AvailableMeals) > 0 {
		for _, meal := range req.AvailableMeals {
			input.AvailableMeals = append(input.AvailableMeals, models.MealType(meal))
		}
	}

	schedule, err := h.scheduleService.CreateSchedule(userID.(string), input)
	if err != nil {
		utils.ErrorResponse(c, 400, "CREATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 201, schedule, "Schedule created successfully")
}

// UpdateScheduleRequest represents the request body for updating a schedule
type UpdateScheduleRequest struct {
	DayStatus      *string   `json:"day_status"`
	Reason         *string   `json:"reason"`
	AvailableMeals *[]string `json:"available_meals"`
}

// UpdateSchedule updates an existing day schedule (Admin and Logistics only)
// @Summary Update schedule
// @Description Update an existing meal schedule (Admin/Logistics only)
// @Tags schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param date path string true "Date (YYYY-MM-DD)"
// @Param request body UpdateScheduleRequest true "Update details"
// @Success 200 {object} map[string]interface{} "Schedule updated successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /schedules/{date} [put]
func (h *ScheduleHandler) UpdateSchedule(c *gin.Context) {
	// Get date from URL parameter
	date := c.Param("date")
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Date parameter is required")
		return
	}

	// Parse request body
	var req UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	// Convert to service input
	input := services.UpdateScheduleInput{}

	if req.DayStatus != nil {
		status := models.DayStatus(*req.DayStatus)
		input.DayStatus = &status
	}

	if req.Reason != nil {
		input.Reason = req.Reason
	}

	if req.AvailableMeals != nil {
		meals := []models.MealType{}
		for _, meal := range *req.AvailableMeals {
			meals = append(meals, models.MealType(meal))
		}
		input.AvailableMeals = &meals
	}

	schedule, err := h.scheduleService.UpdateSchedule(date, input)
	if err != nil {
		utils.ErrorResponse(c, 400, "UPDATE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, schedule, "Schedule updated successfully")
}

// DeleteSchedule deletes a day schedule (Admin and Logistics only)
// @Summary Delete schedule
// @Description Delete a meal schedule (Admin/Logistics only)
// @Tags schedules
// @Produce json
// @Security BearerAuth
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Schedule deleted successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /schedules/{date} [delete]
func (h *ScheduleHandler) DeleteSchedule(c *gin.Context) {
	//Get date from URL parameter
	date := c.Param("date")
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Date parameter is required")
		return
	}

	err := h.scheduleService.DeleteSchedule(date)
	if err != nil {
		utils.ErrorResponse(c, 400, "DELETE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "Schedule deleted successfully")
}

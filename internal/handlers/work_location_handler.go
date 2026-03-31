package handlers

import (
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type WorkLocationHandler struct {
	svc services.WorkLocationService
}

func NewWorkLocationHandler(svc services.WorkLocationService) *WorkLocationHandler {
	return &WorkLocationHandler{svc: svc}
}

type setLocationRequest struct {
	Date     string `json:"date" binding:"required"`
	Location string `json:"location" binding:"required"`
}

// SetMyWorkLocation sets work location for current user
// @Summary Set my work location
// @Description Set work location (office/wfh) for current user on a specific date
// @Tags work-location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body setLocationRequest true "Location details"
// @Success 200 {object} map[string]interface{} "Work location updated successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /work-location [post]
func (h *WorkLocationHandler) SetMyWorkLocation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	var req setLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	if err := h.svc.SetMyLocation(userID.(string), req.Date, req.Location); err != nil {
		utils.ErrorResponse(c, 400, "SET_LOCATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "Work location updated successfully")
}

// GetMyWorkLocation gets work location for current user
// @Summary Get my work location
// @Description Get work location for current user on a specific date
// @Tags work-location
// @Produce json
// @Security BearerAuth
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Work location retrieved"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /work-location [get]
func (h *WorkLocationHandler) GetMyWorkLocation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	date := c.Query("date")
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Query param 'date' is required")
		return
	}

	resp, err := h.svc.GetMyLocation(userID.(string), date)
	if err != nil {
		utils.ErrorResponse(c, 400, "GET_LOCATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, resp, "Work location retrieved")
}

type adminSetLocationRequest struct {
	UserID   string  `json:"user_id" binding:"required"`
	Date     string  `json:"date" binding:"required"`
	Location string  `json:"location" binding:"required"`
	Reason   *string `json:"reason"`
}

// SetWorkLocationFor sets work location for another user (Admin/Team Lead)
// @Summary Override work location
// @Description Set work location for another user (Admin/Team Lead only)
// @Tags work-location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body adminSetLocationRequest true "Override details"
// @Success 200 {object} map[string]interface{} "Work location corrected successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /work-location/override [post]
func (h *WorkLocationHandler) SetWorkLocationFor(c *gin.Context) {
	requesterID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	var req adminSetLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	if err := h.svc.SetLocationFor(requesterID.(string), req.UserID, req.Date, req.Location, req.Reason); err != nil {
		utils.ErrorResponse(c, 400, "SET_LOCATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "Work location corrected successfully")
}

// ListWorkLocationsByDate lists all work locations for a specific date
// @Summary List work locations by date
// @Description Get all users' work locations for a specific date (Admin/Team Lead only)
// @Tags work-location
// @Produce json
// @Security BearerAuth
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Work locations retrieved"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /work-location/list [get]
func (h *WorkLocationHandler) ListWorkLocationsByDate(c *gin.Context) {
	requesterID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	date := c.Query("date")
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Query param 'date' is required")
		return
	}

	result, err := h.svc.ListByDate(requesterID.(string), date)
	if err != nil {
		utils.ErrorResponse(c, 400, "LIST_LOCATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, result, "Work locations retrieved")
}

// GetMonthlySummary gets monthly WFH summary for current user
// @Summary Get monthly WFH summary
// @Description Get monthly work-from-home summary for current user
// @Tags work-location
// @Produce json
// @Security BearerAuth
// @Param month query string false "Month (YYYY-MM, defaults to current month)"
// @Success 200 {object} map[string]interface{} "Monthly WFH summary retrieved"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /work-location/monthly-summary [get]
func (h *WorkLocationHandler) GetMonthlySummary(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
        return
    }

    yearMonth := c.Query("month")
    if yearMonth == "" {
        yearMonth = time.Now().Format("2006-01")
    }

    summary, err := h.svc.GetMonthlySummary(userID.(string), yearMonth)
    if err != nil {
        utils.ErrorResponse(c, 400, "SUMMARY_ERROR", err.Error())
        return
    }

    utils.SuccessResponse(c, 200, summary, "Monthly WFH summary retrieved")
}

// GetTeamMonthlyReport gets monthly WFH report for team
// @Summary Get team monthly WFH report
// @Description Get monthly work-from-home report for team (Admin/Logistics/Team Lead only)
// @Tags work-location
// @Produce json
// @Security BearerAuth
// @Param month query string false "Month (YYYY-MM, defaults to current month)"
// @Success 200 {object} map[string]interface{} "Monthly WFH report retrieved"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /work-location/team-monthly-report [get]
func (h *WorkLocationHandler) GetTeamMonthlyReport(c *gin.Context) {
    requesterID, exists := c.Get("user_id")
    if !exists {
        utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
        return
    }

    yearMonth := c.Query("month")
    if yearMonth == "" {
        yearMonth = time.Now().Format("2006-01")
    }

    rollup, err := h.svc.GetTeamMonthlyReport(requesterID.(string), yearMonth)
    if err != nil {
        utils.ErrorResponse(c, 400, "ROLLUP_ERROR", err.Error())
        return
    }

    utils.SuccessResponse(c, 200, rollup, "Monthly WFH report retrieved")
}

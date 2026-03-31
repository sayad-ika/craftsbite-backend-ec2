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

// POST /api/v1/work-location
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

// GET /api/v1/work-location?date=YYYY-MM-DD
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

// POST /api/v1/work-location/override
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

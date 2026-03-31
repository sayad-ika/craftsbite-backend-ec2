package handlers

import (
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/sse"
	"craftsbite-backend/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HeadcountHandler handles headcount reporting endpoints
type HeadcountHandler struct {
	headcountService services.HeadcountService
	hub              *sse.Hub
}

// NewHeadcountHandler creates a new headcount handler
func NewHeadcountHandler(headcountService services.HeadcountService, hub *sse.Hub) *HeadcountHandler {
	return &HeadcountHandler{
		headcountService: headcountService,
		hub:              hub,
	}
}

// GetTodayHeadcount returns today's and tomorrow's headcount summary
// @Summary Get today's headcount
// @Description Get today's and tomorrow's meal headcount summary (Admin/Logistics only)
// @Tags headcount
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Headcount retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /headcount/today [get]
func (h *HeadcountHandler) GetTodayHeadcount(c *gin.Context) {
	summary, err := h.headcountService.GetTodayHeadcount()
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, summary, "Today's and tomorrow's headcount retrieved successfully")
}

// GetHeadcountByDate returns headcount summary for a specific date
// @Summary Get headcount by date
// @Description Get meal headcount summary for a specific date (Admin/Logistics only)
// @Tags headcount
// @Produce json
// @Security BearerAuth
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Headcount retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /headcount/{date} [get]
func (h *HeadcountHandler) GetHeadcountByDate(c *gin.Context) {
	// Get date from URL parameter
	date := c.Param("date")
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Date parameter is required")
		return
	}

	summary, err := h.headcountService.GetHeadcountByDate(date)
	if err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, summary, "Headcount retrieved successfully")
}

// GetDetailedHeadcount returns detailed headcount for a specific date and meal
// @Summary Get detailed headcount
// @Description Get detailed headcount breakdown for a specific date and meal type (Admin/Logistics only)
// @Tags headcount
// @Produce json
// @Security BearerAuth
// @Param date path string true "Date (YYYY-MM-DD)"
// @Param meal_type path string true "Meal type (breakfast/lunch/dinner)"
// @Success 200 {object} map[string]interface{} "Detailed headcount retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /headcount/{date}/{meal_type} [get]
func (h *HeadcountHandler) GetDetailedHeadcount(c *gin.Context) {
	// Get parameters from URL
	date := c.Param("date")
	mealType := c.Param("meal_type")

	if date == "" || mealType == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Date and meal_type parameters are required")
		return
	}

	details, err := h.headcountService.GetDetailedHeadcount(date, mealType)
	if err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, details, "Detailed headcount retrieved successfully")
}

// GetAnnouncement generates announcement message for a specific date
// @Summary Get announcement
// @Description Generate meal announcement message for a specific date (Admin/Logistics only)
// @Tags headcount
// @Produce json
// @Security BearerAuth
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Announcement generated successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /headcount/{date}/announcement [get]
func (h *HeadcountHandler) GetAnnouncement(c *gin.Context) {
    date := c.Param("date")
    if date == "" {
        utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Date parameter is required")
        return
    }

    message, err := h.headcountService.GenerateAnnouncement(date)
    if err != nil {
        utils.ErrorResponse(c, 400, "ANNOUNCEMENT_ERROR", err.Error())
        return
    }

    utils.SuccessResponse(c, 200, gin.H{
        "date":    date,
        "message": message,
    }, "Announcement generated")
}

// StreamHeadcount streams real-time headcount updates via SSE
// @Summary Stream headcount updates
// @Description Stream real-time headcount updates for a specific date via Server-Sent Events (Admin/Logistics only)
// @Tags headcount
// @Produce text/event-stream
// @Security BearerAuth
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {string} string "SSE stream"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /headcount/{date}/stream [get]
func (h *HeadcountHandler) StreamHeadcount(c *gin.Context) {
	date := c.Param("date")
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Date parameter is required")
		return
	}

	summary, err := h.headcountService.GetHeadcountByDate(date)
	if err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	initial, _ := json.Marshal(summary)

	ch := h.hub.Register(date)
	defer h.hub.Deregister(date, ch)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	c.Stream(func(w io.Writer) bool {
		fmt.Fprintf(w, "data: %s\n\n", initial)
		c.Writer.Flush()

		for payload := range ch {
			fmt.Fprintf(w, "data: %s\n\n", payload)
			c.Writer.Flush()
		}
		return false
	})
}

// GetForecast returns headcount forecast for upcoming days
// @Summary Get headcount forecast
// @Description Get meal headcount forecast for upcoming days (Admin/Logistics only)
// @Tags headcount
// @Produce json
// @Security BearerAuth
// @Param days query int false "Number of days to forecast (default: 7)"
// @Success 200 {object} map[string]interface{} "Forecast retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /headcount/forecast [get]
func (h *HeadcountHandler) GetForecast(c *gin.Context) {
    days := 7
    if daysStr := c.Query("days"); daysStr != "" {
        if parsed, err := strconv.Atoi(daysStr); err == nil {
            days = parsed
        }
    }

    summaries, err := h.headcountService.GetForecast(days)
    if err != nil {
        utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
        return
    }

    utils.SuccessResponse(c, 200, summaries, "Forecast retrieved successfully")
}


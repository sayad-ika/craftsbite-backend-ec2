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
// GET /api/headcount/today
func (h *HeadcountHandler) GetTodayHeadcount(c *gin.Context) {
	summary, err := h.headcountService.GetTodayHeadcount()
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, summary, "Today's and tomorrow's headcount retrieved successfully")
}

// GetHeadcountByDate returns headcount summary for a specific date
// GET /api/headcount/:date
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
// GET /api/headcount/:date/:meal_type
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


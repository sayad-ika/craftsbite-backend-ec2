package handlers

import (
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type WFHPeriodHandler struct {
	svc services.WFHPeriodService
}

func NewWFHPeriodHandler(svc services.WFHPeriodService) *WFHPeriodHandler {
	return &WFHPeriodHandler{svc: svc}
}

type createWFHPeriodRequest struct {
	StartDate string  `json:"start_date" binding:"required"`
	EndDate   string  `json:"end_date" binding:"required"`
	Reason    *string `json:"reason"`
}

// POST /api/v1/wfh-periods
func (h *WFHPeriodHandler) CreateWFHPeriod(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	var req createWFHPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	resp, err := h.svc.CreatePeriod(adminID.(string), req.StartDate, req.EndDate, req.Reason)
	if err != nil {
		utils.ErrorResponse(c, 400, "CREATE_WFH_PERIOD_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 201, resp, "WFH period created successfully")
}

// GET /api/v1/wfh-periods
func (h *WFHPeriodHandler) ListWFHPeriods(c *gin.Context) {
	result, err := h.svc.ListPeriods()
	if err != nil {
		utils.ErrorResponse(c, 500, "LIST_WFH_PERIODS_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, result, "WFH periods retrieved")
}

// DELETE /api/v1/wfh-periods/:id
func (h *WFHPeriodHandler) DeleteWFHPeriod(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Period ID is required")
		return
	}

	if err := h.svc.DeletePeriod(id); err != nil {
		utils.ErrorResponse(c, 400, "DELETE_WFH_PERIOD_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "WFH period deleted successfully")
}

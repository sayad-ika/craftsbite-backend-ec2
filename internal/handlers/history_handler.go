package handlers

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HistoryHandler handles meal participation history endpoints
type HistoryHandler struct {
	historyService services.HistoryService
}

// NewHistoryHandler creates a new history handler
func NewHistoryHandler(historyService services.HistoryService) *HistoryHandler {
	return &HistoryHandler{
		historyService: historyService,
	}
}

// GetHistory returns the participation history for the current user
// GET /api/v1/meals/history
func (h *HistoryHandler) GetHistory(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse query filters
	filters := parseHistoryFilters(c)

	history, err := h.historyService.GetUserHistory(userID.(string), filters)
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, history, "History retrieved successfully")
}

// GetAuditTrail returns the audit trail for the current user
// GET /api/v1/meals/participation-audit
func (h *HistoryHandler) GetAuditTrail(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse query filters
	filters := parseHistoryFilters(c)

	history, err := h.historyService.GetAuditTrail(userID.(string), filters)
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, history, "Audit trail retrieved successfully")
}

// GetUserHistoryAdmin returns the participation history for a specific user (admin/logistics only)
// GET /api/v1/admin/meals/history/:user_id
func (h *HistoryHandler) GetUserHistoryAdmin(c *gin.Context) {
	// Get requester role from context
	role, exists := c.Get("role")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Verify requester has permission (admin or logistics)
	userRole := models.Role(role.(string))
	if userRole != models.RoleAdmin && userRole != models.RoleLogistics {
		utils.ErrorResponse(c, 403, "FORBIDDEN", "Only admins and logistics can view other users' history")
		return
	}

	// Get target user ID from URL parameter
	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "User ID is required")
		return
	}

	// Parse query filters
	filters := parseHistoryFilters(c)

	history, err := h.historyService.GetUserHistory(targetUserID, filters)
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, history, "User history retrieved successfully")
}

// parseHistoryFilters extracts history filters from query parameters
func parseHistoryFilters(c *gin.Context) services.HistoryFilters {
	filters := services.HistoryFilters{
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		MealType:  c.Query("meal_type"),
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	return filters
}

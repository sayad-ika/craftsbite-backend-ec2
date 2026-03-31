package handlers

import (
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// BulkOptOutHandler handles bulk opt-out endpoints
type BulkOptOutHandler struct {
	bulkOptOutService services.BulkOptOutService
}

// NewBulkOptOutHandler creates a new bulk opt-out handler
func NewBulkOptOutHandler(bulkOptOutService services.BulkOptOutService) *BulkOptOutHandler {
	return &BulkOptOutHandler{
		bulkOptOutService: bulkOptOutService,
	}
}

// GetBulkOptOuts returns all bulk opt-outs for the current user
// @Summary Get bulk opt-outs
// @Description Get all bulk opt-outs for current user
// @Tags meals
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Bulk opt-outs retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Router /meals/bulk-optouts [get]
func (h *BulkOptOutHandler) GetBulkOptOuts(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	optOuts, err := h.bulkOptOutService.GetBulkOptOuts(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, optOuts, "Bulk opt-outs retrieved successfully")
}

// CreateBulkOptOutRequest represents the request body for creating a bulk opt-out
type CreateBulkOptOutRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	MealType  string `json:"meal_type" binding:"required"`
}

// CreateBulkOptOut creates a new bulk opt-out for the current user
// @Summary Create bulk opt-out
// @Description Create a new bulk opt-out period for current user
// @Tags meals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateBulkOptOutRequest true "Bulk opt-out details"
// @Success 201 {object} map[string]interface{} "Bulk opt-out created successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /meals/bulk-optouts [post]
func (h *BulkOptOutHandler) CreateBulkOptOut(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse request body
	var req CreateBulkOptOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	// Create bulk opt-out
	input := services.CreateBulkOptOutInput{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		MealType:  req.MealType,
	}

	optOut, err := h.bulkOptOutService.CreateBulkOptOut(userID.(string), input)
	if err != nil {
		utils.ErrorResponse(c, 400, "CREATE_BULK_OPTOUT_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 201, optOut, "Bulk opt-out created successfully")
}

// DeleteBulkOptOut deletes a bulk opt-out
// @Summary Delete bulk opt-out
// @Description Delete a bulk opt-out period
// @Tags meals
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bulk opt-out ID"
// @Success 200 {object} map[string]interface{} "Bulk opt-out deleted successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /meals/bulk-optouts/{id} [delete]
func (h *BulkOptOutHandler) DeleteBulkOptOut(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Get bulk opt-out ID from URL parameter
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Bulk opt-out ID is required")
		return
	}

	// Delete bulk opt-out
	err := h.bulkOptOutService.DeleteBulkOptOut(userID.(string), id)
	if err != nil {
		utils.ErrorResponse(c, 400, "DELETE_BULK_OPTOUT_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "Bulk opt-out deleted successfully")
}

type AdminBulkOptOutRequest struct {
	UserIDs   []string `json:"user_ids"   binding:"required,min=1"`
	StartDate string   `json:"start_date" binding:"required"`
	EndDate   string   `json:"end_date"   binding:"required"`
	MealTypes []string `json:"meal_types" binding:"required,min=1"`
	Reason    string   `json:"reason"     binding:"required"`
}

// AdminBulkOptOut creates bulk opt-outs for multiple users
// @Summary Admin bulk opt-out
// @Description Create bulk opt-outs for multiple users (Admin/Team Lead only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AdminBulkOptOutRequest true "Admin bulk opt-out details"
// @Success 200 {object} map[string]interface{} "Admin bulk opt-out processed"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/meals/bulk-optouts [post]
func (h *BulkOptOutHandler) AdminBulkOptOut(c *gin.Context) {
	actorID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}
	actorRole, exists := c.Get("role")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User role not found in context")
		return
	}

	var req AdminBulkOptOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	input := services.AdminBulkOptOutInput{
		UserIDs:   req.UserIDs,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		MealTypes: req.MealTypes,
		Reason:    req.Reason,
	}

	result, err := h.bulkOptOutService.AdminBulkOptOut(actorID.(string), actorRole.(string), input)
	if err != nil {
		utils.ErrorResponse(c, 400, "ADMIN_BULK_OPTOUT_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, result, "Admin bulk opt-out processed")
}

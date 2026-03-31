package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

// HistoryFilters represents filters for querying history
type HistoryFilters struct {
	StartDate string
	EndDate   string
	MealType  string
	Limit     int
}

// HistoryService defines the interface for participation history operations
type HistoryService interface {
	GetUserHistory(userID string, filters HistoryFilters) ([]models.MealParticipationHistory, error)
	GetAuditTrail(userID string, filters HistoryFilters) ([]models.MealParticipationHistory, error)
	RecordChange(userID, date, mealType string, action models.HistoryAction, previousValue *string, changedByUserID *string, reason *string, ipAddress *string) error
}

// historyService implements HistoryService
type historyService struct {
	historyRepo repository.HistoryRepository
}

// NewHistoryService creates a new history service
func NewHistoryService(historyRepo repository.HistoryRepository) HistoryService {
	return &historyService{
		historyRepo: historyRepo,
	}
}

// GetUserHistory retrieves participation history for a user
func (s *historyService) GetUserHistory(userID string, filters HistoryFilters) ([]models.MealParticipationHistory, error) {
	// If date range is provided, use that
	if filters.StartDate != "" && filters.EndDate != "" {
		history, err := s.historyRepo.FindByUserAndDateRange(userID, filters.StartDate, filters.EndDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get user history: %w", err)
		}
		return history, nil
	}

	// Otherwise, get with limit
	limit := filters.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}

	history, err := s.historyRepo.FindByUser(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user history: %w", err)
	}

	return history, nil
}

// GetAuditTrail retrieves the audit trail for a user (includes changes made by others)
func (s *historyService) GetAuditTrail(userID string, filters HistoryFilters) ([]models.MealParticipationHistory, error) {
	// For the audit trail, we use the same query but include changes made by others
	// The repository already includes all changes for the user regardless of who made them
	return s.GetUserHistory(userID, filters)
}

// RecordChange records a participation change in history
func (s *historyService) RecordChange(userID, date, mealType string, action models.HistoryAction, previousValue *string, changedByUserID *string, reason *string, ipAddress *string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	var changedByUUID *uuid.UUID
	if changedByUserID != nil {
		parsed, err := uuid.Parse(*changedByUserID)
		if err != nil {
			return fmt.Errorf("invalid changed_by user ID: %w", err)
		}
		changedByUUID = &parsed
	}

	history := &models.MealParticipationHistory{
		UserID:          userUUID,
		Date:            date,
		MealType:        models.MealType(mealType),
		Action:          action,
		PreviousValue:   previousValue,
		ChangedByUserID: changedByUUID,
		Reason:          reason,
		IPAddress:       ipAddress,
	}

	if err := s.historyRepo.Create(history); err != nil {
		return fmt.Errorf("failed to record change: %w", err)
	}

	return nil
}
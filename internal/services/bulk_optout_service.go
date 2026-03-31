package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminBulkOptOutInput struct {
	UserIDs   []string
	StartDate string
	EndDate   string
	MealTypes []string
	Reason    string
}

type AdminBulkOptOutFailure struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

type AdminBulkOptOutResult struct {
	Succeeded []string                 `json:"succeeded"`
	Failed    []AdminBulkOptOutFailure `json:"failed"`
}

// CreateBulkOptOutInput represents input for creating a bulk opt-out
type CreateBulkOptOutInput struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	MealType  string `json:"meal_type" binding:"required"`
}

// BulkOptOutService defines the interface for bulk opt-out management
type BulkOptOutService interface {
	GetBulkOptOuts(userID string) ([]models.BulkOptOut, error)
	CreateBulkOptOut(userID string, input CreateBulkOptOutInput) (*models.BulkOptOut, error)
	DeleteBulkOptOut(userID, id string) error
	AdminBulkOptOut(actorID, actorRole string, input AdminBulkOptOutInput) (*AdminBulkOptOutResult, error)
}

// bulkOptOutService implements BulkOptOutService
type bulkOptOutService struct {
	db             *gorm.DB
	bulkOptOutRepo repository.BulkOptOutRepository
	historyRepo    repository.HistoryRepository
	teamRepo       repository.TeamRepository
}

// NewBulkOptOutService creates a new bulk opt-out service
func NewBulkOptOutService(db *gorm.DB, bulkOptOutRepo repository.BulkOptOutRepository, historyRepo repository.HistoryRepository, teamRepo repository.TeamRepository) BulkOptOutService {
	return &bulkOptOutService{
		db:             db,
		bulkOptOutRepo: bulkOptOutRepo,
		historyRepo:    historyRepo,
		teamRepo:       teamRepo,
	}
}

// GetBulkOptOuts retrieves all bulk opt-outs for a user
func (s *bulkOptOutService) GetBulkOptOuts(userID string) ([]models.BulkOptOut, error) {
	optOuts, err := s.bulkOptOutRepo.FindByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk opt-outs: %w", err)
	}
	return optOuts, nil
}

// CreateBulkOptOut creates a new bulk opt-out for a user
func (s *bulkOptOutService) CreateBulkOptOut(userID string, input CreateBulkOptOutInput) (*models.BulkOptOut, error) {
	// Validate date format
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: must be YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: must be YYYY-MM-DD")
	}

	// Validate end_date >= start_date
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end_date must be on or after start_date")
	}

	// Validate meal type
	mealType := models.MealType(input.MealType)
	if !mealType.IsValid() {
		return nil, fmt.Errorf("invalid meal_type: must be one of lunch, snacks, iftar, event_dinner, optional_dinner")
	}

	// Parse user UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	// Create bulk opt-out
	bulkOptOut := &models.BulkOptOut{
		UserID:    userUUID,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
		MealType:  mealType,
		IsActive:  true,
	}

	if err := s.bulkOptOutRepo.Create(bulkOptOut); err != nil {
		return nil, fmt.Errorf("failed to create bulk opt-out: %w", err)
	}

	// Record in history
	historyRecord := &models.MealParticipationHistory{
		UserID:          userUUID,
		Date:            input.StartDate, // Use start date as reference
		MealType:        mealType,
		Action:          models.HistoryActionOptedOut,
		ChangedByUserID: &userUUID,
		Reason:          strPtr(fmt.Sprintf("Bulk opt-out from %s to %s", input.StartDate, input.EndDate)),
	}

	if err := s.historyRepo.Create(historyRecord); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to record bulk opt-out in history: %v\n", err)
	}

	return bulkOptOut, nil
}

// DeleteBulkOptOut deletes a bulk opt-out if it belongs to the user
func (s *bulkOptOutService) DeleteBulkOptOut(userID, id string) error {
	// Get all user's bulk opt-outs to verify ownership
	optOuts, err := s.bulkOptOutRepo.FindByUser(userID)
	if err != nil {
		return fmt.Errorf("failed to verify bulk opt-out ownership: %w", err)
	}

	// Check if the opt-out belongs to the user
	found := false
	for _, optOut := range optOuts {
		if optOut.ID.String() == id {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("bulk opt-out not found or does not belong to user")
	}

	// Delete the opt-out
	if err := s.bulkOptOutRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete bulk opt-out: %w", err)
	}

	return nil
}

// strPtr returns a pointer to a string
func strPtr(s string) *string {
	return &s
}

func (s *bulkOptOutService) AdminBulkOptOut(actorID, actorRole string, input AdminBulkOptOutInput) (*AdminBulkOptOutResult, error) {
	if validateDate(input.StartDate) != nil || validateDate(input.EndDate) != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}
	startDate, _ := time.Parse("2006-01-02", input.StartDate)
	endDate, _ := time.Parse("2006-01-02", input.EndDate)

	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end_date must be on or after start_date")
	}
	if len(input.MealTypes) == 0 {
		return nil, fmt.Errorf("meal_types must not be empty")
	}

	var mealTypes []models.MealType
	for _, mt := range input.MealTypes {
		mealType := models.MealType(mt)
		if !mealType.IsValid() {
			return nil, fmt.Errorf("invalid meal_type '%s': must be one of lunch, snacks, iftar, event_dinner, optional_dinner", mt)
		}
		mealTypes = append(mealTypes, mealType)
	}

	if len(input.UserIDs) == 0 {
		return nil, fmt.Errorf("user_ids must not be empty")
	}

	actorUUID, err := uuid.Parse(actorID)
	if err != nil {
		return nil, fmt.Errorf("invalid actor ID format")
	}

	result, err := s.validateUsers(actorID, actorRole, input.UserIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to validate users: %w", err)
	}

	if len(result.Failed) > 0 {
		return result, nil
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	reason := input.Reason
	if reason == "" {
		reason = fmt.Sprintf("Admin bulk opt-out from %s to %s", input.StartDate, input.EndDate)
	}

	for _, userID := range input.UserIDs {
		userUUID, _ := uuid.Parse(userID)

		for _, mealType := range mealTypes {
			var previousValue *string
			existing, err := s.bulkOptOutRepo.FindActiveByUserAndMealType(userUUID, mealType, input.StartDate)
			if err == nil && existing != nil {
				v := string(existing.MealType)
				previousValue = &v
			}

			bulkOptOut := &models.BulkOptOut{
				UserID:         userUUID,
				StartDate:      input.StartDate,
				EndDate:        input.EndDate,
				MealType:       mealType,
				IsActive:       true,
				OverrideBy:     &actorUUID,
				OverrideReason: reason,
			}
			if err := tx.Create(bulkOptOut).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create bulk opt-out for user %s meal %s: %w", userID, mealType, err)
			}

			historyRecord := &models.MealParticipationHistory{
				UserID:          userUUID,
				Date:            input.StartDate,
				MealType:        mealType,
				Action:          models.HistoryActionOverrideOut,
				ChangedByUserID: &actorUUID,
				Reason:          strPtr(reason),
				PreviousValue:   previousValue,
			}
			if err := tx.Create(historyRecord).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create history for user %s meal %s: %w", userID, mealType, err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return result, nil
}

func (s *bulkOptOutService) validateUsers(actorID, actorRole string, userIDs []string) (*AdminBulkOptOutResult, error) {
	result := &AdminBulkOptOutResult{
		Succeeded: []string{},
		Failed:    []AdminBulkOptOutFailure{},
	}

	for _, userID := range userIDs {
		_, err := uuid.Parse(userID)
		if err != nil {
			result.Failed = append(result.Failed, AdminBulkOptOutFailure{UserID: userID, Reason: "invalid user ID format"})
			continue
		}

		if actorRole == string(models.RoleTeamLead) {
			isMember, err := s.teamRepo.IsUserInAnyTeamLedBy(actorID, userID)
			if err != nil {
				result.Failed = append(result.Failed, AdminBulkOptOutFailure{UserID: userID, Reason: "failed to verify team membership"})
				continue
			}
			if !isMember {
				result.Failed = append(result.Failed, AdminBulkOptOutFailure{UserID: userID, Reason: "user is not a member of your team"})
				continue
			}
		}

		result.Succeeded = append(result.Succeeded, userID)
	}

	return result, nil
}

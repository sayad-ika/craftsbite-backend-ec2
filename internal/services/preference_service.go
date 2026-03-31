package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

// UserPreferences represents user meal preferences
type UserPreferences struct {
	UserID                string `json:"user_id"`
	DefaultMealPreference string `json:"default_meal_preference"`
}

// PreferenceService defines the interface for user preference management
type PreferenceService interface {
	GetPreferences(userID string) (*UserPreferences, error)
	UpdateDefaultPreference(userID string, preference string) error
}

// preferenceService implements PreferenceService
type preferenceService struct {
	userRepo    repository.UserRepository
	historyRepo repository.HistoryRepository
}

// NewPreferenceService creates a new preference service
func NewPreferenceService(userRepo repository.UserRepository, historyRepo repository.HistoryRepository) PreferenceService {
	return &preferenceService{
		userRepo:    userRepo,
		historyRepo: historyRepo,
	}
}

// GetPreferences retrieves a user's meal preferences
func (s *preferenceService) GetPreferences(userID string) (*UserPreferences, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	return &UserPreferences{
		UserID:                user.ID.String(),
		DefaultMealPreference: user.DefaultMealPreference,
	}, nil
}

// UpdateDefaultPreference updates a user's default meal preference
func (s *preferenceService) UpdateDefaultPreference(userID string, preference string) error {
	// Validate preference
	if preference != "opt_in" && preference != "opt_out" {
		return fmt.Errorf("invalid preference: must be 'opt_in' or 'opt_out'")
	}

	// Get current user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Store previous value for history
	previousValue := user.DefaultMealPreference

	// Skip if no change
	if previousValue == preference {
		return nil
	}

	// Update user preference
	user.DefaultMealPreference = preference
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update preference: %w", err)
	}

	// Record change in history
	userUUID, _ := uuid.Parse(userID)
	historyRecord := &models.MealParticipationHistory{
		UserID:          userUUID,
		Date:            "", // Empty - preference change is not date-specific
		MealType:        "", // Empty - applies to all meals
		Action:          models.HistoryAction(fmt.Sprintf("preference_changed_to_%s", preference)),
		PreviousValue:   &previousValue,
		ChangedByUserID: &userUUID, // Self-change
	}

	if err := s.historyRepo.Create(historyRecord); err != nil {
		// Log but don't fail - the preference update succeeded
		fmt.Printf("Warning: failed to record preference change in history: %v\n", err)
	}

	return nil
}

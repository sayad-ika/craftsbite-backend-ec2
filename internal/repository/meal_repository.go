package repository

import (
	"craftsbite-backend/internal/models"
	"fmt"

	"gorm.io/gorm"
)

// MealRepository defines the interface for meal participation data access
type MealRepository interface {
	CreateOrUpdate(participation *models.MealParticipation) error
	FindByUserAndDate(userID, date string) ([]models.MealParticipation, error)
	FindByUserDateMeal(userID, date, mealType string) (*models.MealParticipation, error)
	FindByDate(date string) ([]models.MealParticipation, error)
	FindByDateAndMeal(date, mealType string) ([]models.MealParticipation, error)
}

// mealRepository implements MealRepository
type mealRepository struct {
	db *gorm.DB
}

// NewMealRepository creates a new meal repository
func NewMealRepository(db *gorm.DB) MealRepository {
	return &mealRepository{db: db}
}

// CreateOrUpdate creates or updates a meal participation (upsert)
func (r *mealRepository) CreateOrUpdate(participation *models.MealParticipation) error {
	// Check if this is a new record (ID will be set if existing record)
	var existing models.MealParticipation
	err := r.db.Where("id = ?", participation.ID).First(&existing).Error

	switch err {
	case gorm.ErrRecordNotFound:
		// New record - use map to explicitly set all values including false booleans
		err = r.db.Table("meal_participations").Create(map[string]interface{}{
			"id":               participation.ID,
			"user_id":          participation.UserID,
			"date":             participation.Date,
			"meal_type":        participation.MealType,
			"is_participating": participation.IsParticipating,
			"opted_out_at":     participation.OptedOutAt,
			"override_by":      participation.OverrideBy,
			"override_reason":  participation.OverrideReason,
		}).Error
	case nil:
		// Existing record - use map to explicitly update all fields
		err = r.db.Table("meal_participations").Where("id = ?", participation.ID).Updates(map[string]interface{}{
			"is_participating": participation.IsParticipating,
			"opted_out_at":     participation.OptedOutAt,
			"override_by":      participation.OverrideBy,
			"override_reason":  participation.OverrideReason,
		}).Error
	}

	if err != nil {
		return fmt.Errorf("failed to create or update meal participation: %w", err)
	}
	return nil
}

// FindByUserAndDate finds all meal participations for a user on a specific date
func (r *mealRepository) FindByUserAndDate(userID, date string) ([]models.MealParticipation, error) {
	var participations []models.MealParticipation
	err := r.db.Where("user_id = ? AND date = ?", userID, date).Find(&participations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find meal participations: %w", err)
	}
	return participations, nil
}

// FindByUserDateMeal finds a specific meal participation
func (r *mealRepository) FindByUserDateMeal(userID, date, mealType string) (*models.MealParticipation, error) {
	var participation models.MealParticipation
	err := r.db.Where("user_id = ? AND date = ? AND meal_type = ?", userID, date, mealType).First(&participation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error, return nil
		}
		return nil, fmt.Errorf("failed to find meal participation: %w", err)
	}
	return &participation, nil
}

// FindByDate finds all meal participations for a specific date
func (r *mealRepository) FindByDate(date string) ([]models.MealParticipation, error) {
	var participations []models.MealParticipation
	err := r.db.Where("date = ?", date).Find(&participations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find meal participations by date: %w", err)
	}
	return participations, nil
}

// FindByDateAndMeal finds all participations for a specific date and meal type
func (r *mealRepository) FindByDateAndMeal(date, mealType string) ([]models.MealParticipation, error) {
	var participations []models.MealParticipation
	err := r.db.Where("date = ? AND meal_type = ?", date, mealType).Find(&participations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find meal participations by date and meal: %w", err)
	}
	return participations, nil
}

package repository

import (
	"craftsbite-backend/internal/models"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BulkOptOutRepository defines the interface for bulk opt-out data access
type BulkOptOutRepository interface {
	Create(bulkOptOut *models.BulkOptOut) error
	FindByUser(userID string) ([]models.BulkOptOut, error)
	FindActiveByUserAndDate(userID, date string) ([]models.BulkOptOut, error)
	Delete(id string) error
	Deactivate(id string) error
	FindActiveByUserAndMealType (userID uuid.UUID, mealType models.MealType, date string) (*models.BulkOptOut, error)
}

// bulkOptOutRepository implements BulkOptOutRepository
type bulkOptOutRepository struct {
	db *gorm.DB
}

// NewBulkOptOutRepository creates a new bulk opt-out repository
func NewBulkOptOutRepository(db *gorm.DB) BulkOptOutRepository {
	return &bulkOptOutRepository{db: db}
}

// Create creates a new bulk opt-out
func (r *bulkOptOutRepository) Create(bulkOptOut *models.BulkOptOut) error {
	if err := r.db.Create(bulkOptOut).Error; err != nil {
		return fmt.Errorf("failed to create bulk opt-out: %w", err)
	}
	return nil
}

// FindByUser finds all bulk opt-outs for a user
func (r *bulkOptOutRepository) FindByUser(userID string) ([]models.BulkOptOut, error) {
	var bulkOptOuts []models.BulkOptOut
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&bulkOptOuts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find bulk opt-outs: %w", err)
	}
	return bulkOptOuts, nil
}

// FindActiveByUserAndDate finds active bulk opt-outs for a user that cover a specific date
func (r *bulkOptOutRepository) FindActiveByUserAndDate(userID, date string) ([]models.BulkOptOut, error) {
	var bulkOptOuts []models.BulkOptOut
	err := r.db.Where("user_id = ? AND is_active = ? AND ? BETWEEN start_date AND end_date",
		userID, true, date).
		Find(&bulkOptOuts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find active bulk opt-outs: %w", err)
	}
	return bulkOptOuts, nil
}

// Delete deletes a bulk opt-out by ID
func (r *bulkOptOutRepository) Delete(id string) error {
	if err := r.db.Delete(&models.BulkOptOut{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete bulk opt-out: %w", err)
	}
	return nil
}

// Deactivate deactivates a bulk opt-out by setting is_active to false
func (r *bulkOptOutRepository) Deactivate(id string) error {
	if err := r.db.Model(&models.BulkOptOut{}).
		Where("id = ?", id).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate bulk opt-out: %w", err)
	}
	return nil
}

func (r *bulkOptOutRepository) FindActiveByUserAndMealType(userID uuid.UUID, mealType models.MealType, date string) (*models.BulkOptOut, error) {
    var bulkOptOut models.BulkOptOut
    err := r.db.Where(
        "user_id = ? AND meal_type = ? AND is_active = ? AND ? BETWEEN start_date AND end_date",
        userID, mealType, true, date,
    ).First(&bulkOptOut).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to find active bulk opt-out: %w", err)
    }
    return &bulkOptOut, nil
}

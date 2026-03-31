package repository

import (
	"craftsbite-backend/internal/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// HistoryRepository defines the interface for meal participation history data access
type HistoryRepository interface {
	Create(history *models.MealParticipationHistory) error
	FindByUser(userID string, limit int) ([]models.MealParticipationHistory, error)
	FindByUserAndDateRange(userID, startDate, endDate string) ([]models.MealParticipationHistory, error)
	DeleteOlderThan(months int) (int64, error)
	FindAll(limit int) ([]models.MealParticipationHistory, error)
}

// historyRepository implements HistoryRepository
type historyRepository struct {
	db *gorm.DB
}

// NewHistoryRepository creates a new history repository
func NewHistoryRepository(db *gorm.DB) HistoryRepository {
	return &historyRepository{db: db}
}

// Create creates a new history record
func (r *historyRepository) Create(history *models.MealParticipationHistory) error {
	if err := r.db.Create(history).Error; err != nil {
		return fmt.Errorf("failed to create history record: %w", err)
	}
	return nil
}

// FindByUser finds history records for a user, ordered by created_at DESC
func (r *historyRepository) FindByUser(userID string, limit int) ([]models.MealParticipationHistory, error) {
	var history []models.MealParticipationHistory
	query := r.db.Where("user_id = ?", userID).
		Preload("ChangedBy").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&history).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find history records: %w", err)
	}
	return history, nil
}

// FindByUserAndDateRange finds history records for a user within a date range
func (r *historyRepository) FindByUserAndDateRange(userID, startDate, endDate string) ([]models.MealParticipationHistory, error) {
	var history []models.MealParticipationHistory
	err := r.db.Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Preload("ChangedBy").
		Order("created_at DESC").
		Find(&history).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find history records by date range: %w", err)
	}
	return history, nil
}

// DeleteOlderThan deletes history records older than the specified number of months
func (r *historyRepository) DeleteOlderThan(months int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, -months, 0)

	result := r.db.Where("created_at < ?", cutoffDate).
		Delete(&models.MealParticipationHistory{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete old history records: %w", result.Error)
	}

	return result.RowsAffected, nil
}

func (r *historyRepository) FindAll(limit int) ([]models.MealParticipationHistory, error) {
    var history []models.MealParticipationHistory
    query := r.db.Preload("ChangedBy").
        Order("created_at DESC")

    if limit > 0 {
        query = query.Limit(limit)
    }

    err := query.Find(&history).Error
    if err != nil {
        return nil, fmt.Errorf("failed to find all history records: %w", err)
    }
    return history, nil
}
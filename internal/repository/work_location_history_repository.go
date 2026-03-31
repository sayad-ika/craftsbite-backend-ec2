package repository

import (
	"craftsbite-backend/internal/models"
	"fmt"

	"gorm.io/gorm"
)

type WorkLocationHistoryRepository interface {
	Create(history *models.WorkLocationHistory) error
	FindByUserAndDate(userID, date string) ([]models.WorkLocationHistory, error)
}

type workLocationHistoryRepository struct {
	db *gorm.DB
}

func NewWorkLocationHistoryRepository(db *gorm.DB) WorkLocationHistoryRepository {
	return &workLocationHistoryRepository{db: db}
}

func (r *workLocationHistoryRepository) Create(history *models.WorkLocationHistory) error {
	if err := r.db.Create(history).Error; err != nil {
		return fmt.Errorf("failed to create work location history record: %w", err)
	}
	return nil
}

func (r *workLocationHistoryRepository) FindByUserAndDate(userID, date string) ([]models.WorkLocationHistory, error) {
	var history []models.WorkLocationHistory
	err := r.db.Where("user_id = ? AND date = ?", userID, date).
		Preload("OverrideBy").
		Order("created_at DESC").
		Find(&history).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find work location history: %w", err)
	}
	return history, nil
}

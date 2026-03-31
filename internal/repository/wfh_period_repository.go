package repository

import (
	"craftsbite-backend/internal/models"
	"fmt"

	"gorm.io/gorm"
)

// WFHPeriodRepository defines data-access for company-wide WFH periods
type WFHPeriodRepository interface {
	Create(period *models.WFHPeriod) error
	FindByID(id string) (*models.WFHPeriod, error)
	FindActiveByDate(date string) (*models.WFHPeriod, error)
	FindAll() ([]models.WFHPeriod, error)
	Delete(id string) error
}

type wfhPeriodRepository struct {
	db *gorm.DB
}

func NewWFHPeriodRepository(db *gorm.DB) WFHPeriodRepository {
	return &wfhPeriodRepository{db: db}
}

// Create inserts a new WFH period
func (r *wfhPeriodRepository) Create(period *models.WFHPeriod) error {
	if err := r.db.Create(period).Error; err != nil {
		return fmt.Errorf("failed to create WFH period: %w", err)
	}
	return nil
}

// FindByID returns a single WFH period by its UUID
func (r *wfhPeriodRepository) FindByID(id string) (*models.WFHPeriod, error) {
	var period models.WFHPeriod
	err := r.db.Preload("Creator").Where("id = ?", id).First(&period).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find WFH period: %w", err)
	}
	return &period, nil
}

// FindActiveByDate returns the first active WFH period covering the given date, or nil
func (r *wfhPeriodRepository) FindActiveByDate(date string) (*models.WFHPeriod, error) {
	var period models.WFHPeriod
	err := r.db.Where("active = ? AND start_date <= ? AND end_date >= ?", true, date, date).
		First(&period).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find active WFH period for date: %w", err)
	}
	return &period, nil
}

// FindAll returns all WFH periods ordered by start_date DESC
func (r *wfhPeriodRepository) FindAll() ([]models.WFHPeriod, error) {
	var periods []models.WFHPeriod
	err := r.db.Preload("Creator").Order("start_date DESC").Find(&periods).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list WFH periods: %w", err)
	}
	return periods, nil
}

// Delete hard-deletes a WFH period by ID
func (r *wfhPeriodRepository) Delete(id string) error {
	result := r.db.Delete(&models.WFHPeriod{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete WFH period: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("WFH period not found")
	}
	return nil
}

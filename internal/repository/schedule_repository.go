package repository

import (
	"craftsbite-backend/internal/models"
	"fmt"

	"gorm.io/gorm"
)

// ScheduleRepository defines the interface for day schedule data access
type ScheduleRepository interface {
	Create(schedule *models.DaySchedule) error
	FindByDate(date string) (*models.DaySchedule, error)
	FindByDateRange(startDate, endDate string) ([]models.DaySchedule, error)
	Update(schedule *models.DaySchedule) error
	Delete(id string) error
}

// scheduleRepository implements ScheduleRepository
type scheduleRepository struct {
	db *gorm.DB
}

// NewScheduleRepository creates a new schedule repository
func NewScheduleRepository(db *gorm.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

// Create creates a new day schedule
func (r *scheduleRepository) Create(schedule *models.DaySchedule) error {
	if err := r.db.Create(schedule).Error; err != nil {
		return fmt.Errorf("failed to create day schedule: %w", err)
	}
	return nil
}

// FindByDate finds a day schedule by date
func (r *scheduleRepository) FindByDate(date string) (*models.DaySchedule, error) {
	var schedule models.DaySchedule
	err := r.db.Where("date = ?", date).First(&schedule).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to find day schedule: %w", err)
	}
	return &schedule, nil
}

// FindByDateRange finds all day schedules within a date range
func (r *scheduleRepository) FindByDateRange(startDate, endDate string) ([]models.DaySchedule, error) {
	var schedules []models.DaySchedule
	err := r.db.Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date ASC").
		Find(&schedules).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find day schedules by date range: %w", err)
	}
	return schedules, nil
}

// Update updates a day schedule
func (r *scheduleRepository) Update(schedule *models.DaySchedule) error {
	if err := r.db.Save(schedule).Error; err != nil {
		return fmt.Errorf("failed to update day schedule: %w", err)
	}
	return nil
}

// Delete deletes a day schedule by ID
func (r *scheduleRepository) Delete(id string) error {
	if err := r.db.Delete(&models.DaySchedule{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete day schedule: %w", err)
	}
	return nil
}

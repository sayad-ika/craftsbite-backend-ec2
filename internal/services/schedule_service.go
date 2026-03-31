package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ScheduleService defines the interface for day schedule business logic
type ScheduleService interface {
	GetSchedule(date string) (*models.DaySchedule, error)
	GetScheduleRange(startDate, endDate string) ([]models.DaySchedule, error)
	CreateSchedule(adminID string, input CreateScheduleInput) (*models.DaySchedule, error)
	UpdateSchedule(id string, input UpdateScheduleInput) (*models.DaySchedule, error)
	DeleteSchedule(id string) error
}

// CreateScheduleInput represents input for creating a day schedule
type CreateScheduleInput struct {
	Date           string            `json:"date" binding:"required"`
	DayStatus      models.DayStatus  `json:"day_status" binding:"required"`
	Reason         string            `json:"reason"`
	AvailableMeals []models.MealType `json:"available_meals"`
}

// UpdateScheduleInput represents input for updating a day schedule
type UpdateScheduleInput struct {
	DayStatus      *models.DayStatus  `json:"day_status"`
	Reason         *string            `json:"reason"`
	AvailableMeals *[]models.MealType `json:"available_meals"`
}

// scheduleService implements ScheduleService
type scheduleService struct {
	scheduleRepo repository.ScheduleRepository
}

// NewScheduleService creates a new schedule service
func NewScheduleService(scheduleRepo repository.ScheduleRepository) ScheduleService {
	return &scheduleService{
		scheduleRepo: scheduleRepo,
	}
}

// GetSchedule gets a day schedule by date
func (s *scheduleService) GetSchedule(date string) (*models.DaySchedule, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	return s.scheduleRepo.FindByDate(date)
}

// GetScheduleRange gets day schedules within a date range
func (s *scheduleService) GetScheduleRange(startDate, endDate string) ([]models.DaySchedule, error) {
	// Validate date formats
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return nil, fmt.Errorf("invalid start date format, expected YYYY-MM-DD: %w", err)
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		return nil, fmt.Errorf("invalid end date format, expected YYYY-MM-DD: %w", err)
	}

	return s.scheduleRepo.FindByDateRange(startDate, endDate)
}

// CreateSchedule creates a new day schedule
func (s *scheduleService) CreateSchedule(adminID string, input CreateScheduleInput) (*models.DaySchedule, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", input.Date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Check if schedule already exists for this date
	existing, err := s.scheduleRepo.FindByDate(input.Date)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("schedule already exists for date %s", input.Date)
	}

	// Parse admin UUID
	adminUUID, err := uuid.Parse(adminID)
	if err != nil {
		return nil, fmt.Errorf("invalid admin ID: %w", err)
	}

	if !input.DayStatus.IsValid() {
		return nil, fmt.Errorf("invalid day status: %s", input.DayStatus)
	}

	if len(input.AvailableMeals) > 0 {
		for _, meal := range input.AvailableMeals {
		    if !models.MealType(meal).IsValid() {
		        return nil, fmt.Errorf("invalid meal type: %s", meal)
		    }
		}
	}

	// Convert meal types slice to comma-separated string
	mealsStr := serializeMealTypes(input.AvailableMeals)

	// Create reason pointer if not empty
	var reasonPtr *string
	if input.Reason != "" {
		reasonPtr = &input.Reason
	}

	schedule := &models.DaySchedule{
		ID:             uuid.New(),
		Date:           input.Date,
		DayStatus:      input.DayStatus,
		Reason:         reasonPtr,
		AvailableMeals: &mealsStr,
		CreatedBy:      &adminUUID,
	}

	if err := s.scheduleRepo.Create(schedule); err != nil {
		return nil, err
	}

	return schedule, nil
}

// UpdateSchedule updates an existing day schedule
func (s *scheduleService) UpdateSchedule(id string, input UpdateScheduleInput) (*models.DaySchedule, error) {
	// Find existing schedule
	schedule, err := s.scheduleRepo.FindByDate(id)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return nil, fmt.Errorf("schedule not found")
	}

	// Update fields if provided
	if input.DayStatus != nil {
		schedule.DayStatus = *input.DayStatus
	}
	if input.Reason != nil {
		schedule.Reason = input.Reason
	}
	if input.AvailableMeals != nil {
		mealsStr := serializeMealTypes(*input.AvailableMeals)
		schedule.AvailableMeals = &mealsStr
	}

	if err := s.scheduleRepo.Update(schedule); err != nil {
		return nil, err
	}

	return schedule, nil
}

// DeleteSchedule deletes a day schedule
func (s *scheduleService) DeleteSchedule(id string) error {
	return s.scheduleRepo.Delete(id)
}

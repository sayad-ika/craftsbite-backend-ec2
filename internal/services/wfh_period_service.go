package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// WFHPeriodService defines business logic for company-wide WFH periods
type WFHPeriodService interface {
	CreatePeriod(adminID, startDate, endDate string, reason *string) (*WFHPeriodResponse, error)
	ListPeriods() ([]WFHPeriodResponse, error)
	DeletePeriod(id string) error
	IsDateInWFHPeriod(date string) (bool, error)
}

// WFHPeriodResponse is returned to clients
type WFHPeriodResponse struct {
	ID        string  `json:"id"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Reason    *string `json:"reason,omitempty"`
	CreatedBy string  `json:"created_by"`
	Active    bool    `json:"active"`
	CreatedAt string  `json:"created_at"`
}

type wfhPeriodService struct {
	repo repository.WFHPeriodRepository
}

func NewWFHPeriodService(repo repository.WFHPeriodRepository) WFHPeriodService {
	return &wfhPeriodService{repo: repo}
}

// CreatePeriod creates a new company-wide WFH period
func (s *wfhPeriodService) CreatePeriod(adminID, startDate, endDate string, reason *string) (*WFHPeriodResponse, error) {
	// Validate dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD")
	}
	if end.Before(start) {
		return nil, fmt.Errorf("end_date must not be before start_date")
	}

	adminUUID, err := uuid.Parse(adminID)
	if err != nil {
		return nil, fmt.Errorf("invalid admin ID")
	}

	period := &models.WFHPeriod{
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    reason,
		CreatedBy: adminUUID,
		Active:    true,
	}

	if err := s.repo.Create(period); err != nil {
		return nil, err
	}

	return &WFHPeriodResponse{
		ID:        period.ID.String(),
		StartDate: period.StartDate,
		EndDate:   period.EndDate,
		Reason:    period.Reason,
		CreatedBy: period.CreatedBy.String(),
		Active:    period.Active,
		CreatedAt: period.CreatedAt.Format(time.RFC3339),
	}, nil
}

// ListPeriods returns all WFH periods
func (s *wfhPeriodService) ListPeriods() ([]WFHPeriodResponse, error) {
	periods, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	result := make([]WFHPeriodResponse, 0, len(periods))
	for _, p := range periods {
		result = append(result, WFHPeriodResponse{
			ID:        p.ID.String(),
			StartDate: p.StartDate,
			EndDate:   p.EndDate,
			Reason:    p.Reason,
			CreatedBy: p.CreatedBy.String(),
			Active:    p.Active,
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

// DeletePeriod deletes a WFH period by ID
func (s *wfhPeriodService) DeletePeriod(id string) error {
	return s.repo.Delete(id)
}

// IsDateInWFHPeriod checks if a given date falls within any active WFH period
func (s *wfhPeriodService) IsDateInWFHPeriod(date string) (bool, error) {
	period, err := s.repo.FindActiveByDate(date)
	if err != nil {
		return false, err
	}
	return period != nil, nil
}

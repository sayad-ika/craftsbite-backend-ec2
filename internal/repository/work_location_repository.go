package repository

import (
	"craftsbite-backend/internal/models"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WorkLocationRepository interface {
	Upsert(wl *models.WorkLocation) error
	FindByUserAndDate(userID, date string) (*models.WorkLocation, error)
	FindByDate(date string) ([]models.WorkLocation, error)
	FindByDateAndUserIDs(date string, userIDs []string) ([]models.WorkLocation, error)
	CountWFHByUserAndMonth(userID, yearMonth string) (int64, error)
	GetMonthlyWFHCountsByUsers(yearMonth string, userIDs []string) (map[string]int64, error)
}

type workLocationRepository struct {
	db *gorm.DB
}

func NewWorkLocationRepository(db *gorm.DB) WorkLocationRepository {
	return &workLocationRepository{db: db}
}

func (r *workLocationRepository) Upsert(wl *models.WorkLocation) error {
	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"location", "set_by", "reason", "updated_at"}),
	}).Create(wl)
	if result.Error != nil {
		return fmt.Errorf("failed to upsert work location: %w", result.Error)
	}
	return nil
}

func (r *workLocationRepository) FindByUserAndDate(userID, date string) (*models.WorkLocation, error) {
	var wl models.WorkLocation
	err := r.db.Where("user_id = ? AND date = ?", userID, date).First(&wl).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find work location: %w", err)
	}
	return &wl, nil
}

func (r *workLocationRepository) FindByDate(date string) ([]models.WorkLocation, error) {
	var wls []models.WorkLocation
	err := r.db.Preload("User").Where("date = ?", date).Find(&wls).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list work locations: %w", err)
	}
	return wls, nil
}

func (r *workLocationRepository) FindByDateAndUserIDs(date string, userIDs []string) ([]models.WorkLocation, error) {
	if len(userIDs) == 0 {
		return []models.WorkLocation{}, nil
	}
	var wls []models.WorkLocation
	err := r.db.Preload("User").Where("date = ? AND user_id IN ?", date, userIDs).Find(&wls).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list work locations for team: %w", err)
	}
	return wls, nil
}

func (r *workLocationRepository) CountWFHByUserAndMonth(userID, yearMonth string) (int64, error) {
	// Parse "2026-02" → compute start and exclusive end
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return 0, fmt.Errorf("invalid yearMonth format, expected YYYY-MM: %w", err)
	}

	startDate := t.Format("2006-01-02")
	endDate := t.AddDate(0, 1, 0).Format("2006-01-02")

	var count int64
	err = r.db.Model(&models.WorkLocation{}).
		Where("user_id = ? AND location = 'wfh' AND date >= ? AND date <= ?", userID, startDate, endDate).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count WFH days: %w", err)
	}
	return count, nil
}

func (r *workLocationRepository) GetMonthlyWFHCountsByUsers(yearMonth string, userIDs []string) (map[string]int64, error) {
    if len(userIDs) == 0 {
        return map[string]int64{}, nil
    }

    t, err := time.Parse("2006-01", yearMonth)
    if err != nil {
        return nil, fmt.Errorf("invalid yearMonth format, expected YYYY-MM: %w", err)
    }

    startDate := t.Format("2006-01-02")
    endDate := t.AddDate(0, 1, 0).Format("2006-01-02")

    type row struct {
        UserID string
        Count  int64
    }

    var rows []row
    err = r.db.Model(&models.WorkLocation{}).
        Select("user_id, COUNT(*) as count").
        Where("location = 'wfh' AND date >= ? AND date < ? AND user_id IN ?", startDate, endDate, userIDs).
        Group("user_id").
        Scan(&rows).Error
    if err != nil {
        return nil, fmt.Errorf("failed to count WFH days by user: %w", err)
    }

    result := make(map[string]int64, len(userIDs))
    for _, r := range rows {
        result[r.UserID] = r.Count
    }
    return result, nil
}

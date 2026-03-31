package services

import (
	"craftsbite-backend/internal/config"
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type WorkLocationService interface {
	SetMyLocation(userID, date, location string) error
	GetMyLocation(userID, date string) (*WorkLocationResponse, error)
	SetLocationFor(requesterID, targetUserID, date, location string, reason *string) error
	ListByDate(requesterID, date string) ([]WorkLocationResponse, error)
	GetMonthlySummary(userID, yearMonth string) (*MonthlyWFHSummary, error)
	GetTeamMonthlyReport(requesterID, yearMonth string) (*TeamMonthlyReport, error)
}

type MonthlyWFHSummary struct {
    YearMonth  string `json:"year_month"`
    WFHDays    int64  `json:"wfh_days"`
    Allowance  int    `json:"allowance"`
    IsOverLimit bool  `json:"is_over_limit"`
}

type WorkLocationResponse struct {
    UserID   string  `json:"user_id"`
    Date     string  `json:"date"`
    Location string  `json:"location"`
    SetBy    string  `json:"set_by,omitempty"`
    Reason   *string `json:"reason,omitempty"`
}

type MemberWFHSummary struct {
    UserID      string `json:"user_id"`
    WFHDays     int64  `json:"wfh_days"`
    IsOverLimit bool   `json:"is_over_limit"`
    ExtraDays   int64  `json:"extra_days"`
}

type TeamMonthlyReport struct {
    YearMonth      string             `json:"year_month"`
    Allowance      int                `json:"allowance"`
    TotalEmployees int                `json:"total_employees"`
    OverLimitCount int                `json:"over_limit_count"`
    TotalExtraDays int64              `json:"total_extra_days"`
    Members        []MemberWFHSummary `json:"members"`
}

type workLocationService struct {
	repo        repository.WorkLocationRepository
	userRepo    repository.UserRepository
	teamRepo    repository.TeamRepository
	wfhPeriodRepo repository.WFHPeriodRepository
	historyRepo repository.WorkLocationHistoryRepository
	monthlyWFHAllowance int
}

func NewWorkLocationService(
	repo repository.WorkLocationRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
	wfhPeriodRepo repository.WFHPeriodRepository,
	historyRepo repository.WorkLocationHistoryRepository,
	cfg *config.Config,
) WorkLocationService {
	return &workLocationService{
		repo:                repo,
		userRepo:            userRepo,
		teamRepo:            teamRepo,
		wfhPeriodRepo:       wfhPeriodRepo,
		historyRepo:         historyRepo,
		monthlyWFHAllowance: cfg.WorkLocation.MonthlyWFHAllowance,
	}
}

func validateLocation(location string) error {
	if !models.WorkLocationType(location).IsValid() {
		return fmt.Errorf("location must be 'office' or 'wfh'")
	}
	return nil
}

func (s *workLocationService) SetMyLocation(userID, date, location string) error {
	if err := validateDate(date); err != nil {
		return err
	}
	if err := validateLocation(location); err != nil {
		return err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	existing, err := s.repo.FindByUserAndDate(userID, date)
	if err != nil {
		return fmt.Errorf("failed to check existing work location: %w", err)
	}

	wl := &models.WorkLocation{
		UserID:   userUUID,
		Date:     date,
		Location: models.WorkLocationType(location),
		SetBy:    nil,
	}
	if err := s.repo.Upsert(wl); err != nil {
		return err
	}

	var previousLocation *string
	if existing != nil {
		prev := string(existing.Location)
		previousLocation = &prev
	}

	history := &models.WorkLocationHistory{
		ID:               uuid.New(),
		UserID:           userUUID,
		Date:             date,
		Location:         models.WorkLocationType(location),
		Action:           models.HistoryActionOptedIn,
		PreviousLocation: previousLocation,
		OverrideByUserID: nil,
		OverrideReason:   nil,
	}
	return s.historyRepo.Create(history)
}

func (s *workLocationService) GetMyLocation(userID, date string) (*WorkLocationResponse, error) {
	if err := validateDate(date); err != nil {
		return nil, err
	}

	wl, err := s.repo.FindByUserAndDate(userID, date)
	if err != nil {
		return nil, err
	}

    if wl != nil {
        resp := &WorkLocationResponse{UserID: userID, Date: date, Location: string(wl.Location)}
        if wl.SetBy != nil {
            resp.SetBy = wl.SetBy.String()
        }
        resp.Reason = wl.Reason
        return resp, nil
    }

	period, err := s.wfhPeriodRepo.FindActiveByDate(date)
    if err != nil {
        return nil, err
    }

    if period != nil {
        return &WorkLocationResponse{
            UserID:   userID,
            Date:     date,
            Location: "wfh",
            Reason:   period.Reason,
        }, nil
    }

    return &WorkLocationResponse{UserID: userID, Date: date, Location: "not_set"}, nil
}

func (s *workLocationService) SetLocationFor(requesterID, targetUserID, date, location string, reason *string) error {
	if err := validateDate(date); err != nil {
		return err
	}
	if err := validateLocation(location); err != nil {
		return err
	}

	requester, err := s.userRepo.FindByID(requesterID)
	if err != nil {
		return fmt.Errorf("requester not found")
	}

	if requester.Role == models.RoleTeamLead {
		isMember, err := s.teamRepo.IsUserInAnyTeamLedBy(requesterID, targetUserID)
		if err != nil {
			return fmt.Errorf("failed to verify team membership: %w", err)
		}
		if !isMember {
			return fmt.Errorf("you can only set work location for your own team members")
		}
	}

	targetUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return fmt.Errorf("invalid target user ID")
	}
	requesterUUID, err := uuid.Parse(requesterID)
	if err != nil {
		return fmt.Errorf("invalid requester ID")
	}

	existing, err := s.repo.FindByUserAndDate(targetUserID, date)
	if err != nil {
		return fmt.Errorf("failed to check existing work location: %w", err)
	}

	wl := &models.WorkLocation{
		UserID:   targetUUID,
		Date:     date,
		Location: models.WorkLocationType(location),
		SetBy:    &requesterUUID,
		Reason:   reason,
	}
	if err := s.repo.Upsert(wl); err != nil {
		return err
	}

	var previousLocation *string
	if existing != nil {
		prev := string(existing.Location)
		previousLocation = &prev
	}

	history := &models.WorkLocationHistory{
		ID:               uuid.New(),
		UserID:           targetUUID,
		Date:             date,
		Location:         models.WorkLocationType(location),
		Action:           models.HistoryActionOverrideIn,
		PreviousLocation: previousLocation,
		OverrideBy: &requesterUUID,
		OverrideReason:   reason,
	}
	return s.historyRepo.Create(history)
}

func (s *workLocationService) ListByDate(requesterID, date string) ([]WorkLocationResponse, error) {
    if err := validateDate(date); err != nil {
        return nil, err
    }

    requester, err := s.userRepo.FindByID(requesterID)
    if err != nil {
        return nil, fmt.Errorf("requester not found")
    }

    var wls []models.WorkLocation

    if requester.Role == models.RoleTeamLead {
        teams, err := s.teamRepo.FindByTeamLeadID(requesterID)
        if err != nil {
            return nil, fmt.Errorf("failed to load teams: %w", err)
        }
        var memberIDs []string
        for _, team := range teams {
            for _, member := range team.Members {
                memberIDs = append(memberIDs, member.ID.String())
            }
        }
        wls, err = s.repo.FindByDateAndUserIDs(date, memberIDs)
        if err != nil {
            return nil, err
        }
    } else {
        wls, err = s.repo.FindByDate(date)
        if err != nil {
            return nil, err
        }
    }

    var result []WorkLocationResponse
    for _, wl := range wls {
		item := WorkLocationResponse{
		    UserID:   wl.UserID.String(),
		    Date:     wl.Date,
		    Location: string(wl.Location),
		    Reason:   wl.Reason,
		}
		if wl.SetBy != nil {
		    item.SetBy = wl.SetBy.String()
		}
        result = append(result, item)
    }
    if result == nil {
        result = []WorkLocationResponse{}
    }
    return result, nil
}

func (s *workLocationService) GetMonthlySummary(userID, yearMonth string) (*MonthlyWFHSummary, error) {
    count, err := s.repo.CountWFHByUserAndMonth(userID, yearMonth)
    if err != nil {
        return nil, err
    }
    return &MonthlyWFHSummary{
        YearMonth:   yearMonth,
        WFHDays:     count,
        Allowance:   s.monthlyWFHAllowance,
        IsOverLimit: count > int64(s.monthlyWFHAllowance),
    }, nil
}

func (s *workLocationService) GetTeamMonthlyReport(requesterID, yearMonth string) (*TeamMonthlyReport, error) {
    requester, err := s.userRepo.FindByID(requesterID)
    if err != nil {
        return nil, fmt.Errorf("requester not found")
    }

    var userIDs []string

    if requester.Role == models.RoleTeamLead {
        teams, err := s.teamRepo.FindByTeamLeadID(requesterID)
        if err != nil {
            return nil, fmt.Errorf("failed to load teams: %w", err)
        }
        for _, team := range teams {
            for _, member := range team.Members {
                userIDs = append(userIDs, member.ID.String())
            }
        }
    } else {
        users, err := s.userRepo.FindAll(map[string]interface{}{"active": true})
        if err != nil {
            return nil, fmt.Errorf("failed to load users: %w", err)
        }
        for _, u := range users {
            userIDs = append(userIDs, u.ID.String())
        }
    }

    counts, err := s.repo.GetMonthlyWFHCountsByUsers(yearMonth, userIDs)
    if err != nil {
        return nil, err
    }

    rollup := &TeamMonthlyReport{
        YearMonth:  yearMonth,
        Allowance:  s.monthlyWFHAllowance,
        TotalEmployees: len(userIDs),
        Members:    make([]MemberWFHSummary, 0, len(userIDs)),
    }

    for _, id := range userIDs {
        wfhDays := counts[id]
        extra := wfhDays - int64(s.monthlyWFHAllowance)
        if extra < 0 {
            extra = 0
        }
        member := MemberWFHSummary{
            UserID:      id,
            WFHDays:     wfhDays,
            IsOverLimit: wfhDays > int64(s.monthlyWFHAllowance),
            ExtraDays:   extra,
        }
        if member.IsOverLimit {
            rollup.OverLimitCount++
            rollup.TotalExtraDays += extra
        }
        rollup.Members = append(rollup.Members, member)
    }

    return rollup, nil
}

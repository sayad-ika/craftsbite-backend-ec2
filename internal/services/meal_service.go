package services

import (
	"craftsbite-backend/internal/config"
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// MealService defines the interface for meal participation business logic
type MealService interface {
	GetTodayMeals(userID string) (*TodayMealsResponse, error)
	GetParticipation(userID, date string) ([]ParticipationStatus, error)
	SetParticipation(userID, date, mealType string, participating bool) error
	OverrideParticipation(adminID, userID, date, mealType string, participating bool, reason string) error
	GetTeamParticipation(teamLeadID, date string) (*TeamParticipationResponse, error)
	GetAllTeamsParticipation(date string) (*TeamParticipationResponse, error)
}

// TodayMealsResponse represents the response for today's meals
type TodayMealsResponse struct {
	Date           string                `json:"date"`
	DayStatus      models.DayStatus      `json:"day_status"`
	AvailableMeals []models.MealType     `json:"available_meals"`
	Participations []ParticipationStatus `json:"participations"`
}

// ParticipationStatus represents a user's participation status for a meal
type ParticipationStatus struct {
	MealType        models.MealType `json:"meal_type"`
	IsParticipating bool            `json:"is_participating"`
	Source          string          `json:"source"`
}

// mealService implements MealService
type mealService struct {
	mealRepo       repository.MealRepository
	scheduleRepo   repository.ScheduleRepository
	historyRepo    repository.HistoryRepository
	userRepo       repository.UserRepository
	teamRepo       repository.TeamRepository
    wlRepo              repository.WorkLocationRepository
	resolver       ParticipationResolver
	cutoffTime     string
	cutoffTimezone string
    forwardWindowDays int
    monthlyWFHAllowance int
}

type TeamMemberParticipation struct {
    UserID   string                 `json:"user_id"`
    Name     string                 `json:"name"`
    Email    string                 `json:"email"`
    Meals    map[string]bool        `json:"meals"`
	IsOverWFHLimit  bool            `json:"is_over_wfh_limit"`
}

type TeamParticipationGroup struct {
    TeamID         string                    `json:"team_id"`
    TeamName       string                    `json:"team_name"`
    TeamLeadUserID string                    `json:"team_lead_user_id"`
    Members        []TeamMemberParticipation `json:"members"`
}

type TeamParticipationResponse struct {
    Date  string                   `json:"date"`
    Teams []TeamParticipationGroup `json:"teams"`
}

// NewMealService creates a new meal service
func NewMealService(
	mealRepo repository.MealRepository,
	scheduleRepo repository.ScheduleRepository,
	historyRepo repository.HistoryRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
	workLocationRepo repository.WorkLocationRepository,
	resolver ParticipationResolver,
	cfg *config.Config,
) MealService {
	return &mealService{
		mealRepo:       mealRepo,
		scheduleRepo:   scheduleRepo,
		historyRepo:    historyRepo,
		userRepo:       userRepo,
		teamRepo:       teamRepo,
		wlRepo:         workLocationRepo,
		resolver:       resolver,
		cutoffTime:     cfg.Meal.CutoffTime,
		cutoffTimezone: cfg.Meal.CutoffTimezone,
	    forwardWindowDays: cfg.Meal.ForwardWindowDays,
		monthlyWFHAllowance: cfg.WorkLocation.MonthlyWFHAllowance,
	}
}

// GetTodayMeals gets tomorrow's meals and participation status for a user
// Note: "Today" refers to the meals users should decide on today, which are tomorrow's meals
// because the cutoff is previous day at 9:00 PM
func (s *mealService) GetTodayMeals(userID string) (*TodayMealsResponse, error) {
	// Return tomorrow's meals (cutoff is previous day 9:00 PM)
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	// Get day schedule for tomorrow
	schedule, err := s.scheduleRepo.FindByDate(tomorrow)
	if err != nil {
		return nil, err
	}

	response := &TodayMealsResponse{
		Date:           tomorrow,
		DayStatus:      models.DayStatusNormal,
		AvailableMeals: []models.MealType{},
		Participations: []ParticipationStatus{},
	}

	if schedule != nil {
		response.DayStatus = schedule.DayStatus
		if schedule.AvailableMeals != nil {
			response.AvailableMeals = parseMealTypes(*schedule.AvailableMeals)
		}
	}

	// Resolve participation for each available meal
	for _, mealType := range response.AvailableMeals {
		isParticipating, source, err := s.resolver.ResolveParticipation(userID, tomorrow, string(mealType))
		if err != nil {
			return nil, err
		}

		response.Participations = append(response.Participations, ParticipationStatus{
			MealType:        mealType,
			IsParticipating: isParticipating,
			Source:          source,
		})
	}

	return response, nil
}

// GetParticipation gets participation status for a user on a specific date
func (s *mealService) GetParticipation(userID, date string) ([]ParticipationStatus, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Get day schedule to know available meals
	schedule, err := s.scheduleRepo.FindByDate(date)
	if err != nil {
		return nil, err
	}

	availableMeals := []models.MealType{models.MealTypeLunch, models.MealTypeSnacks}
	if schedule != nil && schedule.AvailableMeals != nil {
		availableMeals = parseMealTypes(*schedule.AvailableMeals)
	}

	participations := []ParticipationStatus{}
	for _, mealType := range availableMeals {
		isParticipating, source, err := s.resolver.ResolveParticipation(userID, date, string(mealType))
		if err != nil {
			return nil, err
		}

		participations = append(participations, ParticipationStatus{
			MealType:        mealType,
			IsParticipating: isParticipating,
			Source:          source,
		})
	}

	return participations, nil
}

// SetParticipation sets a user's participation for a specific date and meal
func (s *mealService) SetParticipation(userID, date, mealType string, participating bool) error {
	err := s.validateDateWindow(date)
	if err != nil {
		return err
	}

	// Check if existing record exists to get its ID for proper upsert
	existing, err := s.mealRepo.FindByUserDateMeal(userID, date, mealType)
	if err != nil {
		return fmt.Errorf("failed to check existing participation: %w", err)
	}

	// Create or update participation record
	var participationID uuid.UUID
	if existing != nil {
		participationID = existing.ID // Reuse existing ID for UPDATE
	} else {
		participationID = uuid.New() // New ID for INSERT
	}

	participation := &models.MealParticipation{
		ID:              participationID,
		UserID:          uuid.MustParse(userID),
		Date:            date,
		MealType:        models.MealType(mealType),
		IsParticipating: participating,
	}

	if participating {
		participation.OptedOutAt = nil
	} else {
		now := time.Now()
		participation.OptedOutAt = &now
	}

	if err := s.mealRepo.CreateOrUpdate(participation); err != nil {
		return err
	}

	// Record in history
	action := models.HistoryActionOptedOut
	if participating {
		action = models.HistoryActionOptedIn
	}

	history := &models.MealParticipationHistory{
		ID:       uuid.New(),
		UserID:   uuid.MustParse(userID),
		Date:     date,
		MealType: models.MealType(mealType),
		Action:   action,
	}

	return s.historyRepo.Create(history)
}

// OverrideParticipation allows an admin or team lead to override a user's participation
// Team leads can only override their own team members
func (s *mealService) OverrideParticipation(requesterID, userID, date, mealType string, participating bool, reason string) error {
	err := s.validateDateWindow(date)
	if err != nil {
		return err
	}

	// Parse requester UUID
	requesterUUID, err := uuid.Parse(requesterID)
	if err != nil {
		return fmt.Errorf("invalid requester ID: %w", err)
	}

	// Get requester's role to validate permissions
	requester, err := s.userRepo.FindByID(requesterID)
	if err != nil {
		return fmt.Errorf("failed to find requester: %w", err)
	}

	// If requester is a team lead, verify they manage the target user
	if requester.Role == models.RoleTeamLead {
		isMember, err := s.teamRepo.IsUserInAnyTeamLedBy(requesterID, userID)
		if err != nil {
			return fmt.Errorf("failed to check team membership: %w", err)
		}
		if !isMember {
			return fmt.Errorf("team lead can only override participation for their own team members")
		}
	}
	// Admin and Logistics roles can override anyone (no additional check needed)

	// Check if existing record exists to get its ID for proper upsert
	existing, err := s.mealRepo.FindByUserDateMeal(userID, date, mealType)
	if err != nil {
		return fmt.Errorf("failed to check existing participation: %w", err)
	}

	// Create or update participation record with override info
	var participationID uuid.UUID
	if existing != nil {
		participationID = existing.ID // Reuse existing ID for UPDATE
	} else {
		participationID = uuid.New() // New ID for INSERT
	}

	participation := &models.MealParticipation{
		ID:              participationID,
		UserID:          uuid.MustParse(userID),
		Date:            date,
		MealType:        models.MealType(mealType),
		IsParticipating: participating,
		OverrideBy:      &requesterUUID,
		OverrideReason:  &reason,
	}

	if err := s.mealRepo.CreateOrUpdate(participation); err != nil {
		return err
	}

	// Record in history
	action := models.HistoryActionOverrideOut
	if participating {
		action = models.HistoryActionOverrideIn
	}

	previousValue := func() *string {
    if existing != nil {
        v := strconv.FormatBool(existing.IsParticipating)
        return &v
    }
    return nil
	}()

	history := &models.MealParticipationHistory{
		ID:              uuid.New(),
		UserID:          uuid.MustParse(userID),
		Date:            date,
		MealType:        models.MealType(mealType),
		Action:          action,
		PreviousValue:   previousValue,
		Reason:          &reason,
		ChangedByUserID: &requesterUUID,
	}

	return s.historyRepo.Create(history)
}

// validateCutoffTime checks if the current time is before the cutoff time for the given date
// Cutoff is on the PREVIOUS day at the configured time (e.g., 9:00 PM the day before)
func (s *mealService) validateCutoffTime(targetDate time.Time) error {
	// Load timezone
	loc, err := time.LoadLocation(s.cutoffTimezone)
	if err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}

	// Parse cutoff time (e.g., "21:00" for 9:00 PM)
	cutoffParts := s.cutoffTime
	cutoffTime, err := time.Parse("15:04", cutoffParts)
	if err != nil {
		return fmt.Errorf("invalid cutoff time format: %w", err)
	}

	// Cutoff is on the PREVIOUS day at the configured time
	cutoffDate := targetDate.AddDate(0, 0, -1)
	cutoffDateTime := time.Date(
		cutoffDate.Year(),
		cutoffDate.Month(),
		cutoffDate.Day(),
		cutoffTime.Hour(),
		cutoffTime.Minute(),
		0, 0, loc,
	)

	// Get current time in the configured timezone
	now := time.Now().In(loc)

	// Check if current time is past the cutoff
	if now.After(cutoffDateTime) {
		return fmt.Errorf("cutoff time (%s %s on %s) has passed for date %s",
			s.cutoffTime, s.cutoffTimezone, cutoffDate.Format("2006-01-02"), targetDate.Format("2006-01-02"))
	}

	return nil
}

func (s *mealService) getMealStatus(userID, date string, availableMeals []models.MealType) (map[string]bool, error) {
    mealStatus := make(map[string]bool)
    for _, mealType := range availableMeals {
        isParticipating, _, err := s.resolver.ResolveParticipation(userID, date, string(mealType))
        if err != nil {
            return nil, err
        }
        mealStatus[string(mealType)] = isParticipating
    }
    return mealStatus, nil
}

func (s *mealService) getAvailableMeals(date string) ([]models.MealType, error) {
    schedule, err := s.scheduleRepo.FindByDate(date)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch schedule for %s: %w", date, err)
    }
    if schedule != nil && schedule.AvailableMeals != nil {
        return parseMealTypes(*schedule.AvailableMeals), nil
    }
    return []models.MealType{}, nil
}

func (s *mealService) GetTeamParticipation(teamLeadID, date string) (*TeamParticipationResponse, error) {
    availableMeals, err := s.getAvailableMeals(date)
    if err != nil {
        return nil, err
    }
    return s.getTeamParticipationWithMeals(teamLeadID, date, availableMeals)
}

func (s *mealService) getTeamParticipationWithMeals(teamLeadID, date string, availableMeals []models.MealType) (*TeamParticipationResponse, error) {
    teams, err := s.teamRepo.FindByTeamLeadID(teamLeadID)
    if err != nil {
        return nil, fmt.Errorf("failed to find teams: %w", err)
    }

    var teamGroups []TeamParticipationGroup
    for _, team := range teams {
        var members []TeamMemberParticipation
        for _, member := range team.Members {
            memberID := member.ID.String()
            mealStatus, err := s.getMealStatus(memberID, date, availableMeals)
            if err != nil {
                return nil, err
            }
            members = append(members, TeamMemberParticipation{
                UserID: memberID,
                Name:   member.Name,
                Email:  member.Email,
                Meals:  mealStatus,
    			IsOverWFHLimit: s.isOverWFHLimit(memberID),
            })
        }
        if members == nil {
            members = []TeamMemberParticipation{}
        }
        teamGroups = append(teamGroups, TeamParticipationGroup{
            TeamID:         team.ID.String(),
            TeamName:       team.Name,
            TeamLeadUserID: teamLeadID,
            Members:        members,
        })
    }

    if teamGroups == nil {
        teamGroups = []TeamParticipationGroup{}
    }
    return &TeamParticipationResponse{
        Date:  date,
        Teams: teamGroups,
    }, nil
}

func (s *mealService) GetAllTeamsParticipation(date string) (*TeamParticipationResponse, error) {
    teams, err := s.teamRepo.FindAllWithMembers()
    if err != nil {
        return nil, fmt.Errorf("failed to find teams: %w", err)
    }

    availableMeals, err := s.getAvailableMeals(date)
    if err != nil {
        return nil, err
    }

    var teamGroups []TeamParticipationGroup
    for _, team := range teams {
        leadMealStatus, err := s.getMealStatus(team.TeamLeadID.String(), date, availableMeals)
        if err != nil {
            return nil, err
        }

        leadMember := TeamMemberParticipation{
            UserID: team.TeamLeadID.String(),
            Name:   team.TeamLead.Name,
            Email:  team.TeamLead.Email,
            Meals:  leadMealStatus,
		    IsOverWFHLimit: s.isOverWFHLimit(team.TeamLeadID.String()),
        }

        var members []TeamMemberParticipation
        for _, member := range team.Members {
            mealStatus, err := s.getMealStatus(member.ID.String(), date, availableMeals)
            if err != nil {
                return nil, err
            }
            members = append(members, TeamMemberParticipation{
                UserID: member.ID.String(),
                Name:   member.Name,
                Email:  member.Email,
                Meals:  mealStatus,
				IsOverWFHLimit: s.isOverWFHLimit(member.ID.String()),
            })
        }

        teamGroups = append(teamGroups, TeamParticipationGroup{
            TeamID:         team.ID.String(),
            TeamName:       team.Name,
            TeamLeadUserID: team.TeamLeadID.String(),
            Members:        append([]TeamMemberParticipation{leadMember}, members...),
        })
    }

    if teamGroups == nil {
        teamGroups = []TeamParticipationGroup{}
    }
    return &TeamParticipationResponse{
        Date:  date,
        Teams: teamGroups,
    }, nil
}

// Helper function to validate date window and cutoff time
func (s *mealService) validateDateWindow(date string) error {
	if err := validateDate(date); err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	parsedDate, _ := time.Parse("2006-01-02", date)

	today := time.Now().Truncate(24 * time.Hour)
	targetDay := parsedDate.Truncate(24 * time.Hour)

	if targetDay.Before(today) {
		return fmt.Errorf("cannot set participation for a past date: %s", date)
	}

	maxDate := today.AddDate(0, 0, s.forwardWindowDays)
	if targetDay.After(maxDate) {
		return fmt.Errorf("cannot set participation more than %d days in advance (requested: %s)", s.forwardWindowDays, date)
	}

	if err := s.validateCutoffTime(parsedDate); err != nil {
		return err
	}

	return nil
}

func (s *mealService) isOverWFHLimit(userID string) bool {
    yearMonth := time.Now().Format("2006-01")
    count, err := s.wlRepo.CountWFHByUserAndMonth(userID, yearMonth)
    if err != nil {
        return false
    }
    return count > int64(s.monthlyWFHAllowance)
}

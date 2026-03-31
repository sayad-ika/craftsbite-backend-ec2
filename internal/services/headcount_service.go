package services

import (
	"craftsbite-backend/internal/config"
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"strings"
	"time"
)

// HeadcountService defines the interface for headcount calculations
type HeadcountService interface {
	GetTodayHeadcount() ([]*DailyHeadcountSummary, error)
	GetHeadcountByDate(date string) (*DailyHeadcountSummary, error)
	GetDetailedHeadcount(date, mealType string) (*DetailedHeadcount, error)
	GenerateAnnouncement(date string) (string, error)
	GetForecast(days int) ([]*DailyHeadcountSummary, error)
}

// MealHeadcount represents participation breakdown for a single meal
type MealHeadcount struct {
	Participating int `json:"participating"`
	OptedOut      int `json:"opted_out"`
}

// DailyHeadcountSummary represents the headcount summary for a day
type DailyHeadcountSummary struct {
	Date             string                   `json:"date"`
	DayStatus        models.DayStatus         `json:"day_status"`
	TotalActiveUsers int                      `json:"total_active_users"`
	LocationSplit    LocationSplit            `json:"location_split"`
	Meals            map[string]MealHeadcount `json:"meals"`
	Teams            []TeamHeadcount          `json:"teams"`
}

type LocationSplit struct {
	Office int `json:"office"`
	WFH    int `json:"wfh"`
	NotSet int `json:"not_set"`
}

type TeamHeadcount struct {
	TeamID        string                   `json:"team_id"`
	TeamName      string                   `json:"team_name"`
	TotalMembers  int                      `json:"total_members"`
	LocationSplit LocationSplit            `json:"location_split"`
	Meals         map[string]MealHeadcount `json:"meals"`
}

// DetailedHeadcount represents detailed headcount for a specific meal
type DetailedHeadcount struct {
	Date            string            `json:"date"`
	MealType        string            `json:"meal_type"`
	Participants    []ParticipantInfo `json:"participants"`
	NonParticipants []ParticipantInfo `json:"non_participants"`
	TotalCount      int               `json:"total_count"`
}

// ParticipantInfo represents a user's participation info
type ParticipantInfo struct {
	UserID          string `json:"user_id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	IsParticipating bool   `json:"is_participating"`
	Source          string `json:"source"`
}

// headcountService implements HeadcountService
type headcountService struct {
	userRepo         repository.UserRepository
	scheduleRepo     repository.ScheduleRepository
	resolver         ParticipationResolver
	teamRepo         repository.TeamRepository
	workLocationRepo repository.WorkLocationRepository
	wfhPeriodRepo    repository.WFHPeriodRepository
	maxForecastDays   int
}

// NewHeadcountService creates a new headcount service
func NewHeadcountService(
	userRepo repository.UserRepository,
	scheduleRepo repository.ScheduleRepository,
	resolver ParticipationResolver,
	teamRepo repository.TeamRepository,
	workLocationRepo repository.WorkLocationRepository,
	wfhPeriodRepo repository.WFHPeriodRepository,
	cfg *config.Config,
) HeadcountService {
	return &headcountService{
		userRepo:         userRepo,
		scheduleRepo:     scheduleRepo,
		resolver:         resolver,
		teamRepo:         teamRepo,
		workLocationRepo: workLocationRepo,
		wfhPeriodRepo:    wfhPeriodRepo,
		maxForecastDays: cfg.Headcount.MaxForecastDays,
	}
}

// GetTodayHeadcount gets today's and tomorrow's headcount summary
func (s *headcountService) GetTodayHeadcount() ([]*DailyHeadcountSummary, error) {
	today := time.Now().Format("2006-01-02")

	todaySummary, err := s.GetHeadcountByDate(today)
	if err != nil {
		return nil, err
	}

	return []*DailyHeadcountSummary{todaySummary}, nil
}

func (s *headcountService) GetHeadcountByDate(date string) (*DailyHeadcountSummary, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	return s.getHeadcountByDate(date)
}

func (s *headcountService) getHeadcountByDate(date string) (*DailyHeadcountSummary, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	filters := map[string]interface{}{
		"active": true,
	}
	users, err := s.userRepo.FindAll(filters)
	if err != nil {
		return nil, err
	}

	schedule, err := s.scheduleRepo.FindByDate(date)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return nil, nil
	}

	dayStatus := schedule.DayStatus
	var availableMeals []models.MealType
	if schedule.AvailableMeals != nil {
		availableMeals = parseMealTypes(*schedule.AvailableMeals)
	}
	if len(availableMeals) == 0 {
		return nil, nil
	}

	totalActiveUsers := len(users)

	userLocationMap := make(map[string]string)
	globalLocationSplit := LocationSplit{}

	for _, user := range users {
		loc, err := s.resolveUserLocation(user.ID.String(), date)
		if err != nil {
			return nil, err
		}
		userLocationMap[user.ID.String()] = loc
		switch loc {
		case "office":
			globalLocationSplit.Office++
		case "wfh":
			globalLocationSplit.WFH++
		default:
			globalLocationSplit.NotSet++
		}
	}

	// ── Meal headcount (unchanged logic) ───────────────────────
	meals := make(map[string]MealHeadcount)
	type userMealResult struct {
		isParticipating bool
		source          string
	}

	userParticipation := make(map[string]map[string]userMealResult)

	for _, mealType := range availableMeals {
		mtKey := string(mealType)
		userParticipation[mtKey] = make(map[string]userMealResult)
		participating := 0

		for _, user := range users {
			uid := user.ID.String()
			isP, src, err := s.resolver.ResolveParticipation(uid, date, mtKey)
			if err != nil {
				return nil, err
			}
			userParticipation[mtKey][uid] = userMealResult{isParticipating: isP, source: src}
			if isP {
				participating++
			}
		}

		meals[mtKey] = MealHeadcount{
			Participating: participating,
			OptedOut:      totalActiveUsers - participating,
		}
	}

	// ── Team breakdown ─────────────────────────────────────────
	teams, err := s.teamRepo.FindAllWithMembers()
	if err != nil {
		return nil, err
	}

	teamHeadcounts := make([]TeamHeadcount, 0, len(teams))
	for _, team := range teams {
		th := TeamHeadcount{
			TeamID:       team.ID.String(),
			TeamName:     team.Name,
			TotalMembers: len(team.Members),
			Meals:        make(map[string]MealHeadcount),
		}

		// Location split for this team
		for _, member := range team.Members {
			uid := member.ID.String()
			loc := userLocationMap[uid]
			switch loc {
			case "office":
				th.LocationSplit.Office++
			case "wfh":
				th.LocationSplit.WFH++
			default:
				th.LocationSplit.NotSet++
			}
		}

		for _, mealType := range availableMeals {
			mtKey := string(mealType)
			participating := 0
			for _, member := range team.Members {
				uid := member.ID.String()
				if res, ok := userParticipation[mtKey][uid]; ok && res.isParticipating {
					participating++
				}
			}
			th.Meals[mtKey] = MealHeadcount{
				Participating: participating,
				OptedOut:      len(team.Members) - participating,
			}
		}

		teamHeadcounts = append(teamHeadcounts, th)
	}

	return &DailyHeadcountSummary{
		Date:             date,
		DayStatus:        dayStatus,
		TotalActiveUsers: totalActiveUsers,
		LocationSplit:    globalLocationSplit,
		Meals:            meals,
		Teams:            teamHeadcounts,
	}, nil
}

// GetDetailedHeadcount gets detailed headcount for a specific date and meal
func (s *headcountService) GetDetailedHeadcount(date, mealType string) (*DetailedHeadcount, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Get all active users
	filters := map[string]interface{}{
		"active": true,
	}
	users, err := s.userRepo.FindAll(filters)
	if err != nil {
		return nil, err
	}

	participants := []ParticipantInfo{}
	nonParticipants := []ParticipantInfo{}
	totalCount := 0

	for _, user := range users {
		if !user.Active {
			continue
		}

		isParticipating, source, err := s.resolver.ResolveParticipation(user.ID.String(), date, mealType)
		if err != nil {
			return nil, err
		}

		info := ParticipantInfo{
			UserID:          user.ID.String(),
			Name:            user.Name,
			Email:           user.Email,
			IsParticipating: isParticipating,
			Source:          source,
		}

		if isParticipating {
			participants = append(participants, info)
			totalCount++
		} else {
			nonParticipants = append(nonParticipants, info)
		}
	}

	return &DetailedHeadcount{
		Date:            date,
		MealType:        mealType,
		Participants:    participants,
		NonParticipants: nonParticipants,
		TotalCount:      totalCount,
	}, nil
}

func (s *headcountService) resolveUserLocation(userID, date string) (string, error) {
	wl, err := s.workLocationRepo.FindByUserAndDate(userID, date)
	if err != nil {
		return "", err
	}
	if wl != nil {
		return string(wl.Location), nil
	}

	period, err := s.wfhPeriodRepo.FindActiveByDate(date)
	if err != nil {
		return "", err
	}
	if period != nil {
		return "wfh", nil
	}

	return "not_set", nil
}

func (s *headcountService) GenerateAnnouncement(date string) (string, error) {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %w", err)
	}
	humanDate := parsed.Format("Monday, 2 January 2006")
	shortDate := parsed.Format("2 January 2006")
	weekday := strings.ToLower(parsed.Weekday().String())

	schedule, err := s.scheduleRepo.FindByDate(date)
	if err != nil {
		return "", err
	}

	hasScheduleWithMeals := false
	var availableMeals []models.MealType
	if schedule != nil && schedule.AvailableMeals != nil {
		availableMeals = parseMealTypes(*schedule.AvailableMeals)
		hasScheduleWithMeals = len(availableMeals) > 0
	}

	if !hasScheduleWithMeals {
		if weekday == "saturday" || weekday == "sunday" {
			return fmt.Sprintf("📅 %s\n🌅 Weekend — Enjoy your day off!", shortDate), nil
		}
		if schedule != nil {
			switch schedule.DayStatus {
			case models.DayStatusOfficeClosed:
				return fmt.Sprintf("📅 %s\n🚫 Office Closed — No meals today.", shortDate), nil
			case models.DayStatusGovtHoliday:
				return fmt.Sprintf("📅 %s\n🏛️ Government Holiday — No meals today.", shortDate), nil
			}
		}
		return fmt.Sprintf("📅 %s\n📭 No meals scheduled today.", shortDate), nil
	}

	summary, err := s.GetHeadcountByDate(date)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📅 Meal Update — %s", humanDate))

	switch schedule.DayStatus {
	case models.DayStatusGovtHoliday:
		sb.WriteString("\n🏛️  Government Holiday — Extra working day (meals available)")
	case models.DayStatusCelebration:
		sb.WriteString("\n🎉  Celebration Day!")
	}

	wfhPeriod, err := s.wfhPeriodRepo.FindActiveByDate(date)
	if err != nil {
		return "", err
	}
	if wfhPeriod != nil {
		note := "Company-wide WFH"
		if wfhPeriod.Reason != nil && *wfhPeriod.Reason != "" {
			note += " — " + *wfhPeriod.Reason
		}
		sb.WriteString(fmt.Sprintf("\n🏠  %s", note))
	}

	ls := summary.LocationSplit
	sb.WriteString(fmt.Sprintf(
		"\n\n👥 Total staff: %d  |  🏢 Office: %d  |  🏠 WFH: %d  |  ❓ Not set: %d",
		summary.TotalActiveUsers, ls.Office, ls.WFH, ls.NotSet,
	))

	mealOrder := []string{"lunch", "snacks", "iftar", "event_dinner", "optional_dinner"}
	mealEmoji := map[string]string{
		"lunch":           "🍽️ ",
		"snacks":          "🍪",
		"iftar":           "🌙",
		"event_dinner":    "🍴",
		"optional_dinner": "🥘",
	}
	mealLabel := map[string]string{
		"lunch":           "Lunch",
		"snacks":          "Snacks",
		"iftar":           "Iftar",
		"event_dinner":    "Event Dinner",
		"optional_dinner": "Optional Dinner",
	}
	sb.WriteString("\n")
	for _, mt := range mealOrder {
		counts, ok := summary.Meals[mt]
		if !ok {
			continue
		}
		sb.WriteString(fmt.Sprintf(
			"\n%s  %-15s →  %d joining, %d not joining",
			mealEmoji[mt], mealLabel[mt], counts.Participating, counts.OptedOut,
		))
	}

	sb.WriteString("\n\nPlease confirm your meal preference if you haven't already. Thank you! 🙏")

	return sb.String(), nil
}

func (s *headcountService) GetForecast(days int) ([]*DailyHeadcountSummary, error) {
	maxDays := s.maxForecastDays
    if days <= 0 {
        days = 7
    }
    if days > maxDays {
        days = maxDays
    }

    results := make([]*DailyHeadcountSummary, 0, days)
    today := time.Now()

    for i := 0; i <= days; i++ {
        next := today.AddDate(0, 0, i)
        if next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
            continue
        }

        date := next.Format("2006-01-02")
        summary, err := s.getHeadcountByDate(date)
        if err != nil {
            return nil, err
        }

        if summary == nil {
            results = append(results, &DailyHeadcountSummary{
                Date:      date,
                DayStatus: models.DayStatusNormal,
                Meals:     map[string]MealHeadcount{},
            })
        } else {
            results = append(results, summary)
        }
    }

    return results, nil
}

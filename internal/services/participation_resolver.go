package services

import (
	"craftsbite-backend/internal/config"
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"fmt"
	"strings"
	"time"
)

// ParticipationResolver defines the interface for resolving meal participation status
type ParticipationResolver interface {
	ResolveParticipation(userID, date, mealType string) (isParticipating bool, source string, Error error)
}

// participationResolver implements ParticipationResolver
type participationResolver struct {
	mealRepo       repository.MealRepository
	scheduleRepo   repository.ScheduleRepository
	bulkOptOutRepo repository.BulkOptOutRepository
	userRepo       repository.UserRepository
	weekendDays    map[string]bool
}

// NewParticipationResolver creates a new participation resolver
func NewParticipationResolver(
	mealRepo repository.MealRepository,
	scheduleRepo repository.ScheduleRepository,
	bulkOptOutRepo repository.BulkOptOutRepository,
	userRepo repository.UserRepository,
	cfg *config.Config,
) ParticipationResolver {
	// Build weekend days map for quick lookup
	weekendDays := make(map[string]bool)
	for _, day := range cfg.Meal.WeekendDays {
		weekendDays[strings.ToLower(strings.TrimSpace(day))] = true
	}

	return &participationResolver{
		mealRepo:       mealRepo,
		scheduleRepo:   scheduleRepo,
		bulkOptOutRepo: bulkOptOutRepo,
		userRepo:       userRepo,
		weekendDays:    weekendDays,
	}
}

// ResolveParticipation resolves a user's participation status for a specific date and meal type
// Priority order:
// 0. Weekend Check
// 1. Day Schedule
// 2. Explicit Participation
// 3. Bulk Opt-Out
// 4. User Default
// 5. System Default
func (r *participationResolver) ResolveParticipation(userID, date, mealType string) (bool, string, error) {
	// Priority 0: Check if date is a weekend
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false, "", fmt.Errorf("invalid date format: %w", err)
	}

	weekdayName := strings.ToLower(parsedDate.Weekday().String())
	if r.weekendDays[weekdayName] {
		// Check if there's a day schedule that overrides the weekend
		schedule, err := r.scheduleRepo.FindByDate(date)
		if err != nil {
			return false, "", err
		}

		// If there's a schedule with normal or celebration status, allow meals
		if schedule != nil && (schedule.DayStatus == models.DayStatusNormal || schedule.DayStatus == models.DayStatusCelebration) {
			// Continue to next priority checks
		} else {
			// Weekend with no override schedule
			return false, "weekend", nil
		}
	}

	// Priority 1: Check day schedule
	schedule, err := r.scheduleRepo.FindByDate(date)
	if err != nil {
		return false, "", err
	}

	if schedule != nil {
	    if schedule.DayStatus == models.DayStatusOfficeClosed || 
	       (schedule.DayStatus == models.DayStatusGovtHoliday && schedule.AvailableMeals == nil) {
	        return false, "day_schedule", nil
	    }
	}

	// Priority 2: Check explicit participation record
	participation, err := r.mealRepo.FindByUserDateMeal(userID, date, mealType)
	if err != nil {
		return false, "", err
	}

	if participation != nil {
		return participation.IsParticipating, "explicit", nil
	}

	// Priority 3: Check bulk opt-outs
	bulkOptOuts, err := r.bulkOptOutRepo.FindActiveByUserAndDate(userID, date)
	if err != nil {
		return false, "", err
	}

	for _, optOut := range bulkOptOuts {
		if string(optOut.MealType) == mealType {
			return false, "bulk_opt_out", nil
		}
	}

	// Priority 4: Check user's default preference
	user, err := r.userRepo.FindByID(userID)
	if err != nil {
		return false, "", err
	}

	if user.DefaultMealPreference == "opt_out" {
		return false, "user_default", nil
	}

	// Priority 5: System default (opt-in)
	return true, "system_default", nil
}

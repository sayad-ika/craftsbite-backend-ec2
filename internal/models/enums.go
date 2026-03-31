package models

// MealType represents the type of meal
type MealType string

const (
	MealTypeLunch          MealType = "lunch"
	MealTypeSnacks         MealType = "snacks"
	MealTypeIftar          MealType = "iftar"
	MealTypeEventDinner    MealType = "event_dinner"
	MealTypeOptionalDinner MealType = "optional_dinner"
)

// IsValid checks if the meal type is valid
func (m MealType) IsValid() bool {
	switch m {
	case MealTypeLunch, MealTypeSnacks, MealTypeIftar, MealTypeEventDinner, MealTypeOptionalDinner:
		return true
	}
	return false
}

// String returns the string representation of the meal type
func (m MealType) String() string {
	return string(m)
}

// DayStatus represents the status of a day
type DayStatus string

const (
	DayStatusNormal       DayStatus = "normal"
	DayStatusOfficeClosed DayStatus = "office_closed"
	DayStatusGovtHoliday  DayStatus = "govt_holiday"
	DayStatusCelebration  DayStatus = "celebration"
	DayStatusWeekend      DayStatus = "weekend"
    DayStatusEventDay     DayStatus = "event_day"
)

// IsValid checks if the day status is valid
func (d DayStatus) IsValid() bool {
	switch d {
	case DayStatusNormal, DayStatusOfficeClosed, DayStatusGovtHoliday, DayStatusCelebration, DayStatusWeekend, DayStatusEventDay:
		return true
	}
	return false
}

// String returns the string representation of the day status
func (d DayStatus) String() string {
	return string(d)
}

// HistoryAction represents an action in the participation history
type HistoryAction string

const (
	HistoryActionOptedIn     HistoryAction = "opted_in"
	HistoryActionOptedOut    HistoryAction = "opted_out"
	HistoryActionOverrideIn  HistoryAction = "override_in"
	HistoryActionOverrideOut HistoryAction = "override_out"
)

// IsValid checks if the history action is valid
func (h HistoryAction) IsValid() bool {
	switch h {
	case HistoryActionOptedIn, HistoryActionOptedOut, HistoryActionOverrideIn, HistoryActionOverrideOut:
		return true
	}
	return false
}

// String returns the string representation of the history action
func (h HistoryAction) String() string {
	return string(h)
}

type WorkLocationType string

const (
	WorkLocationOffice WorkLocationType = "office"
	WorkLocationWFH    WorkLocationType = "wfh"
)

func (w WorkLocationType) IsValid() bool {
	return w == WorkLocationOffice || w == WorkLocationWFH
}

func (w WorkLocationType) String() string {
	return string(w)
}

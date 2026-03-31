package models

import (
	"time"

	"github.com/google/uuid"
)

// DaySchedule represents the schedule configuration for a specific day
type DaySchedule struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Date           string     `gorm:"type:date;not null;unique" json:"date" validate:"required"`
	DayStatus      DayStatus  `gorm:"type:varchar(50);not null;default:'normal'" json:"day_status"`
	Reason         *string    `gorm:"type:text" json:"reason,omitempty"`
	AvailableMeals *string    `gorm:"type:text" json:"available_meals,omitempty"`
	CreatedBy      *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName specifies the table name for GORM
func (DaySchedule) TableName() string {
	return "day_schedules"
}

package models

import (
	"time"

	"github.com/google/uuid"
)

// MealParticipation represents a user's participation for a specific meal on a specific date
type MealParticipation struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index:idx_meal_participations_user_date" json:"user_id" validate:"required"`
	Date            string     `gorm:"type:date;not null;index:idx_meal_participations_date;index:idx_meal_participations_user_date" json:"date" validate:"required"`
	MealType        MealType   `gorm:"type:varchar(50);not null;uniqueIndex:unique_user_date_meal" json:"meal_type" validate:"required"`
	IsParticipating bool       `gorm:"not null;default:true" json:"is_participating"`
	OptedOutAt      *time.Time `gorm:"type:timestamp with time zone" json:"opted_out_at,omitempty"`
	OverrideBy      *uuid.UUID `gorm:"type:uuid" json:"override_by,omitempty"`
	OverrideReason  *string    `gorm:"type:text" json:"override_reason,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User         User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	OverrideUser *User `gorm:"foreignKey:OverrideBy" json:"override_user,omitempty"`
}

// TableName specifies the table name for GORM
func (MealParticipation) TableName() string {
	return "meal_participations"
}

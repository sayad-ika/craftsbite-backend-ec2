package models

import (
	"time"

	"github.com/google/uuid"
)

// MealParticipationHistory represents the audit trail for meal participation changes
type MealParticipationHistory struct {
	ID              uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID     `gorm:"type:uuid;not null;index:idx_history_user_date" json:"user_id" validate:"required"`
	Date            string        `gorm:"type:date;not null;index:idx_history_user_date" json:"date" validate:"required"`
	MealType        MealType      `gorm:"type:varchar(50);not null" json:"meal_type" validate:"required"`
	Action          HistoryAction `gorm:"type:varchar(20);not null" json:"action" validate:"required"`
	PreviousValue   *string       `gorm:"type:varchar(20)" json:"previous_value,omitempty"`
	ChangedByUserID *uuid.UUID    `gorm:"type:uuid" json:"changed_by_user_id,omitempty"`
	Reason          *string       `gorm:"type:varchar(255)" json:"reason,omitempty"`
	IPAddress       *string       `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	CreatedAt       time.Time     `gorm:"autoCreateTime;index:idx_history_created_at" json:"created_at"`

	// Relationships
	User      User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	ChangedBy *User `gorm:"foreignKey:ChangedByUserID;constraint:OnDelete:SET NULL" json:"changed_by,omitempty"`
}

// TableName specifies the table name for GORM
func (MealParticipationHistory) TableName() string {
	return "meal_participation_history"
}

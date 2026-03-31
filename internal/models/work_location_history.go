package models

import (
	"time"

	"github.com/google/uuid"
)

type WorkLocationHistory struct {
	ID               uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID        `gorm:"type:uuid;not null;index:idx_history_user_date" json:"user_id" validate:"required"`
	Date             string           `gorm:"type:date;not null;index:idx_history_user_date" json:"date" validate:"required"`
	Location         WorkLocationType `gorm:"type:varchar(50);not null" json:"location" validate:"required"`
	Action           HistoryAction    `gorm:"type:varchar(20);not null" json:"action" validate:"required"`
	PreviousLocation *string          `gorm:"type:varchar(20)" json:"previous_value,omitempty"`
	OverrideBy       *uuid.UUID       `gorm:"type:uuid" json:"override_by,omitempty"`
	OverrideReason   *string          `gorm:"type:varchar(255)" json:"override_reason,omitempty"`
	CreatedAt        time.Time        `gorm:"autoCreateTime;index:idx_history_created_at" json:"created_at"`

	User             User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	OverrideByUserID *User `gorm:"foreignKey:OverrideBy;constraint:OnDelete:SET NULL" json:"override_by,omitempty"`
}

func (WorkLocationHistory) TableName() string {
	return "work_location_history"
}

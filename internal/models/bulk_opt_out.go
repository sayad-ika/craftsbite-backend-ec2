package models

import (
	"time"

	"github.com/google/uuid"
)

// BulkOptOut represents a date-range opt-out for a user
type BulkOptOut struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_bulk_user_dates" json:"user_id" validate:"required"`
	StartDate string    `gorm:"type:date;not null;index:idx_bulk_user_dates" json:"start_date" validate:"required"`
	EndDate   string    `gorm:"type:date;not null;index:idx_bulk_user_dates" json:"end_date" validate:"required"`
	MealType  MealType  `gorm:"type:varchar(50);not null" json:"meal_type" validate:"required"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
    OverrideBy *uuid.UUID `gorm:"column:override_by;type:uuid" json:"override_by"`
    OverrideReason         string     `gorm:"type:text" json:"override_reason"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	OverrideUser User `gorm:"foreignKey:OverrideBy;constraint:OnDelete:SET NULL" json:"override_user"`
}

// TableName specifies the table name for GORM
func (BulkOptOut) TableName() string {
	return "bulk_opt_outs"
}

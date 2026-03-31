package models

import (
	"time"

	"github.com/google/uuid"
)

// WFHPeriod represents a company-wide Work From Home period
type WFHPeriod struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	StartDate string    `gorm:"type:date;not null" json:"start_date"`
	EndDate   string    `gorm:"type:date;not null" json:"end_date"`
	Reason    *string   `gorm:"type:text" json:"reason,omitempty"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	Active    bool      `gorm:"not null;default:true" json:"active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (WFHPeriod) TableName() string {
	return "wfh_periods"
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type WorkLocation struct {
	ID        uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:uq_work_location_user_date" json:"user_id"`
	Date      string           `gorm:"type:date;not null;uniqueIndex:uq_work_location_user_date" json:"date"`
	Location  WorkLocationType `gorm:"type:varchar(20);not null;default:'office'" json:"location"`
	SetBy     *uuid.UUID       `gorm:"type:uuid" json:"set_by,omitempty"`
	Reason    *string          `gorm:"type:text" json:"reason,omitempty"`
	CreatedAt time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time        `gorm:"autoUpdateTime" json:"updated_at"`

	User   User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Setter *User `gorm:"foreignKey:SetBy" json:"setter,omitempty"`
}

func (WorkLocation) TableName() string {
	return "work_locations"
}

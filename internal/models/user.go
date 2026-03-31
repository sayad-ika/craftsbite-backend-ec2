package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents an employee in the system
type User struct {
	ID                    uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email                 string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email" validate:"required,email"`
	Name                  string    `gorm:"type:varchar(255);not null" json:"name" validate:"required"`
	Password              string    `gorm:"type:text;not null" json:"-"`
	Role                  Role      `gorm:"type:varchar(50);not null;default:'employee'" json:"role" validate:"required"`
	Active                bool      `gorm:"not null;default:true" json:"active"`
	DefaultMealPreference string    `gorm:"type:varchar(20);not null;default:'opt_in'" json:"default_meal_preference"`
	CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Teams []Team `gorm:"many2many:team_members" json:"teams,omitempty"` // Many-to-many: user can belong to multiple teams
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

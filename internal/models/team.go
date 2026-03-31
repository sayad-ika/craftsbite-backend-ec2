package models

import (
	"time"

	"github.com/google/uuid"
)

// Team represents a team within the organization
type Team struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name" validate:"required"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	TeamLeadID  uuid.UUID `gorm:"type:uuid;not null;index" json:"team_lead_id" validate:"required"`
	Active      bool      `gorm:"not null;default:true" json:"active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	TeamLead *User  `gorm:"foreignKey:TeamLeadID;constraint:OnDelete:CASCADE" json:"team_lead,omitempty"`
	Members  []User `gorm:"many2many:team_members" json:"members,omitempty"`
}

// TableName specifies the table name for GORM
func (Team) TableName() string {
	return "teams"
}

// TeamMember represents the junction table for team membership
type TeamMember struct {
	TeamID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"team_id"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	JoinedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"joined_at"`

	// Relationships
	Team *Team `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"-"`
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName specifies the table name for GORM
func (TeamMember) TableName() string {
	return "team_members"
}

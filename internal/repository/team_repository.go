package repository

import (
	"craftsbite-backend/internal/models"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TeamRepository defines the interface for team data access
type TeamRepository interface {
	Create(team *models.Team) error
	FindByID(id string) (*models.Team, error)
	FindByTeamLeadID(teamLeadID string) ([]models.Team, error)
	Update(team *models.Team) error
	Delete(id string) error
	FindAll() ([]models.Team, error)
	AddMember(teamID, userID string) error
	RemoveMember(teamID, userID string) error
	GetTeamMembers(teamID string) ([]models.User, error)
	IsTeamMember(teamID, userID string) (bool, error)
	IsUserInAnyTeamLedBy(teamLeadID, userID string) (bool, error)
	FindTeamByUserId(userID string) (*models.Team, error)
	FindAllWithMembers() ([]models.Team, error)
}

// teamRepository implements TeamRepository
type teamRepository struct {
	db *gorm.DB
}

// NewTeamRepository creates a new team repository
func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

// Create creates a new team
func (r *teamRepository) Create(team *models.Team) error {
	if err := r.db.Create(team).Error; err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}
	return nil
}

// FindByID finds a team by ID with team lead and members preloaded
func (r *teamRepository) FindByID(id string) (*models.Team, error) {
	var team models.Team
	if err := r.db.Preload("TeamLead").Preload("Members").Where("id = ?", id).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to find team: %w", err)
	}
	return &team, nil
}

// FindByTeamLeadID finds all teams led by a specific team lead
func (r *teamRepository) FindByTeamLeadID(teamLeadID string) ([]models.Team, error) {
	var teams []models.Team
	if err := r.db.Preload("Members").Where("team_lead_id = ? AND active = ?", teamLeadID, true).Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("failed to find teams by team lead: %w", err)
	}
	return teams, nil
}

// Update updates a team
func (r *teamRepository) Update(team *models.Team) error {
	if err := r.db.Save(team).Error; err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}
	return nil
}

// Delete soft deletes a team by setting active to false
func (r *teamRepository) Delete(id string) error {
	if err := r.db.Model(&models.Team{}).Where("id = ?", id).Update("active", false).Error; err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	return nil
}

// FindAll finds all active teams
func (r *teamRepository) FindAll() ([]models.Team, error) {
	var teams []models.Team
	if err := r.db.Preload("TeamLead").Where("active = ?", true).Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("failed to find teams: %w", err)
	}
	return teams, nil
}

// AddMember adds a user to a team
func (r *teamRepository) AddMember(teamID, userID string) error {
	teamUUID, err := uuid.Parse(teamID)
	if err != nil {
		return fmt.Errorf("invalid team ID: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	teamMember := &models.TeamMember{
		TeamID: teamUUID,
		UserID: userUUID,
	}

	if err := r.db.Create(teamMember).Error; err != nil {
		return fmt.Errorf("failed to add team member: %w", err)
	}
	return nil
}

// RemoveMember removes a user from a team
func (r *teamRepository) RemoveMember(teamID, userID string) error {
	if err := r.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&models.TeamMember{}).Error; err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}
	return nil
}

// GetTeamMembers returns all users who are members of a specific team
func (r *teamRepository) GetTeamMembers(teamID string) ([]models.User, error) {
	var users []models.User
	if err := r.db.Joins("JOIN team_members ON team_members.user_id = users.id").
		Where("team_members.team_id = ? AND users.active = ?", teamID, true).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to find team members: %w", err)
	}
	return users, nil
}

// IsTeamMember checks if a user is a member of a specific team
func (r *teamRepository) IsTeamMember(teamID, userID string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check team membership: %w", err)
	}
	return count > 0, nil
}

// IsUserInAnyTeamLedBy checks if a user is a member of ANY team led by the given team lead
func (r *teamRepository) IsUserInAnyTeamLedBy(teamLeadID, userID string) (bool, error) {
	var count int64
	if err := r.db.Table("team_members").
		Joins("JOIN teams ON teams.id = team_members.team_id").
		Where("teams.team_lead_id = ? AND team_members.user_id = ? AND teams.active = ?", teamLeadID, userID, true).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check team lead membership: %w", err)
	}
	return count > 0, nil
}


func (r *teamRepository) FindTeamByUserId(userID string) (*models.Team, error) {
    var team models.Team
    if err := r.db.
        Joins("JOIN team_members ON team_members.team_id = teams.id").
        Preload("TeamLead").
        Where("team_members.user_id = ? AND teams.active = ?", userID, true).
        First(&team).Error; err != nil {
        return nil, fmt.Errorf("failed to find team by user ID: %w", err)
    }
    return &team, nil
}

func (r *teamRepository) FindAllWithMembers() ([]models.Team, error) {
	var teams []models.Team
	if err := r.db.Preload("TeamLead").Preload("Members").Where("active = ?", true).Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("failed to find teams: %w", err)
	}
	return teams, nil
}

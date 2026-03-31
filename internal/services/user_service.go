package services

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"craftsbite-backend/internal/utils"
	"fmt"

	"github.com/google/uuid"
)

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	Email                 string      `json:"email" validate:"required,email"`
	Name                  string      `json:"name" validate:"required"`
	Password              string      `json:"password" validate:"required,min=8"`
	Role                  models.Role `json:"role" validate:"required"`
	DefaultMealPreference string      `json:"default_meal_preference"`
}

// UpdateUserInput represents input for updating a user
type UpdateUserInput struct {
	Name                  *string      `json:"name"`
	Role                  *models.Role `json:"role"`
	DefaultMealPreference *string      `json:"default_meal_preference"`
	Password              *string      `json:"password" validate:"omitempty,min=8"`
}

// TeamMemberResponse represents a single team member in the response
type TeamMemberResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
}

// TeamMembersResponse represents the response for the team members endpoint
type TeamMembersResponse struct {
	TeamLeadID   string               `json:"team_lead_id"`
	TeamLeadName string               `json:"team_lead_name"`
	TotalMembers int                  `json:"total_members"`
	Members      []TeamMemberResponse `json:"members"`
}

type MyTeamResponse struct {
    TeamID       string `json:"team_id"`
    TeamName     string `json:"team_name"`
    Description  string `json:"description,omitempty"`
    TeamLeadName string `json:"team_lead_name"`
}

// UserService defines the interface for user management operations
type UserService interface {
	CreateUser(input CreateUserInput) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	UpdateUser(id string, input UpdateUserInput) (*models.User, error)
	DeactivateUser(id string) error
	ListUsers(filters map[string]interface{}) ([]models.User, error)
	GetMyTeamMembers(teamLeadID string) (*TeamMembersResponse, error)
	GetMyTeam(userID string) (*MyTeamResponse, error)
}

// userService implements UserService
type userService struct {
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, teamRepo repository.TeamRepository) UserService {
	return &userService{userRepo: userRepo, teamRepo: teamRepo}
}

// CreateUser creates a new user
func (s *userService) CreateUser(input CreateUserInput) (*models.User, error) {
	// Check if email already exists
	existingUser, _ := s.userRepo.FindByEmail(input.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default meal preference if not provided
	if input.DefaultMealPreference == "" {
		input.DefaultMealPreference = "opt_in"
	}

	// Create user
	user := &models.User{
		ID:                    uuid.New(),
		Email:                 input.Email,
		Name:                  input.Name,
		Password:              hashedPassword,
		Role:                  input.Role,
		Active:                true,
		DefaultMealPreference: input.DefaultMealPreference,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(id string) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUser updates a user
func (s *userService) UpdateUser(id string, input UpdateUserInput) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.DefaultMealPreference != nil {
		user.DefaultMealPreference = *input.DefaultMealPreference
	}
	if input.Password != nil {
		hashedPassword, err := utils.HashPassword(*input.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeactivateUser deactivates a user
func (s *userService) DeactivateUser(id string) error {
	return s.userRepo.Delete(id)
}

// ListUsers lists all users with optional filters
func (s *userService) ListUsers(filters map[string]interface{}) ([]models.User, error) {
	return s.userRepo.FindAll(filters)
}

// GetMyTeamMembers returns all members of teams led by the given team lead
func (s *userService) GetMyTeamMembers(teamLeadID string) (*TeamMembersResponse, error) {
	// Verify the user exists and is a team lead
	teamLead, err := s.userRepo.FindByID(teamLeadID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team lead: %w", err)
	}

	if teamLead.Role != models.RoleTeamLead {
		return nil, fmt.Errorf("user is not a team lead")
	}

	// Get all teams led by this user (members are preloaded)
	teams, err := s.teamRepo.FindByTeamLeadID(teamLeadID)
	if err != nil {
		return nil, fmt.Errorf("failed to find teams: %w", err)
	}

	// Aggregate members across all teams, avoid duplicates
	seen := make(map[string]bool)
	var members []TeamMemberResponse

	for _, team := range teams {
		for _, member := range team.Members {
			memberID := member.ID.String()
			if seen[memberID] {
				continue
			}
			seen[memberID] = true

			members = append(members, TeamMemberResponse{
				ID:       memberID,
				Name:     member.Name,
				Email:    member.Email,
				Role:     string(member.Role),
				TeamID:   team.ID.String(),
				TeamName: team.Name,
			})
		}
	}

	return &TeamMembersResponse{
		TeamLeadID:   teamLead.ID.String(),
		TeamLeadName: teamLead.Name,
		TotalMembers: len(members),
		Members:      members,
	}, nil
}

func (s *userService) GetMyTeam(userID string) (*MyTeamResponse, error) {
    team, err := s.teamRepo.FindTeamByUserId(userID)
    if err != nil {
        return nil, fmt.Errorf("user does not belong to any team")
    }

    r := &MyTeamResponse{
        TeamID:      team.ID.String(),
        TeamName:    team.Name,
        Description: team.Description,
    }
    if team.TeamLead != nil {
        r.TeamLeadName = team.TeamLead.Name
    }
    return r, nil
}

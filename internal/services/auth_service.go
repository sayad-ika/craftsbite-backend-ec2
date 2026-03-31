package services

import (
	"craftsbite-backend/internal/config"
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"craftsbite-backend/internal/utils"
	"fmt"
	"time"
)

// LoginResponse represents the response after successful login
type LoginResponse struct {
	Token     string       `json:"token,omitempty"`
	User      *models.User `json:"user"`
	ExpiresAt time.Time    `json:"expires_at"`
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	Login(email, password string) (*LoginResponse, error)
	GetCurrentUser(userID string) (*models.User, error)
}

// authService implements AuthService
type authService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		config:   cfg,
	}
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(email, password string) (*LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.Active {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Verify password
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	expiresAt := time.Now().Add(s.config.JWT.Expiration)
	token, err := utils.GenerateToken(
		user.ID.String(),
		user.Email,
		user.Role.String(),
		s.config.JWT.Secret,
		s.config.JWT.Expiration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:     token,
		User:      user,
		ExpiresAt: expiresAt,
	}, nil
}

// GetCurrentUser retrieves the current user by ID
func (s *authService) GetCurrentUser(userID string) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.Active {
		return nil, fmt.Errorf("user account is deactivated")
	}

	return user, nil
}

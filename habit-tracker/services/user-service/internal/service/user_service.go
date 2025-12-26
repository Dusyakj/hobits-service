package service

import (
	"context"
	"fmt"
	"time"

	"user-service/internal/domain/entity"
	"user-service/internal/domain/repository"
	"user-service/internal/domain/service"
	"user-service/pkg/hash"

	"github.com/google/uuid"
)

// userService implements service.UserService
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository) service.UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with hashed password
func (s *userService) CreateUser(ctx context.Context, userCreate *entity.UserCreate) (*entity.User, error) {
	emailExists, err := s.userRepo.EmailExists(ctx, userCreate.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if emailExists {
		return nil, fmt.Errorf("email already registered")
	}

	usernameExists, err := s.userRepo.UsernameExists(ctx, userCreate.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if usernameExists {
		return nil, fmt.Errorf("username already taken")
	}

	passwordHash, err := hash.HashPassword(userCreate.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := &entity.User{
		ID:            uuid.New(),
		Email:         userCreate.Email,
		Username:      userCreate.Username,
		PasswordHash:  passwordHash,
		FirstName:     userCreate.FirstName,
		IsActive:      true,
		EmailVerified: false,
		Timezone:      userCreate.Timezone,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateUser updates user information
func (s *userService) UpdateUser(ctx context.Context, userID uuid.UUID, userUpdate *entity.UserUpdate) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if userUpdate.FirstName != nil {
		user.FirstName = userUpdate.FirstName
	}
	if userUpdate.Timezone != nil {
		user.Timezone = *userUpdate.Timezone
	}
	if userUpdate.EmailVerified != nil {
		user.EmailVerified = *userUpdate.EmailVerified
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// ChangePassword changes user password
func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := hash.ComparePassword(user.PasswordHash, oldPassword); err != nil {
		return fmt.Errorf("invalid current password")
	}

	newPasswordHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// UpdatePassword updates user password without checking old password (for password reset)
func (s *userService) UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	newPasswordHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// VerifyEmail marks user email as verified
func (s *userService) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.UpdateEmailVerified(ctx, userID, true); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}
	return nil
}

// DeactivateUser deactivates a user account
func (s *userService) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}
	return nil
}

// ValidatePassword validates user password
func (s *userService) ValidatePassword(ctx context.Context, user *entity.User, password string) error {
	if err := hash.ComparePassword(user.PasswordHash, password); err != nil {
		return fmt.Errorf("invalid password")
	}
	return nil
}

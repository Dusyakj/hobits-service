package service

import (
	"context"

	"user-service/internal/domain/entity"

	"github.com/google/uuid"
)

// UserService defines business logic for user management
type UserService interface {
	// CreateUser creates a new user with hashed password
	CreateUser(ctx context.Context, userCreate *entity.UserCreate) (*entity.User, error)

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)

	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)

	// UpdateUser updates user information
	UpdateUser(ctx context.Context, userID uuid.UUID, userUpdate *entity.UserUpdate) (*entity.User, error)

	// ChangePassword changes user password
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error

	// UpdatePassword updates user password without checking old password (for password reset)
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error

	// VerifyEmail marks user email as verified
	VerifyEmail(ctx context.Context, userID uuid.UUID) error

	// DeactivateUser deactivates a user account
	DeactivateUser(ctx context.Context, userID uuid.UUID) error

	// ValidatePassword validates user password
	ValidatePassword(ctx context.Context, user *entity.User, password string) error
}

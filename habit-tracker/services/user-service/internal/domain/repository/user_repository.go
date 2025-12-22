package repository

import (
	"context"

	"user-service/internal/domain/entity"

	"github.com/google/uuid"
)

// UserRepository defines methods for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*entity.User, error)

	// GetByEmailOrUsername retrieves a user by email or username
	GetByEmailOrUsername(ctx context.Context, emailOrUsername string) (*entity.User, error)

	// Update updates user information
	Update(ctx context.Context, user *entity.User) error

	// UpdatePassword updates user password
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error

	// UpdateEmailVerified updates email verification status
	UpdateEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error

	// Delete deletes a user (soft delete by setting is_active to false)
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists checks if a user exists by ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// EmailExists checks if email is already taken
	EmailExists(ctx context.Context, email string) (bool, error)

	// UsernameExists checks if username is already taken
	UsernameExists(ctx context.Context, username string) (bool, error)
}

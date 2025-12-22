package repository

import (
	"context"
	"user-service/internal/domain/entity"

	"github.com/google/uuid"
)

// SessionRepository defines methods for session data access
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *entity.Session) error

	// GetByID retrieves a session by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Session, error)

	// GetByTokenHash retrieves a session by token hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error)

	// GetByUserID retrieves all sessions for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error)

	// GetActiveByUserID retrieves all active (non-expired) sessions for a user
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error)

	// UpdateLastActivity updates the last activity timestamp
	UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error

	// Delete deletes a session by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserID deletes all sessions for a user
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// DeleteExpired deletes all expired sessions
	DeleteExpired(ctx context.Context) (int64, error)

	// CountActiveByUserID counts active sessions for a user
	CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// Exists checks if a session exists by ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

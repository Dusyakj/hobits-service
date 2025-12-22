package repository

import (
	"context"
	"habits-service/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

// HabitConfirmationRepository defines the interface for habit confirmation persistence
type HabitConfirmationRepository interface {
	// Create creates a new habit confirmation
	Create(ctx context.Context, confirmation *entity.HabitConfirmation) error

	// GetByHabitID retrieves confirmations for a habit with pagination
	GetByHabitID(ctx context.Context, habitID uuid.UUID, limit, offset int32) ([]*entity.HabitConfirmation, error)

	// CountByHabitID returns the total count of confirmations for a habit
	CountByHabitID(ctx context.Context, habitID uuid.UUID) (int32, error)

	// GetLatestByHabitID retrieves the most recent confirmation for a habit
	GetLatestByHabitID(ctx context.Context, habitID uuid.UUID) (*entity.HabitConfirmation, error)

	// ExistsForDate checks if a confirmation exists for a habit on a specific date
	ExistsForDate(ctx context.Context, habitID uuid.UUID, date string) (bool, error)

	// GetStats retrieves statistics for a habit
	GetStats(ctx context.Context, habitID uuid.UUID) (*HabitStats, error)
}

// HabitStats represents habit statistics
type HabitStats struct {
	CurrentStreak        int32
	LongestStreak        int32
	TotalConfirmations   int32
	CompletionRate       float64
	FirstConfirmation    *time.Time
	LastConfirmation     *time.Time
}

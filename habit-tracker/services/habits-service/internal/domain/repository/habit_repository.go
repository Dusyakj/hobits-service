package repository

import (
	"context"
	"habits-service/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

// HabitRepository defines the interface for habit persistence
type HabitRepository interface {
	// Create creates a new habit
	Create(ctx context.Context, habit *entity.Habit) error

	// GetByID retrieves a habit by ID
	GetByID(ctx context.Context, habitID uuid.UUID) (*entity.Habit, error)

	// GetByUserID retrieves all habits for a user
	GetByUserID(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]*entity.Habit, error)

	// Update updates a habit
	Update(ctx context.Context, habit *entity.Habit) error

	// Delete soft deletes a habit (sets is_active = false)
	Delete(ctx context.Context, habitID uuid.UUID) error

	// UpdateStreakAndDeadline updates the streak and next deadline for a habit
	UpdateStreakAndDeadline(ctx context.Context, habitID uuid.UUID, streak int32, nextDeadline time.Time, confirmed bool) error

	// GetHabitsWithMissedDeadlines retrieves habits that have passed their deadline and haven't been confirmed
	GetHabitsWithMissedDeadlines(ctx context.Context) ([]*entity.Habit, error)

	// GetHabitsToResetConfirmation retrieves confirmed habits where deadline is within time window
	// (to reset confirmation flag for new period)
	GetHabitsToResetConfirmation(ctx context.Context, fromTime, toTime time.Time) ([]*entity.Habit, error)

	// ResetConfirmationFlag resets the confirmed_for_current_period flag for a habit
	ResetConfirmationFlag(ctx context.Context, habitID uuid.UUID) error

	// GetByIDAndUserID retrieves a habit by ID and user ID (for authorization)
	GetByIDAndUserID(ctx context.Context, habitID, userID uuid.UUID) (*entity.Habit, error)
}

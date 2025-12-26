package service

import (
	"context"
	"habits-service/internal/domain/entity"
	"habits-service/internal/domain/repository"

	"github.com/google/uuid"
)

// HabitService defines the interface for habit business logic
type HabitService interface {
	// CreateHabit creates a new habit with timezone conversion
	CreateHabit(ctx context.Context, userID uuid.UUID, name string, description, color *string,
		scheduleType entity.ScheduleType, intervalDays *int32, weeklyDays []int32, timezone string) (*entity.Habit, error)

	// GetHabit retrieves a habit by ID
	GetHabit(ctx context.Context, habitID, userID uuid.UUID) (*entity.Habit, error)

	// ListHabits retrieves all habits for a user
	ListHabits(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]*entity.Habit, int32, error)

	// UpdateHabit updates a habit
	UpdateHabit(ctx context.Context, habitID, userID uuid.UUID, name *string, description, color *string,
		scheduleType *entity.ScheduleType, intervalDays *int32, weeklyDays []int32, timezone *string) (*entity.Habit, error)

	// DeleteHabit soft deletes a habit
	DeleteHabit(ctx context.Context, habitID, userID uuid.UUID) error

	// ConfirmHabit confirms habit completion for the current period
	ConfirmHabit(ctx context.Context, habitID, userID uuid.UUID, notes *string) (*entity.Habit, *entity.HabitConfirmation, error)

	// GetHabitHistory retrieves confirmation history for a habit
	GetHabitHistory(ctx context.Context, habitID, userID uuid.UUID, limit, offset int32) ([]*entity.HabitConfirmation, int32, error)

	// GetHabitStats retrieves statistics for a habit
	GetHabitStats(ctx context.Context, habitID, userID uuid.UUID) (*repository.HabitStats, error)

	// ProcessMissedDeadlines checks for missed deadlines and resets streaks
	ProcessMissedDeadlines(ctx context.Context) error

	// ResetConfirmationFlags resets confirmation flags for habits entering new period
	ResetConfirmationFlags(ctx context.Context) error

	// ProcessExpiredConfirmedDeadlines moves confirmed habits with expired deadlines to next period
	ProcessExpiredConfirmedDeadlines(ctx context.Context) error
}

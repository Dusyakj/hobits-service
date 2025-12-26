package postgres

import (
	"context"
	"fmt"
	"habits-service/internal/domain/entity"
	"habits-service/internal/domain/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type habitRepository struct {
	pool *pgxpool.Pool
}

// NewHabitRepository creates a new PostgreSQL habit repository
func NewHabitRepository(pool *pgxpool.Pool) repository.HabitRepository {
	return &habitRepository{pool: pool}
}

func (r *habitRepository) Create(ctx context.Context, habit *entity.Habit) error {
	query := `
		INSERT INTO habits (
			id, user_id, name, description, color,
			schedule_type, interval_days, weekly_days, timezone_offset_hours,
			streak, next_deadline_utc, confirmed_for_current_period, last_confirmed_at,
			is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16
		)
	`

	_, err := r.pool.Exec(ctx, query,
		habit.ID, habit.UserID, habit.Name, habit.Description, habit.Color,
		habit.ScheduleType, habit.IntervalDays, habit.WeeklyDays, habit.TimezoneOffsetHours,
		habit.Streak, habit.NextDeadlineUTC, habit.ConfirmedForCurrentPeriod, habit.LastConfirmedAt,
		habit.IsActive, habit.CreatedAt, habit.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create habit: %w", err)
	}

	return nil
}

func (r *habitRepository) GetByID(ctx context.Context, habitID uuid.UUID) (*entity.Habit, error) {
	query := `
		SELECT
			id, user_id, name, description, color,
			schedule_type, interval_days, weekly_days, timezone_offset_hours,
			streak, next_deadline_utc, confirmed_for_current_period, last_confirmed_at,
			is_active, created_at, updated_at
		FROM habits
		WHERE id = $1
	`

	habit := &entity.Habit{}
	err := r.pool.QueryRow(ctx, query, habitID).Scan(
		&habit.ID, &habit.UserID, &habit.Name, &habit.Description, &habit.Color,
		&habit.ScheduleType, &habit.IntervalDays, &habit.WeeklyDays, &habit.TimezoneOffsetHours,
		&habit.Streak, &habit.NextDeadlineUTC, &habit.ConfirmedForCurrentPeriod, &habit.LastConfirmedAt,
		&habit.IsActive, &habit.CreatedAt, &habit.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("habit not found")
		}
		return nil, fmt.Errorf("failed to get habit: %w", err)
	}

	return habit, nil
}

func (r *habitRepository) GetByIDAndUserID(ctx context.Context, habitID, userID uuid.UUID) (*entity.Habit, error) {
	query := `
		SELECT
			id, user_id, name, description, color,
			schedule_type, interval_days, weekly_days, timezone_offset_hours,
			streak, next_deadline_utc, confirmed_for_current_period, last_confirmed_at,
			is_active, created_at, updated_at
		FROM habits
		WHERE id = $1 AND user_id = $2
	`

	habit := &entity.Habit{}
	err := r.pool.QueryRow(ctx, query, habitID, userID).Scan(
		&habit.ID, &habit.UserID, &habit.Name, &habit.Description, &habit.Color,
		&habit.ScheduleType, &habit.IntervalDays, &habit.WeeklyDays, &habit.TimezoneOffsetHours,
		&habit.Streak, &habit.NextDeadlineUTC, &habit.ConfirmedForCurrentPeriod, &habit.LastConfirmedAt,
		&habit.IsActive, &habit.CreatedAt, &habit.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("habit not found or unauthorized")
		}
		return nil, fmt.Errorf("failed to get habit: %w", err)
	}

	return habit, nil
}

func (r *habitRepository) GetByUserID(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]*entity.Habit, error) {
	query := `
		SELECT
			id, user_id, name, description, color,
			schedule_type, interval_days, weekly_days, timezone_offset_hours,
			streak, next_deadline_utc, confirmed_for_current_period, last_confirmed_at,
			is_active, created_at, updated_at
		FROM habits
		WHERE user_id = $1
	`

	if activeOnly {
		query += " AND is_active = TRUE"
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get habits: %w", err)
	}
	defer rows.Close()

	var habits []*entity.Habit
	for rows.Next() {
		habit := &entity.Habit{}
		err := rows.Scan(
			&habit.ID, &habit.UserID, &habit.Name, &habit.Description, &habit.Color,
			&habit.ScheduleType, &habit.IntervalDays, &habit.WeeklyDays, &habit.TimezoneOffsetHours,
			&habit.Streak, &habit.NextDeadlineUTC, &habit.ConfirmedForCurrentPeriod, &habit.LastConfirmedAt,
			&habit.IsActive, &habit.CreatedAt, &habit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan habit: %w", err)
		}
		habits = append(habits, habit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate habits: %w", err)
	}

	return habits, nil
}

func (r *habitRepository) Update(ctx context.Context, habit *entity.Habit) error {
	query := `
		UPDATE habits SET
			name = $1,
			description = $2,
			color = $3,
			schedule_type = $4,
			interval_days = $5,
			weekly_days = $6,
			timezone_offset_hours = $7,
			updated_at = $8
		WHERE id = $9
	`

	result, err := r.pool.Exec(ctx, query,
		habit.Name, habit.Description, habit.Color,
		habit.ScheduleType, habit.IntervalDays, habit.WeeklyDays, habit.TimezoneOffsetHours,
		time.Now().UTC(), habit.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update habit: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("habit not found")
	}

	return nil
}

func (r *habitRepository) Delete(ctx context.Context, habitID uuid.UUID) error {
	query := `
		UPDATE habits SET
			is_active = FALSE,
			updated_at = $1
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, time.Now().UTC(), habitID)
	if err != nil {
		return fmt.Errorf("failed to delete habit: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("habit not found")
	}

	return nil
}

func (r *habitRepository) UpdateStreakAndDeadline(ctx context.Context, habitID uuid.UUID, streak int32, nextDeadline time.Time, confirmed bool) error {
	query := `
		UPDATE habits SET
			streak = $1,
			next_deadline_utc = $2,
			confirmed_for_current_period = $3,
			last_confirmed_at = $4,
			updated_at = $5
		WHERE id = $6
	`

	now := time.Now().UTC()
	result, err := r.pool.Exec(ctx, query, streak, nextDeadline, confirmed, now, now, habitID)
	if err != nil {
		return fmt.Errorf("failed to update streak and deadline: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("habit not found")
	}

	return nil
}

func (r *habitRepository) GetHabitsWithMissedDeadlines(ctx context.Context) ([]*entity.Habit, error) {
	query := `
		SELECT
			id, user_id, name, description, color,
			schedule_type, interval_days, weekly_days, timezone_offset_hours,
			streak, next_deadline_utc, confirmed_for_current_period, last_confirmed_at,
			is_active, created_at, updated_at
		FROM habits
		WHERE is_active = TRUE
		  AND confirmed_for_current_period = FALSE
		  AND next_deadline_utc <= $1
		ORDER BY next_deadline_utc ASC
	`

	rows, err := r.pool.Query(ctx, query, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to get habits with missed deadlines: %w", err)
	}
	defer rows.Close()

	var habits []*entity.Habit
	for rows.Next() {
		habit := &entity.Habit{}
		err := rows.Scan(
			&habit.ID, &habit.UserID, &habit.Name, &habit.Description, &habit.Color,
			&habit.ScheduleType, &habit.IntervalDays, &habit.WeeklyDays, &habit.TimezoneOffsetHours,
			&habit.Streak, &habit.NextDeadlineUTC, &habit.ConfirmedForCurrentPeriod, &habit.LastConfirmedAt,
			&habit.IsActive, &habit.CreatedAt, &habit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan habit: %w", err)
		}
		habits = append(habits, habit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate habits: %w", err)
	}

	return habits, nil
}

func (r *habitRepository) GetHabitsToResetConfirmation(ctx context.Context, fromTime, toTime time.Time) ([]*entity.Habit, error) {
	query := `
		SELECT
			id, user_id, name, description, color,
			schedule_type, interval_days, weekly_days, timezone_offset_hours,
			streak, next_deadline_utc, confirmed_for_current_period, last_confirmed_at,
			is_active, created_at, updated_at
		FROM habits
		WHERE is_active = TRUE
		  AND confirmed_for_current_period = TRUE
		  AND next_deadline_utc >= $1
		  AND next_deadline_utc <= $2
		ORDER BY next_deadline_utc ASC
	`

	rows, err := r.pool.Query(ctx, query, fromTime, toTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get habits to reset confirmation: %w", err)
	}
	defer rows.Close()

	var habits []*entity.Habit
	for rows.Next() {
		habit := &entity.Habit{}
		err := rows.Scan(
			&habit.ID, &habit.UserID, &habit.Name, &habit.Description, &habit.Color,
			&habit.ScheduleType, &habit.IntervalDays, &habit.WeeklyDays, &habit.TimezoneOffsetHours,
			&habit.Streak, &habit.NextDeadlineUTC, &habit.ConfirmedForCurrentPeriod, &habit.LastConfirmedAt,
			&habit.IsActive, &habit.CreatedAt, &habit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan habit: %w", err)
		}
		habits = append(habits, habit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate habits: %w", err)
	}

	return habits, nil
}

func (r *habitRepository) GetConfirmedHabitsWithExpiredDeadlines(ctx context.Context, beforeTime time.Time) ([]*entity.Habit, error) {
	query := `
		SELECT
			id, user_id, name, description, color,
			schedule_type, interval_days, weekly_days, timezone_offset_hours,
			streak, next_deadline_utc, confirmed_for_current_period, last_confirmed_at,
			is_active, created_at, updated_at
		FROM habits
		WHERE is_active = TRUE
		  AND confirmed_for_current_period = TRUE
		  AND next_deadline_utc < $1
		ORDER BY next_deadline_utc ASC
	`

	rows, err := r.pool.Query(ctx, query, beforeTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get confirmed habits with expired deadlines: %w", err)
	}
	defer rows.Close()

	var habits []*entity.Habit
	for rows.Next() {
		habit := &entity.Habit{}
		err := rows.Scan(
			&habit.ID, &habit.UserID, &habit.Name, &habit.Description, &habit.Color,
			&habit.ScheduleType, &habit.IntervalDays, &habit.WeeklyDays, &habit.TimezoneOffsetHours,
			&habit.Streak, &habit.NextDeadlineUTC, &habit.ConfirmedForCurrentPeriod, &habit.LastConfirmedAt,
			&habit.IsActive, &habit.CreatedAt, &habit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan habit: %w", err)
		}
		habits = append(habits, habit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate habits: %w", err)
	}

	return habits, nil
}

func (r *habitRepository) ResetConfirmationFlag(ctx context.Context, habitID uuid.UUID) error {
	query := `
		UPDATE habits SET
			confirmed_for_current_period = FALSE,
			updated_at = $1
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, time.Now().UTC(), habitID)
	if err != nil {
		return fmt.Errorf("failed to reset confirmation flag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("habit not found")
	}

	return nil
}

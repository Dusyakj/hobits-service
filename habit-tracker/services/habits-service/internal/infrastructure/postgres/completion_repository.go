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

type habitConfirmationRepository struct {
	pool *pgxpool.Pool
}

// NewHabitConfirmationRepository creates a new PostgreSQL habit confirmation repository
func NewHabitConfirmationRepository(pool *pgxpool.Pool) repository.HabitConfirmationRepository {
	return &habitConfirmationRepository{pool: pool}
}

func (r *habitConfirmationRepository) Create(ctx context.Context, confirmation *entity.HabitConfirmation) error {
	query := `
		INSERT INTO habit_confirmations (
			id, habit_id, user_id, confirmed_at, confirmed_for_date, notes, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	_, err := r.pool.Exec(ctx, query,
		confirmation.ID,
		confirmation.HabitID,
		confirmation.UserID,
		confirmation.ConfirmedAt,
		confirmation.ConfirmedForDate,
		confirmation.Notes,
		confirmation.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create habit confirmation: %w", err)
	}

	return nil
}

func (r *habitConfirmationRepository) GetByHabitID(ctx context.Context, habitID uuid.UUID, limit, offset int32) ([]*entity.HabitConfirmation, error) {
	if limit <= 0 {
		limit = 30 // Default limit
	}

	query := `
		SELECT
			id, habit_id, user_id, confirmed_at, confirmed_for_date, notes, created_at
		FROM habit_confirmations
		WHERE habit_id = $1
		ORDER BY confirmed_for_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, habitID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get habit confirmations: %w", err)
	}
	defer rows.Close()

	var confirmations []*entity.HabitConfirmation
	for rows.Next() {
		confirmation := &entity.HabitConfirmation{}
		err := rows.Scan(
			&confirmation.ID,
			&confirmation.HabitID,
			&confirmation.UserID,
			&confirmation.ConfirmedAt,
			&confirmation.ConfirmedForDate,
			&confirmation.Notes,
			&confirmation.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan confirmation: %w", err)
		}
		confirmations = append(confirmations, confirmation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate confirmations: %w", err)
	}

	return confirmations, nil
}

func (r *habitConfirmationRepository) CountByHabitID(ctx context.Context, habitID uuid.UUID) (int32, error) {
	query := `
		SELECT COUNT(*) FROM habit_confirmations WHERE habit_id = $1
	`

	var count int32
	err := r.pool.QueryRow(ctx, query, habitID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count confirmations: %w", err)
	}

	return count, nil
}

func (r *habitConfirmationRepository) GetLatestByHabitID(ctx context.Context, habitID uuid.UUID) (*entity.HabitConfirmation, error) {
	query := `
		SELECT
			id, habit_id, user_id, confirmed_at, confirmed_for_date, notes, created_at
		FROM habit_confirmations
		WHERE habit_id = $1
		ORDER BY confirmed_for_date DESC
		LIMIT 1
	`

	confirmation := &entity.HabitConfirmation{}
	err := r.pool.QueryRow(ctx, query, habitID).Scan(
		&confirmation.ID,
		&confirmation.HabitID,
		&confirmation.UserID,
		&confirmation.ConfirmedAt,
		&confirmation.ConfirmedForDate,
		&confirmation.Notes,
		&confirmation.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No confirmations yet
		}
		return nil, fmt.Errorf("failed to get latest confirmation: %w", err)
	}

	return confirmation, nil
}

func (r *habitConfirmationRepository) ExistsForDate(ctx context.Context, habitID uuid.UUID, date string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM habit_confirmations
			WHERE habit_id = $1 AND confirmed_for_date = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, habitID, date).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check confirmation existence: %w", err)
	}

	return exists, nil
}

func (r *habitConfirmationRepository) GetStats(ctx context.Context, habitID uuid.UUID) (*repository.HabitStats, error) {
	// Get basic stats
	statsQuery := `
		SELECT
			COUNT(*) as total_confirmations,
			MIN(confirmed_at) as first_confirmation,
			MAX(confirmed_at) as last_confirmation
		FROM habit_confirmations
		WHERE habit_id = $1
	`

	stats := &repository.HabitStats{}
	var firstConfirmation, lastConfirmation *time.Time

	err := r.pool.QueryRow(ctx, statsQuery, habitID).Scan(
		&stats.TotalConfirmations,
		&firstConfirmation,
		&lastConfirmation,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic stats: %w", err)
	}

	stats.FirstConfirmation = firstConfirmation
	stats.LastConfirmation = lastConfirmation

	// Get current and longest streak
	// This is a simplified implementation - a more accurate one would need the habit's schedule
	streakQuery := `
		WITH confirmation_dates AS (
			SELECT
				confirmed_for_date,
				confirmed_for_date::date - ROW_NUMBER() OVER (ORDER BY confirmed_for_date)::int AS streak_group
			FROM habit_confirmations
			WHERE habit_id = $1
			ORDER BY confirmed_for_date
		),
		streaks AS (
			SELECT
				COUNT(*) as streak_length,
				MAX(confirmed_for_date) as last_date
			FROM confirmation_dates
			GROUP BY streak_group
		)
		SELECT
			COALESCE(MAX(streak_length), 0) as longest_streak,
			COALESCE(
				(SELECT streak_length FROM streaks WHERE last_date = (SELECT MAX(confirmed_for_date) FROM habit_confirmations WHERE habit_id = $1)),
				0
			) as current_streak
		FROM streaks
	`

	err = r.pool.QueryRow(ctx, streakQuery, habitID).Scan(
		&stats.LongestStreak,
		&stats.CurrentStreak,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get streak stats: %w", err)
	}

	// Calculate completion rate (simplified - assumes daily habit)
	// A more accurate implementation would calculate based on habit schedule
	if lastConfirmation != nil && firstConfirmation != nil {
		daysSinceStart := int32(time.Since(*firstConfirmation).Hours() / 24)
		if daysSinceStart > 0 {
			stats.CompletionRate = float64(stats.TotalConfirmations) / float64(daysSinceStart) * 100
			if stats.CompletionRate > 100 {
				stats.CompletionRate = 100
			}
		}
	}

	return stats, nil
}

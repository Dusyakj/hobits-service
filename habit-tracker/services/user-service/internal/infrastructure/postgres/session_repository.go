package postgres

import (
	"context"
	"errors"
	"fmt"

	"user-service/internal/domain/entity"
	"user-service/internal/domain/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// sessionRepository implements repository.SessionRepository
type sessionRepository struct {
	pool *pgxpool.Pool
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(pool *pgxpool.Pool) repository.SessionRepository {
	return &sessionRepository{
		pool: pool,
	}
}

// Create creates a new session
func (r *sessionRepository) Create(ctx context.Context, session *entity.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, token_hash, ip_address, user_agent, expires_at, created_at, last_activity_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.CreatedAt,
		session.LastActivityAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID retrieves a session by ID
func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Session, error) {
	query := `
		SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, created_at, last_activity_at
		FROM sessions
		WHERE id = $1
	`

	var session entity.Session
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastActivityAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}

	return &session, nil
}

// GetByTokenHash retrieves a session by token hash
func (r *sessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error) {
	query := `
		SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, created_at, last_activity_at
		FROM sessions
		WHERE token_hash = $1 AND expires_at > NOW()
	`

	var session entity.Session
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastActivityAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to get session by token hash: %w", err)
	}

	return &session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *sessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error) {
	query := `
		SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, created_at, last_activity_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user ID: %w", err)
	}
	defer rows.Close()

	var sessions []*entity.Session
	for rows.Next() {
		var session entity.Session
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.TokenHash,
			&session.IPAddress,
			&session.UserAgent,
			&session.ExpiresAt,
			&session.CreatedAt,
			&session.LastActivityAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// GetActiveByUserID retrieves all active (non-expired) sessions for a user
func (r *sessionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error) {
	query := `
		SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, created_at, last_activity_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY last_activity_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*entity.Session
	for rows.Next() {
		var session entity.Session
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.TokenHash,
			&session.IPAddress,
			&session.UserAgent,
			&session.ExpiresAt,
			&session.CreatedAt,
			&session.LastActivityAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// UpdateLastActivity updates the last activity timestamp
func (r *sessionRepository) UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE sessions
		SET last_activity_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update last activity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// Delete deletes a session by ID
func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// DeleteByUserID deletes all sessions for a user
func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired sessions
func (r *sessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`

	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return result.RowsAffected(), nil
}

// CountActiveByUserID counts active sessions for a user
func (r *sessionRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active sessions: %w", err)
	}

	return count, nil
}

// Exists checks if a session exists by ID
func (r *sessionRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM sessions WHERE id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return exists, nil
}

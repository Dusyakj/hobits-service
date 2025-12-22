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

// userRepository implements repository.UserRepository
type userRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &userRepository{
		pool: pool,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, first_name, is_active, email_verified, timezone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.FirstName,
		user.IsActive,
		user.EmailVerified,
		user.Timezone,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, is_active, email_verified, timezone, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = true
	`

	var user entity.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.IsActive,
		&user.EmailVerified,
		&user.Timezone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, is_active, email_verified, timezone, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = true
	`

	var user entity.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.IsActive,
		&user.EmailVerified,
		&user.Timezone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, is_active, email_verified, timezone, created_at, updated_at
		FROM users
		WHERE username = $1 AND is_active = true
	`

	var user entity.User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.IsActive,
		&user.EmailVerified,
		&user.Timezone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetByEmailOrUsername retrieves a user by email or username
func (r *userRepository) GetByEmailOrUsername(ctx context.Context, emailOrUsername string) (*entity.User, error) {
	query := `
		SELECT id, email, username, password_hash, first_name, is_active, email_verified, timezone, created_at, updated_at
		FROM users
		WHERE (email = $1 OR username = $1) AND is_active = true
	`

	var user entity.User
	err := r.pool.QueryRow(ctx, query, emailOrUsername).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.IsActive,
		&user.EmailVerified,
		&user.Timezone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update updates user information
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET first_name = $2, timezone = $3, email_verified = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND is_active = true
	`

	result, err := r.pool.Exec(ctx, query,
		user.ID,
		user.FirstName,
		user.Timezone,
		user.EmailVerified,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or inactive")
	}

	return nil
}

// UpdatePassword updates user password
func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND is_active = true
	`

	result, err := r.pool.Exec(ctx, query, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or inactive")
	}

	return nil
}

// UpdateEmailVerified updates email verification status
func (r *userRepository) UpdateEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	query := `
		UPDATE users
		SET email_verified = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND is_active = true
	`

	result, err := r.pool.Exec(ctx, query, userID, verified)
	if err != nil {
		return fmt.Errorf("failed to update email verification: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or inactive")
	}

	return nil
}

// Delete soft deletes a user by setting is_active to false
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Exists checks if a user exists by ID
func (r *userRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND is_active = true)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// EmailExists checks if email is already taken
func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// UsernameExists checks if username is already taken
func (r *userRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return exists, nil
}

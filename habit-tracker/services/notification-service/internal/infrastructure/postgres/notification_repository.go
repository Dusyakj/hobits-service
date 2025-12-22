package postgres

import (
	"context"
	"fmt"
	"time"

	"notification-service/internal/domain/entity"
	"notification-service/internal/domain/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type notificationRepository struct {
	db *pgxpool.Pool
}

// NewNotificationRepository creates a new PostgreSQL notification repository
func NewNotificationRepository(db *pgxpool.Pool) repository.NotificationRepository {
	return &notificationRepository{
		db: db,
	}
}

func (r *notificationRepository) Create(ctx context.Context, notification *entity.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, type, status, subject, content, recipient, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	notification.ID = uuid.New().String()
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Status,
		notification.Subject,
		notification.Content,
		notification.To,
		notification.Metadata,
		notification.CreatedAt,
		notification.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id string) (*entity.Notification, error) {
	query := `
		SELECT id, user_id, type, status, subject, content, recipient, metadata,
		       sent_at, failed_at, error, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	var notification entity.Notification
	err := r.db.QueryRow(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Status,
		&notification.Subject,
		&notification.Content,
		&notification.To,
		&notification.Metadata,
		&notification.SentAt,
		&notification.FailedAt,
		&notification.Error,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return &notification, nil
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Notification, error) {
	query := `
		SELECT id, user_id, type, status, subject, content, recipient, metadata,
		       sent_at, failed_at, error, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*entity.Notification
	for rows.Next() {
		var notification entity.Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Status,
			&notification.Subject,
			&notification.Content,
			&notification.To,
			&notification.Metadata,
			&notification.SentAt,
			&notification.FailedAt,
			&notification.Error,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

func (r *notificationRepository) UpdateStatus(ctx context.Context, id string, status entity.NotificationStatus, sentAt *string, failedAt *string, errorMsg *string) error {
	query := `
		UPDATE notifications
		SET status = $1, sent_at = $2, failed_at = $3, error = $4, updated_at = $5
		WHERE id = $6
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, status, sentAt, failedAt, errorMsg, now, id)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}

func (r *notificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]*entity.Notification, error) {
	query := `
		SELECT id, user_id, type, status, subject, content, recipient, metadata,
		       sent_at, failed_at, error, created_at, updated_at
		FROM notifications
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, entity.NotificationStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*entity.Notification
	for rows.Next() {
		var notification entity.Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Status,
			&notification.Subject,
			&notification.Content,
			&notification.To,
			&notification.Metadata,
			&notification.SentAt,
			&notification.FailedAt,
			&notification.Error,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

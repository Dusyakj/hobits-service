package repository

import (
	"context"
	"notification-service/internal/domain/entity"
)

// NotificationRepository defines the interface for notification persistence
type NotificationRepository interface {
	// Create creates a new notification record
	Create(ctx context.Context, notification *entity.Notification) error

	// GetByID retrieves a notification by ID
	GetByID(ctx context.Context, id string) (*entity.Notification, error)

	// GetByUserID retrieves all notifications for a user
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Notification, error)

	// UpdateStatus updates the status of a notification
	UpdateStatus(ctx context.Context, id string, status entity.NotificationStatus, sentAt *string, failedAt *string, errorMsg *string) error

	// GetPendingNotifications retrieves all pending notifications
	GetPendingNotifications(ctx context.Context, limit int) ([]*entity.Notification, error)
}

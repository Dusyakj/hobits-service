package service

import (
	"context"
	"notification-service/internal/domain/entity"
)

// NotificationService defines the interface for notification business logic
type NotificationService interface {
	// SendEmailVerification sends an email verification notification
	SendEmailVerification(ctx context.Context, data *entity.EmailVerificationData) error

	// SendPasswordReset sends a password reset notification
	SendPasswordReset(ctx context.Context, data *entity.PasswordResetData) error

	// GetNotificationHistory retrieves notification history for a user
	GetNotificationHistory(ctx context.Context, userID string, limit, offset int) ([]*entity.Notification, error)
}

// EmailService defines the interface for email sending
type EmailService interface {
	// SendVerificationEmail sends a verification email
	SendVerificationEmail(ctx context.Context, to, username, firstName, verificationURL string) error

	// SendPasswordResetEmail sends a password reset email
	SendPasswordResetEmail(ctx context.Context, to, username, firstName, resetURL string) error

	// SendPasswordChangedEmail sends a notification when password is changed
	SendPasswordChangedEmail(ctx context.Context, to string, wasReset bool) error
}

package service

import (
	"context"
	"fmt"
	"time"

	"notification-service/internal/domain/entity"
	"notification-service/internal/domain/repository"
	"notification-service/internal/domain/service"
)

type notificationService struct {
	repo         repository.NotificationRepository
	emailService service.EmailService
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	repo repository.NotificationRepository,
	emailService service.EmailService,
) service.NotificationService {
	return &notificationService{
		repo:         repo,
		emailService: emailService,
	}
}

func (s *notificationService) SendEmailVerification(ctx context.Context, data *entity.EmailVerificationData) error {
	// Create notification record
	notification := &entity.Notification{
		UserID:  data.UserID,
		Type:    entity.NotificationTypeEmail,
		Status:  entity.NotificationStatusPending,
		Subject: "Email Verification",
		Content: fmt.Sprintf("Verification email sent to %s", data.Email),
		To:      data.Email,
		Metadata: map[string]string{
			"verification_token": data.VerificationToken,
			"username":          data.Username,
			"first_name":        data.FirstName,
		},
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	// Send email
	err := s.emailService.SendVerificationEmail(
		ctx,
		data.Email,
		data.Username,
		data.FirstName,
		data.VerificationToken,
	)

	// Update notification status
	now := time.Now().Format(time.RFC3339)
	if err != nil {
		errMsg := err.Error()
		if updateErr := s.repo.UpdateStatus(ctx, notification.ID, entity.NotificationStatusFailed, nil, &now, &errMsg); updateErr != nil {
			return fmt.Errorf("failed to update notification status: %w", updateErr)
		}
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	if err := s.repo.UpdateStatus(ctx, notification.ID, entity.NotificationStatusSent, &now, nil, nil); err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}

func (s *notificationService) SendPasswordReset(ctx context.Context, data *entity.PasswordResetData) error {
	// Create notification record
	notification := &entity.Notification{
		UserID:  data.UserID,
		Type:    entity.NotificationTypeEmail,
		Status:  entity.NotificationStatusPending,
		Subject: "Password Reset",
		Content: fmt.Sprintf("Password reset email sent to %s", data.Email),
		To:      data.Email,
		Metadata: map[string]string{
			"reset_token": data.ResetToken,
			"username":    data.Username,
			"first_name":  data.FirstName,
		},
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	// Send email
	err := s.emailService.SendPasswordResetEmail(
		ctx,
		data.Email,
		data.Username,
		data.FirstName,
		data.ResetToken,
	)

	// Update notification status
	now := time.Now().Format(time.RFC3339)
	if err != nil {
		errMsg := err.Error()
		if updateErr := s.repo.UpdateStatus(ctx, notification.ID, entity.NotificationStatusFailed, nil, &now, &errMsg); updateErr != nil {
			return fmt.Errorf("failed to update notification status: %w", updateErr)
		}
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	if err := s.repo.UpdateStatus(ctx, notification.ID, entity.NotificationStatusSent, &now, nil, nil); err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}

func (s *notificationService) GetNotificationHistory(ctx context.Context, userID string, limit, offset int) ([]*entity.Notification, error) {
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

package service

import (
	"context"

	"notification-service/internal/domain/service"
	"notification-service/internal/infrastructure/smtp"
)

type emailService struct {
	smtpClient *smtp.Client
}

// NewEmailService creates a new email service
func NewEmailService(smtpClient *smtp.Client) service.EmailService {
	return &emailService{
		smtpClient: smtpClient,
	}
}

func (s *emailService) SendVerificationEmail(ctx context.Context, to, username, firstName, verificationToken string) error {
	return s.smtpClient.SendVerificationEmail(ctx, to, username, firstName, verificationToken)
}

func (s *emailService) SendPasswordResetEmail(ctx context.Context, to, username, firstName, resetToken string) error {
	return s.smtpClient.SendPasswordResetEmail(ctx, to, username, firstName, resetToken)
}

func (s *emailService) SendPasswordChangedEmail(ctx context.Context, to string, wasReset bool) error {
	return s.smtpClient.SendPasswordChangedEmail(ctx, to, wasReset)
}

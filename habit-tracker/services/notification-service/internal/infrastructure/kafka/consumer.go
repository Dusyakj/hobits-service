package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"notification-service/internal/config"
	"notification-service/internal/domain/service"
	eventspb "notification-service/proto/events/v1"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	reader             *kafka.Reader
	notificationService service.NotificationService
	emailService       service.EmailService
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(
	cfg *config.KafkaConfig,
	notificationService service.NotificationService,
	emailService service.EmailService,
) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          cfg.Topics[0], // user-events
		MinBytes:       10e3,           // 10KB
		MaxBytes:       10e6,           // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	return &Consumer{
		reader:             reader,
		notificationService: notificationService,
		emailService:       emailService,
	}
}

// Start starts consuming messages from Kafka
func (c *Consumer) Start(ctx context.Context) error {
	log.Println("Starting Kafka consumer...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer...")
			return c.reader.Close()
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			if err := c.processMessage(ctx, message); err != nil {
				log.Printf("Error processing message: %v", err)
				// Continue processing other messages even if one fails
			}
		}
	}
}

// processMessage processes a Kafka message
func (c *Consumer) processMessage(ctx context.Context, message kafka.Message) error {
	var event eventspb.Event
	if err := proto.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	log.Printf("Received event: %s (ID: %s)", event.EventType.String(), event.EventId)

	switch event.EventType {
	case eventspb.EventType_EVENT_TYPE_USER_REGISTERED:
		return c.handleUserRegistered(ctx, event.GetUserRegistered())
	case eventspb.EventType_EVENT_TYPE_EMAIL_VERIFICATION_REQUESTED:
		return c.handleEmailVerificationRequested(ctx, event.GetEmailVerificationRequested())
	case eventspb.EventType_EVENT_TYPE_PASSWORD_RESET_REQUESTED:
		return c.handlePasswordResetRequested(ctx, event.GetPasswordResetRequested())
	case eventspb.EventType_EVENT_TYPE_PASSWORD_CHANGED:
		return c.handlePasswordChanged(ctx, event.GetPasswordChanged())
	default:
		log.Printf("Unknown event type: %s", event.EventType.String())
		return nil
	}
}

// handleUserRegistered handles user registration events
func (c *Consumer) handleUserRegistered(ctx context.Context, event *eventspb.UserRegisteredEvent) error {
	if event == nil {
		return fmt.Errorf("user registered event is nil")
	}

	log.Printf("Sending verification email to %s (user_id: %s)", event.Email, event.UserId)

	// Send verification email
	err := c.emailService.SendVerificationEmail(
		ctx,
		event.Email,
		event.Username,
		event.FirstName,
		event.VerificationToken,
	)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	log.Printf("Verification email sent successfully to %s", event.Email)
	return nil
}

// handleEmailVerificationRequested handles email verification request events
func (c *Consumer) handleEmailVerificationRequested(ctx context.Context, event *eventspb.EmailVerificationRequestedEvent) error {
	if event == nil {
		return fmt.Errorf("email verification requested event is nil")
	}

	log.Printf("Resending verification email to %s (user_id: %s)", event.Email, event.UserId)

	// Send verification email
	err := c.emailService.SendVerificationEmail(
		ctx,
		event.Email,
		"", // Username not provided in this event
		"", // First name not provided in this event
		event.VerificationToken,
	)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	log.Printf("Verification email resent successfully to %s", event.Email)
	return nil
}

// handlePasswordResetRequested handles password reset request events
func (c *Consumer) handlePasswordResetRequested(ctx context.Context, event *eventspb.PasswordResetRequestedEvent) error {
	if event == nil {
		return fmt.Errorf("password reset requested event is nil")
	}

	log.Printf("Sending password reset email to %s (user_id: %s)", event.Email, event.UserId)

	// Send password reset email
	err := c.emailService.SendPasswordResetEmail(
		ctx,
		event.Email,
		"", // Username not provided in this event
		"", // First name not provided in this event
		event.ResetToken,
	)
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	log.Printf("Password reset email sent successfully to %s", event.Email)
	return nil
}

// handlePasswordChanged handles password changed events
func (c *Consumer) handlePasswordChanged(ctx context.Context, event *eventspb.PasswordChangedEvent) error {
	if event == nil {
		return fmt.Errorf("password changed event is nil")
	}

	log.Printf("Sending password changed notification to %s (user_id: %s, was_reset: %v)",
		event.Email, event.UserId, event.WasReset)

	// Send password changed notification email
	err := c.emailService.SendPasswordChangedEmail(
		ctx,
		event.Email,
		event.WasReset,
	)
	if err != nil {
		return fmt.Errorf("failed to send password changed email: %w", err)
	}

	log.Printf("Password changed notification sent successfully to %s", event.Email)
	return nil
}

// Close closes the Kafka consumer
func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

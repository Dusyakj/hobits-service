package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"user-service/internal/config"
	eventspb "user-service/proto/events/v1"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Producer handles publishing events to Kafka
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg *config.KafkaConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    10,
		BatchTimeout: 10 * time.Millisecond,
		Async:        true, // Async for better performance
	}

	return &Producer{
		writer: writer,
	}
}

// PublishUserRegisteredEvent publishes a user registration event
func (p *Producer) PublishUserRegisteredEvent(ctx context.Context, event *UserRegisteredEvent) error {
	// Create protobuf event
	protoEvent := &eventspb.Event{
		EventId:   event.EventID,
		EventType: eventspb.EventType_EVENT_TYPE_USER_REGISTERED,
		Timestamp: timestamppb.New(event.CreatedAt),
		Payload: &eventspb.Event_UserRegistered{
			UserRegistered: &eventspb.UserRegisteredEvent{
				UserId:            event.UserID,
				Email:             event.Email,
				Username:          event.Username,
				FirstName:         event.FirstName,
				VerificationToken: event.VerificationToken,
				Timezone:          event.Timezone,
				CreatedAt:         timestamppb.New(event.CreatedAt),
			},
		},
	}

	// Marshal to protobuf bytes
	data, err := proto.Marshal(protoEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.UserID),
		Value: data,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to publish user registered event: %w", err)
	}

	log.Printf("Published user registered event for user_id: %s", event.UserID)
	return nil
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// PublishPasswordResetRequestedEvent publishes a password reset requested event
func (p *Producer) PublishPasswordResetRequestedEvent(ctx context.Context, event *PasswordResetRequestedEvent) error {
	protoEvent := &eventspb.Event{
		EventId:   event.EventID,
		EventType: eventspb.EventType_EVENT_TYPE_PASSWORD_RESET_REQUESTED,
		Timestamp: timestamppb.New(event.RequestedAt),
		Payload: &eventspb.Event_PasswordResetRequested{
			PasswordResetRequested: &eventspb.PasswordResetRequestedEvent{
				UserId:      event.UserID,
				Email:       event.Email,
				ResetToken:  event.ResetToken,
				RequestedAt: timestamppb.New(event.RequestedAt),
			},
		},
	}

	data, err := proto.Marshal(protoEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.UserID),
		Value: data,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to publish password reset requested event: %w", err)
	}

	log.Printf("Published password reset requested event for user_id: %s", event.UserID)
	return nil
}

// PublishPasswordChangedEvent publishes a password changed event
func (p *Producer) PublishPasswordChangedEvent(ctx context.Context, event *PasswordChangedEvent) error {
	protoEvent := &eventspb.Event{
		EventId:   event.EventID,
		EventType: eventspb.EventType_EVENT_TYPE_PASSWORD_CHANGED,
		Timestamp: timestamppb.New(event.ChangedAt),
		Payload: &eventspb.Event_PasswordChanged{
			PasswordChanged: &eventspb.PasswordChangedEvent{
				UserId:    event.UserID,
				Email:     event.Email,
				ChangedAt: timestamppb.New(event.ChangedAt),
				WasReset:  event.WasReset,
			},
		},
	}

	data, err := proto.Marshal(protoEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.UserID),
		Value: data,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to publish password changed event: %w", err)
	}

	log.Printf("Published password changed event for user_id: %s (was_reset: %v)", event.UserID, event.WasReset)
	return nil
}

// UserRegisteredEvent represents a user registration event
type UserRegisteredEvent struct {
	EventID           string
	UserID            string
	Email             string
	Username          string
	FirstName         string
	VerificationToken string
	Timezone          string
	CreatedAt         time.Time
}

// PasswordResetRequestedEvent represents a password reset request event
type PasswordResetRequestedEvent struct {
	EventID     string
	UserID      string
	Email       string
	ResetToken  string
	RequestedAt time.Time
}

// PasswordChangedEvent represents a password change event
type PasswordChangedEvent struct {
	EventID   string
	UserID    string
	Email     string
	ChangedAt time.Time
	WasReset  bool
}

// Helper function to create event ID
func NewEventID() string {
	return uuid.New().String()
}

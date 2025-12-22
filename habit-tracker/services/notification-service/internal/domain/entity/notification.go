package entity

import (
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
	NotificationTypePush  NotificationType = "push"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
)

// Notification represents a notification entity
type Notification struct {
	ID        string
	UserID    string
	Type      NotificationType
	Status    NotificationStatus
	Subject   string
	Content   string
	To        string
	Metadata  map[string]string
	SentAt    *time.Time
	FailedAt  *time.Time
	Error     *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// EmailVerificationData contains data for email verification notification
type EmailVerificationData struct {
	UserID            string
	Email             string
	Username          string
	FirstName         string
	VerificationToken string
	VerificationURL   string
}

// PasswordResetData contains data for password reset notification
type PasswordResetData struct {
	UserID    string
	Email     string
	Username  string
	FirstName string
	ResetToken string
	ResetURL   string
}

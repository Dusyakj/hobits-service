package entity

import (
	"time"

	"github.com/google/uuid"
)

// HabitConfirmation represents a single confirmation/completion of a habit
type HabitConfirmation struct {
	ID      uuid.UUID
	HabitID uuid.UUID
	UserID  uuid.UUID

	// Confirmation details
	ConfirmedAt      time.Time
	ConfirmedForDate string // Date in format "YYYY-MM-DD" (in habit's timezone)

	// Metadata
	Notes     *string
	CreatedAt time.Time
}

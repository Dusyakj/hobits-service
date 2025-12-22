package entity

import (
	"time"

	"github.com/google/uuid"
)

// ScheduleType represents the type of habit schedule
type ScheduleType string

const (
	ScheduleTypeInterval ScheduleType = "interval" // Every N days
	ScheduleTypeWeekly   ScheduleType = "weekly"   // Specific days of week
)

// Habit represents a user's habit
type Habit struct {
	ID     uuid.UUID
	UserID uuid.UUID

	// Basic info
	Name        string
	Description *string
	Color       *string // HEX color, e.g., "#FF5722"

	// Schedule configuration
	ScheduleType ScheduleType
	IntervalDays *int32  // Required for interval type (1=daily, 2=every other day, etc.)
	WeeklyDays   []int32 // Required for weekly type (0=Sunday, 1=Monday, ..., 6=Saturday)

	// Timezone offset in hours from UTC (-12 to +14)
	TimezoneOffsetHours int32

	// Streak state
	Streak                    int32
	NextDeadlineUTC           time.Time
	ConfirmedForCurrentPeriod bool
	LastConfirmedAt           *time.Time

	// Metadata
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsInterval returns true if the habit uses interval scheduling
func (h *Habit) IsInterval() bool {
	return h.ScheduleType == ScheduleTypeInterval
}

// IsWeekly returns true if the habit uses weekly scheduling
func (h *Habit) IsWeekly() bool {
	return h.ScheduleType == ScheduleTypeWeekly
}

// GetLocalTime converts a UTC time to the habit's local timezone
func (h *Habit) GetLocalTime(utcTime time.Time) time.Time {
	offset := time.Duration(h.TimezoneOffsetHours) * time.Hour
	return utcTime.Add(offset)
}

// GetLocalNow returns the current time in the habit's timezone
func (h *Habit) GetLocalNow() time.Time {
	return h.GetLocalTime(time.Now().UTC())
}

// GetLocalDate returns the date string in habit's timezone (YYYY-MM-DD)
func (h *Habit) GetLocalDate(utcTime time.Time) string {
	localTime := h.GetLocalTime(utcTime)
	return localTime.Format("2006-01-02")
}

// GetCurrentLocalDate returns today's date in habit's timezone (YYYY-MM-DD)
func (h *Habit) GetCurrentLocalDate() string {
	return h.GetLocalDate(time.Now().UTC())
}

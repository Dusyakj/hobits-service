package service

import (
	"context"
	"fmt"
	"habits-service/internal/domain/entity"
	"habits-service/internal/domain/repository"
	"habits-service/internal/domain/service"
	"time"

	"github.com/google/uuid"
)

type habitService struct {
	habitRepo       repository.HabitRepository
	confirmationRepo repository.HabitConfirmationRepository
}

// NewHabitService creates a new habit service
func NewHabitService(
	habitRepo repository.HabitRepository,
	confirmationRepo repository.HabitConfirmationRepository,
) service.HabitService {
	return &habitService{
		habitRepo:       habitRepo,
		confirmationRepo: confirmationRepo,
	}
}

// ConvertTimezoneToOffset converts IANA timezone string to UTC offset in hours
func ConvertTimezoneToOffset(timezone string) (int32, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return 0, fmt.Errorf("invalid timezone: %w", err)
	}

	// Get current time in that timezone
	now := time.Now().In(loc)

	// Get the offset in seconds and convert to hours
	_, offset := now.Zone()
	offsetHours := int32(offset / 3600)

	return offsetHours, nil
}

// CalculateInitialDeadline calculates the initial deadline when creating a habit
func (s *habitService) CalculateInitialDeadline(habit *entity.Habit, fromTime time.Time) time.Time {
	// Get local time for the habit's timezone
	localTime := habit.GetLocalTime(fromTime)

	// Calculate initial deadline date in local timezone
	var deadlineLocalDate time.Time

	if habit.IsInterval() {
		// For interval: first deadline is always today
		deadlineLocalDate = localTime
	} else if habit.IsWeekly() {
		// For weekly: check if today is one of the scheduled days
		currentWeekday := int32(localTime.Weekday())
		isTodayScheduled := false

		for _, scheduledDay := range habit.WeeklyDays {
			if scheduledDay == currentWeekday {
				isTodayScheduled = true
				break
			}
		}

		if isTodayScheduled {
			// Today is scheduled, deadline is today
			deadlineLocalDate = localTime
		} else {
			// Find next scheduled day
			deadlineLocalDate = s.findNextWeeklyDeadline(localTime, habit.WeeklyDays)
		}
	}

	// Set to end of day (23:59:59) in local timezone
	deadlineLocalDate = time.Date(
		deadlineLocalDate.Year(),
		deadlineLocalDate.Month(),
		deadlineLocalDate.Day(),
		23, 59, 59, 0,
		time.UTC, // We'll adjust for timezone offset below
	)

	// Convert back to UTC by subtracting the timezone offset
	offset := time.Duration(habit.TimezoneOffsetHours) * time.Hour
	nextUTC := deadlineLocalDate.Add(-offset)

	return nextUTC
}

// CalculateNextDeadline calculates the next deadline after confirmation
func (s *habitService) CalculateNextDeadline(habit *entity.Habit, fromTime time.Time) time.Time {
	// Get local time for the habit's timezone
	localTime := habit.GetLocalTime(fromTime)

	// Calculate next deadline date in local timezone
	var nextLocalDeadline time.Time

	if habit.IsInterval() {
		// For interval: add interval_days to current date
		daysToAdd := int(*habit.IntervalDays)
		nextLocalDeadline = localTime.AddDate(0, 0, daysToAdd)
	} else if habit.IsWeekly() {
		// For weekly: find next occurrence of any scheduled weekday
		nextLocalDeadline = s.findNextWeeklyDeadline(localTime, habit.WeeklyDays)
	}

	// Set to end of day (23:59:59) in local timezone
	nextLocalDeadline = time.Date(
		nextLocalDeadline.Year(),
		nextLocalDeadline.Month(),
		nextLocalDeadline.Day(),
		23, 59, 59, 0,
		time.UTC, // We'll adjust for timezone offset below
	)

	// Convert back to UTC by subtracting the timezone offset
	offset := time.Duration(habit.TimezoneOffsetHours) * time.Hour
	nextUTC := nextLocalDeadline.Add(-offset)

	return nextUTC
}

// findNextWeeklyDeadline finds the next occurrence of any scheduled weekday (starting from tomorrow)
func (s *habitService) findNextWeeklyDeadline(fromTime time.Time, weeklyDays []int32) time.Time {
	currentWeekday := int32(fromTime.Weekday())

	// Try to find a scheduled day in the next 7 days (starting from tomorrow)
	for daysAhead := int32(1); daysAhead <= 7; daysAhead++ {
		nextWeekday := (currentWeekday + daysAhead) % 7

		// Check if this weekday is scheduled
		for _, scheduledDay := range weeklyDays {
			if scheduledDay == nextWeekday {
				return fromTime.AddDate(0, 0, int(daysAhead))
			}
		}
	}

	// Fallback: add 7 days (shouldn't happen if weeklyDays is valid)
	return fromTime.AddDate(0, 0, 7)
}

func (s *habitService) CreateHabit(ctx context.Context, userID uuid.UUID, name string, description, color *string,
	scheduleType entity.ScheduleType, intervalDays *int32, weeklyDays []int32, timezone string) (*entity.Habit, error) {

	// Validate schedule type
	if scheduleType == entity.ScheduleTypeInterval && (intervalDays == nil || *intervalDays <= 0) {
		return nil, fmt.Errorf("interval_days is required and must be positive for interval schedule")
	}

	if scheduleType == entity.ScheduleTypeWeekly && len(weeklyDays) == 0 {
		return nil, fmt.Errorf("weekly_days is required for weekly schedule")
	}

	// Convert timezone to offset
	timezoneOffset, err := ConvertTimezoneToOffset(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to convert timezone: %w", err)
	}

	// Create habit entity
	habit := &entity.Habit{
		ID:                        uuid.New(),
		UserID:                    userID,
		Name:                      name,
		Description:               description,
		Color:                     color,
		ScheduleType:              scheduleType,
		IntervalDays:              intervalDays,
		WeeklyDays:                weeklyDays,
		TimezoneOffsetHours:       timezoneOffset,
		Streak:                    0,
		ConfirmedForCurrentPeriod: false,
		LastConfirmedAt:           nil,
		IsActive:                  true,
		CreatedAt:                 time.Now().UTC(),
		UpdatedAt:                 time.Now().UTC(),
	}

	// Calculate initial deadline (special logic for first deadline)
	habit.NextDeadlineUTC = s.CalculateInitialDeadline(habit, time.Now().UTC())

	// Save to database
	if err := s.habitRepo.Create(ctx, habit); err != nil {
		return nil, fmt.Errorf("failed to create habit: %w", err)
	}

	return habit, nil
}

func (s *habitService) GetHabit(ctx context.Context, habitID, userID uuid.UUID) (*entity.Habit, error) {
	habit, err := s.habitRepo.GetByIDAndUserID(ctx, habitID, userID)
	if err != nil {
		return nil, err
	}

	return habit, nil
}

func (s *habitService) ListHabits(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]*entity.Habit, int32, error) {
	habits, err := s.habitRepo.GetByUserID(ctx, userID, activeOnly)
	if err != nil {
		return nil, 0, err
	}

	return habits, int32(len(habits)), nil
}

func (s *habitService) UpdateHabit(ctx context.Context, habitID, userID uuid.UUID, name *string, description, color *string,
	scheduleType *entity.ScheduleType, intervalDays *int32, weeklyDays []int32, timezone *string) (*entity.Habit, error) {

	// Get existing habit
	habit, err := s.habitRepo.GetByIDAndUserID(ctx, habitID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if name != nil {
		habit.Name = *name
	}

	if description != nil {
		habit.Description = description
	}

	if color != nil {
		habit.Color = color
	}

	needsDeadlineRecalc := false

	if scheduleType != nil {
		habit.ScheduleType = *scheduleType
		needsDeadlineRecalc = true
	}

	if intervalDays != nil {
		habit.IntervalDays = intervalDays
		needsDeadlineRecalc = true
	}

	if len(weeklyDays) > 0 {
		habit.WeeklyDays = weeklyDays
		needsDeadlineRecalc = true
	}

	if timezone != nil {
		timezoneOffset, err := ConvertTimezoneToOffset(*timezone)
		if err != nil {
			return nil, fmt.Errorf("failed to convert timezone: %w", err)
		}
		habit.TimezoneOffsetHours = timezoneOffset
		needsDeadlineRecalc = true
	}

	// Recalculate deadline if schedule or timezone changed
	if needsDeadlineRecalc {
		habit.NextDeadlineUTC = s.CalculateNextDeadline(habit, time.Now().UTC())
		habit.ConfirmedForCurrentPeriod = false
	}

	habit.UpdatedAt = time.Now().UTC()

	// Validate schedule
	if habit.ScheduleType == entity.ScheduleTypeInterval && (habit.IntervalDays == nil || *habit.IntervalDays <= 0) {
		return nil, fmt.Errorf("interval_days is required and must be positive for interval schedule")
	}

	if habit.ScheduleType == entity.ScheduleTypeWeekly && len(habit.WeeklyDays) == 0 {
		return nil, fmt.Errorf("weekly_days is required for weekly schedule")
	}

	// Save to database
	if err := s.habitRepo.Update(ctx, habit); err != nil {
		return nil, fmt.Errorf("failed to update habit: %w", err)
	}

	return habit, nil
}

func (s *habitService) DeleteHabit(ctx context.Context, habitID, userID uuid.UUID) error {
	// Verify ownership
	_, err := s.habitRepo.GetByIDAndUserID(ctx, habitID, userID)
	if err != nil {
		return err
	}

	return s.habitRepo.Delete(ctx, habitID)
}

func (s *habitService) ConfirmHabit(ctx context.Context, habitID, userID uuid.UUID, notes *string) (*entity.Habit, *entity.HabitConfirmation, error) {
	// Get habit
	habit, err := s.habitRepo.GetByIDAndUserID(ctx, habitID, userID)
	if err != nil {
		return nil, nil, err
	}

	// Check if already confirmed for current period
	if habit.ConfirmedForCurrentPeriod {
		return nil, nil, fmt.Errorf("habit already confirmed for current period")
	}

	// Get current date in habit's timezone
	currentDate := habit.GetCurrentLocalDate()

	// Check if already confirmed for today (double-check)
	exists, err := s.confirmationRepo.ExistsForDate(ctx, habitID, currentDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check existing confirmation: %w", err)
	}

	if exists {
		return nil, nil, fmt.Errorf("habit already confirmed for date %s", currentDate)
	}

	// Create confirmation
	confirmation := &entity.HabitConfirmation{
		ID:               uuid.New(),
		HabitID:          habitID,
		UserID:           userID,
		ConfirmedAt:      time.Now().UTC(),
		ConfirmedForDate: currentDate,
		Notes:            notes,
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.confirmationRepo.Create(ctx, confirmation); err != nil {
		return nil, nil, fmt.Errorf("failed to create confirmation: %w", err)
	}

	// Increment streak
	habit.Streak++
	habit.ConfirmedForCurrentPeriod = true
	now := time.Now().UTC()
	habit.LastConfirmedAt = &now

	// Calculate next deadline
	habit.NextDeadlineUTC = s.CalculateNextDeadline(habit, now)

	// Update habit
	if err := s.habitRepo.UpdateStreakAndDeadline(ctx, habitID, habit.Streak, habit.NextDeadlineUTC, true); err != nil {
		return nil, nil, fmt.Errorf("failed to update habit: %w", err)
	}

	return habit, confirmation, nil
}

func (s *habitService) GetHabitHistory(ctx context.Context, habitID, userID uuid.UUID, limit, offset int32) ([]*entity.HabitConfirmation, int32, error) {
	// Verify ownership
	_, err := s.habitRepo.GetByIDAndUserID(ctx, habitID, userID)
	if err != nil {
		return nil, 0, err
	}

	confirmations, err := s.confirmationRepo.GetByHabitID(ctx, habitID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.confirmationRepo.CountByHabitID(ctx, habitID)
	if err != nil {
		return nil, 0, err
	}

	return confirmations, count, nil
}

func (s *habitService) GetHabitStats(ctx context.Context, habitID, userID uuid.UUID) (*repository.HabitStats, error) {
	// Verify ownership
	_, err := s.habitRepo.GetByIDAndUserID(ctx, habitID, userID)
	if err != nil {
		return nil, err
	}

	stats, err := s.confirmationRepo.GetStats(ctx, habitID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *habitService) ProcessMissedDeadlines(ctx context.Context) error {
	// Get all habits with missed deadlines
	habits, err := s.habitRepo.GetHabitsWithMissedDeadlines(ctx)
	if err != nil {
		return fmt.Errorf("failed to get habits with missed deadlines: %w", err)
	}

	for _, habit := range habits {
		// Reset streak to 0
		habit.Streak = 0

		// Calculate new deadline
		habit.NextDeadlineUTC = s.CalculateNextDeadline(habit, time.Now().UTC())

		// Reset confirmation flag
		habit.ConfirmedForCurrentPeriod = false

		// Update in database
		if err := s.habitRepo.UpdateStreakAndDeadline(ctx, habit.ID, 0, habit.NextDeadlineUTC, false); err != nil {
			fmt.Printf("Failed to reset streak for habit %s: %v\n", habit.ID, err)
			continue
		}

		fmt.Printf("Reset streak for habit %s (user: %s)\n", habit.ID, habit.UserID)
	}

	return nil
}

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
	habitRepo        repository.HabitRepository
	confirmationRepo repository.HabitConfirmationRepository
}

// NewHabitService creates a new habit service
func NewHabitService(
	habitRepo repository.HabitRepository,
	confirmationRepo repository.HabitConfirmationRepository,
) service.HabitService {
	return &habitService{
		habitRepo:        habitRepo,
		confirmationRepo: confirmationRepo,
	}
}

// ConvertTimezoneToOffset converts IANA timezone string to UTC offset in hours
func ConvertTimezoneToOffset(timezone string) (int32, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return 0, fmt.Errorf("invalid timezone: %w", err)
	}

	now := time.Now().In(loc)

	_, offset := now.Zone()
	offsetHours := int32(offset / 3600)

	return offsetHours, nil
}

// CalculateInitialDeadline calculates the initial deadline when creating a habit
func (s *habitService) CalculateInitialDeadline(habit *entity.Habit, fromTime time.Time) time.Time {
	localTime := habit.GetLocalTime(fromTime)

	var deadlineLocalDate time.Time

	if habit.IsInterval() {
		// For interval: first deadline is always today
		deadlineLocalDate = localTime
	} else if habit.IsWeekly() {
		currentWeekday := int32(localTime.Weekday())
		isTodayScheduled := false

		for _, scheduledDay := range habit.WeeklyDays {
			if scheduledDay == currentWeekday {
				isTodayScheduled = true
				break
			}
		}

		if isTodayScheduled {
			deadlineLocalDate = localTime
		} else {
			deadlineLocalDate = s.findNextWeeklyDeadline(localTime, habit.WeeklyDays)
		}
	}

	// Set to end of day (23:59:59) in local timezone
	deadlineLocalDate = time.Date(
		deadlineLocalDate.Year(),
		deadlineLocalDate.Month(),
		deadlineLocalDate.Day(),
		23, 59, 59, 0,
		time.UTC,
	)

	offset := time.Duration(habit.TimezoneOffsetHours) * time.Hour
	nextUTC := deadlineLocalDate.Add(-offset)

	return nextUTC
}

// CalculateNextDeadline calculates the next deadline after confirmation
func (s *habitService) CalculateNextDeadline(habit *entity.Habit, fromTime time.Time) time.Time {
	localTime := habit.GetLocalTime(fromTime)

	var nextLocalDeadline time.Time

	if habit.IsInterval() {
		daysToAdd := int(*habit.IntervalDays)
		nextLocalDeadline = localTime.AddDate(0, 0, daysToAdd)
	} else if habit.IsWeekly() {
		nextLocalDeadline = s.findNextWeeklyDeadline(localTime, habit.WeeklyDays)
	}

	// Set to end of day (23:59:59) in local timezone
	nextLocalDeadline = time.Date(
		nextLocalDeadline.Year(),
		nextLocalDeadline.Month(),
		nextLocalDeadline.Day(),
		23, 59, 59, 0,
		time.UTC,
	)

	offset := time.Duration(habit.TimezoneOffsetHours) * time.Hour
	nextUTC := nextLocalDeadline.Add(-offset)

	return nextUTC
}

// findNextWeeklyDeadline finds the next occurrence of any scheduled weekday (starting from tomorrow)
func (s *habitService) findNextWeeklyDeadline(fromTime time.Time, weeklyDays []int32) time.Time {
	currentWeekday := int32(fromTime.Weekday())

	for daysAhead := int32(1); daysAhead <= 7; daysAhead++ {
		nextWeekday := (currentWeekday + daysAhead) % 7

		for _, scheduledDay := range weeklyDays {
			if scheduledDay == nextWeekday {
				return fromTime.AddDate(0, 0, int(daysAhead))
			}
		}
	}

	return fromTime.AddDate(0, 0, 7)
}

func (s *habitService) CreateHabit(ctx context.Context, userID uuid.UUID, name string, description, color *string,
	scheduleType entity.ScheduleType, intervalDays *int32, weeklyDays []int32, timezone string) (*entity.Habit, error) {

	if scheduleType == entity.ScheduleTypeInterval && (intervalDays == nil || *intervalDays <= 0) {
		return nil, fmt.Errorf("interval_days is required and must be positive for interval schedule")
	}

	if scheduleType == entity.ScheduleTypeWeekly && len(weeklyDays) == 0 {
		return nil, fmt.Errorf("weekly_days is required for weekly schedule")
	}

	timezoneOffset, err := ConvertTimezoneToOffset(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to convert timezone: %w", err)
	}

	confirmedForCurrentPeriod := false
	if scheduleType == entity.ScheduleTypeWeekly {
		// Get current time in user's timezone
		loc, _ := time.LoadLocation(timezone)
		now := time.Now().In(loc)
		currentWeekday := int32(now.Weekday())

		isTodayScheduled := false
		for _, scheduledDay := range weeklyDays {
			if scheduledDay == currentWeekday {
				isTodayScheduled = true
				break
			}
		}

		if !isTodayScheduled {
			confirmedForCurrentPeriod = true
		}
	}

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
		ConfirmedForCurrentPeriod: confirmedForCurrentPeriod,
		LastConfirmedAt:           nil,
		IsActive:                  true,
		CreatedAt:                 time.Now().UTC(),
		UpdatedAt:                 time.Now().UTC(),
	}

	// Calculate initial deadline (special logic for first deadline)
	habit.NextDeadlineUTC = s.CalculateInitialDeadline(habit, time.Now().UTC())

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

	habit, err := s.habitRepo.GetByIDAndUserID(ctx, habitID, userID)
	if err != nil {
		return nil, err
	}

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

	if habit.ScheduleType == entity.ScheduleTypeInterval && (habit.IntervalDays == nil || *habit.IntervalDays <= 0) {
		return nil, fmt.Errorf("interval_days is required and must be positive for interval schedule")
	}

	if habit.ScheduleType == entity.ScheduleTypeWeekly && len(habit.WeeklyDays) == 0 {
		return nil, fmt.Errorf("weekly_days is required for weekly schedule")
	}

	if err := s.habitRepo.Update(ctx, habit); err != nil {
		return nil, fmt.Errorf("failed to update habit: %w", err)
	}

	return habit, nil
}

func (s *habitService) DeleteHabit(ctx context.Context, habitID, userID uuid.UUID) error {
	_, err := s.habitRepo.GetByIDAndUserID(ctx, habitID, userID)
	if err != nil {
		return err
	}

	return s.habitRepo.Delete(ctx, habitID)
}

func (s *habitService) ConfirmHabit(ctx context.Context, habitID, userID uuid.UUID, notes *string) (*entity.Habit, *entity.HabitConfirmation, error) {
	habit, err := s.habitRepo.GetByIDAndUserID(ctx, habitID, userID)
	if err != nil {
		return nil, nil, err
	}

	if habit.ConfirmedForCurrentPeriod {
		return nil, nil, fmt.Errorf("habit already confirmed for current period")
	}

	// Get current date in habit's timezone
	currentDate := habit.GetCurrentLocalDate()

	exists, err := s.confirmationRepo.ExistsForDate(ctx, habitID, currentDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check existing confirmation: %w", err)
	}

	if exists {
		return nil, nil, fmt.Errorf("habit already confirmed for date %s", currentDate)
	}

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

	habit.Streak++
	habit.ConfirmedForCurrentPeriod = true
	now := time.Now().UTC()
	habit.LastConfirmedAt = &now

	habit.NextDeadlineUTC = s.CalculateNextDeadline(habit, now)

	if err := s.habitRepo.UpdateStreakAndDeadline(ctx, habitID, habit.Streak, habit.NextDeadlineUTC, true); err != nil {
		return nil, nil, fmt.Errorf("failed to update habit: %w", err)
	}

	return habit, confirmation, nil
}

func (s *habitService) GetHabitHistory(ctx context.Context, habitID, userID uuid.UUID, limit, offset int32) ([]*entity.HabitConfirmation, int32, error) {
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
	habits, err := s.habitRepo.GetHabitsWithMissedDeadlines(ctx)
	if err != nil {
		return fmt.Errorf("failed to get habits with missed deadlines: %w", err)
	}

	for _, habit := range habits {
		habit.Streak = 0

		habit.NextDeadlineUTC = s.CalculateNextDeadline(habit, time.Now().UTC())

		habit.ConfirmedForCurrentPeriod = false

		if err := s.habitRepo.UpdateStreakAndDeadline(ctx, habit.ID, 0, habit.NextDeadlineUTC, false); err != nil {
			fmt.Printf("Failed to reset streak for habit %s: %v\n", habit.ID, err)
			continue
		}

		fmt.Printf("Reset streak for habit %s (user: %s)\n", habit.ID, habit.UserID)
	}

	return nil
}

// ResetConfirmationFlags resets ConfirmedForCurrentPeriod flag for habits
// where today is the deadline day (new period started)
func (s *habitService) ResetConfirmationFlags(ctx context.Context) error {
	now := time.Now().UTC()

	// Get habits with deadline in ±24 hour window (to handle all timezones)
	fromTime := now.Add(-24 * time.Hour)
	toTime := now.Add(24 * time.Hour)

	habits, err := s.habitRepo.GetHabitsToResetConfirmation(ctx, fromTime, toTime)
	if err != nil {
		return fmt.Errorf("failed to get habits to reset confirmation: %w", err)
	}

	for _, habit := range habits {
		// Get current date in habit's timezone
		currentLocalDate := habit.GetCurrentLocalDate()

		// Get deadline date in habit's timezone
		deadlineLocal := habit.GetLocalTime(habit.NextDeadlineUTC)
		deadlineDate := time.Date(
			deadlineLocal.Year(),
			deadlineLocal.Month(),
			deadlineLocal.Day(),
			0, 0, 0, 0,
			time.UTC,
		).Format("2006-01-02")

		// If today is the deadline day, reset confirmation flag
		if currentLocalDate == deadlineDate {
			if err := s.habitRepo.ResetConfirmationFlag(ctx, habit.ID); err != nil {
				fmt.Printf("Failed to reset confirmation flag for habit %s: %v\n", habit.ID, err)
				continue
			}

			fmt.Printf("Reset confirmation flag for habit %s (user: %s) - new period started\n", habit.ID, habit.UserID)
		}
	}

	return nil
}

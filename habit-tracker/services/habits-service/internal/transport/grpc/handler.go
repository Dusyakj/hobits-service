package grpc

import (
	"context"
	"fmt"
	"habits-service/internal/domain/entity"
	"habits-service/internal/domain/service"
	pb "habits-service/proto/habits/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type HabitServiceHandler struct {
	pb.UnimplementedHabitServiceServer
	habitService service.HabitService
}

func NewHabitServiceHandler(habitService service.HabitService) *HabitServiceHandler {
	return &HabitServiceHandler{
		habitService: habitService,
	}
}

func mapScheduleTypeToProto(scheduleType entity.ScheduleType) pb.ScheduleType {
	switch scheduleType {
	case entity.ScheduleTypeInterval:
		return pb.ScheduleType_SCHEDULE_TYPE_INTERVAL
	case entity.ScheduleTypeWeekly:
		return pb.ScheduleType_SCHEDULE_TYPE_WEEKLY
	default:
		return pb.ScheduleType_SCHEDULE_TYPE_UNSPECIFIED
	}
}

func mapScheduleTypeFromProto(scheduleType pb.ScheduleType) entity.ScheduleType {
	switch scheduleType {
	case pb.ScheduleType_SCHEDULE_TYPE_INTERVAL:
		return entity.ScheduleTypeInterval
	case pb.ScheduleType_SCHEDULE_TYPE_WEEKLY:
		return entity.ScheduleTypeWeekly
	default:
		return ""
	}
}

func mapHabitToProto(habit *entity.Habit) *pb.Habit {
	h := &pb.Habit{
		Id:                        habit.ID.String(),
		UserId:                    habit.UserID.String(),
		Name:                      habit.Name,
		ScheduleType:              mapScheduleTypeToProto(habit.ScheduleType),
		TimezoneOffsetHours:       habit.TimezoneOffsetHours,
		Streak:                    habit.Streak,
		NextDeadlineUtc:           timestamppb.New(habit.NextDeadlineUTC),
		ConfirmedForCurrentPeriod: habit.ConfirmedForCurrentPeriod,
		IsActive:                  habit.IsActive,
		CreatedAt:                 timestamppb.New(habit.CreatedAt),
		UpdatedAt:                 timestamppb.New(habit.UpdatedAt),
	}

	if habit.Description != nil {
		h.Description = habit.Description
	}

	if habit.Color != nil {
		h.Color = habit.Color
	}

	if habit.IntervalDays != nil {
		h.IntervalDays = habit.IntervalDays
	}

	if habit.WeeklyDays != nil {
		h.WeeklyDays = habit.WeeklyDays
	}

	if habit.LastConfirmedAt != nil {
		h.LastConfirmedAt = timestamppb.New(*habit.LastConfirmedAt)
	}

	return h
}

func mapConfirmationToProto(confirmation *entity.HabitConfirmation) *pb.HabitConfirmation {
	c := &pb.HabitConfirmation{
		Id:               confirmation.ID.String(),
		HabitId:          confirmation.HabitID.String(),
		UserId:           confirmation.UserID.String(),
		ConfirmedAt:      timestamppb.New(confirmation.ConfirmedAt),
		ConfirmedForDate: confirmation.ConfirmedForDate,
		CreatedAt:        timestamppb.New(confirmation.CreatedAt),
	}

	if confirmation.Notes != nil {
		c.Notes = confirmation.Notes
	}

	return c
}

// RPC Handlers

func (h *HabitServiceHandler) CreateHabit(ctx context.Context, req *pb.CreateHabitRequest) (*pb.CreateHabitResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	if req.Timezone == "" {
		return nil, status.Error(codes.InvalidArgument, "timezone is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	scheduleType := mapScheduleTypeFromProto(req.ScheduleType)
	if scheduleType == "" {
		return nil, status.Error(codes.InvalidArgument, "valid schedule_type is required")
	}

	var description, color *string
	if req.Description != nil {
		description = req.Description
	}
	if req.Color != nil {
		color = req.Color
	}

	var intervalDays *int32
	if req.IntervalDays != nil {
		intervalDays = req.IntervalDays
	}

	habit, err := h.habitService.CreateHabit(
		ctx, userID, req.Name, description, color,
		scheduleType, intervalDays, req.WeeklyDays, req.Timezone,
	)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create habit: %v", err))
	}

	return &pb.CreateHabitResponse{
		Habit: mapHabitToProto(habit),
	}, nil
}

func (h *HabitServiceHandler) GetHabit(ctx context.Context, req *pb.GetHabitRequest) (*pb.GetHabitResponse, error) {
	if req.HabitId == "" {
		return nil, status.Error(codes.InvalidArgument, "habit_id is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	habitID, err := uuid.Parse(req.HabitId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid habit_id")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	habit, err := h.habitService.GetHabit(ctx, habitID, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "habit not found")
	}

	return &pb.GetHabitResponse{
		Habit: mapHabitToProto(habit),
	}, nil
}

func (h *HabitServiceHandler) ListHabits(ctx context.Context, req *pb.ListHabitsRequest) (*pb.ListHabitsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	activeOnly := false
	if req.ActiveOnly != nil {
		activeOnly = *req.ActiveOnly
	}

	habits, totalCount, err := h.habitService.ListHabits(ctx, userID, activeOnly)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list habits: %v", err))
	}

	protoHabits := make([]*pb.Habit, len(habits))
	for i, habit := range habits {
		protoHabits[i] = mapHabitToProto(habit)
	}

	return &pb.ListHabitsResponse{
		Habits:     protoHabits,
		TotalCount: totalCount,
	}, nil
}

func (h *HabitServiceHandler) UpdateHabit(ctx context.Context, req *pb.UpdateHabitRequest) (*pb.UpdateHabitResponse, error) {
	if req.HabitId == "" {
		return nil, status.Error(codes.InvalidArgument, "habit_id is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	habitID, err := uuid.Parse(req.HabitId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid habit_id")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	var scheduleType *entity.ScheduleType
	if req.ScheduleType != nil {
		st := mapScheduleTypeFromProto(*req.ScheduleType)
		scheduleType = &st
	}

	habit, err := h.habitService.UpdateHabit(
		ctx, habitID, userID,
		req.Name, req.Description, req.Color,
		scheduleType, req.IntervalDays, req.WeeklyDays, req.Timezone,
	)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update habit: %v", err))
	}

	return &pb.UpdateHabitResponse{
		Habit: mapHabitToProto(habit),
	}, nil
}

func (h *HabitServiceHandler) DeleteHabit(ctx context.Context, req *pb.DeleteHabitRequest) (*pb.DeleteHabitResponse, error) {
	if req.HabitId == "" {
		return nil, status.Error(codes.InvalidArgument, "habit_id is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	habitID, err := uuid.Parse(req.HabitId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid habit_id")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	err = h.habitService.DeleteHabit(ctx, habitID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete habit: %v", err))
	}

	return &pb.DeleteHabitResponse{
		Success: true,
	}, nil
}

func (h *HabitServiceHandler) ConfirmHabit(ctx context.Context, req *pb.ConfirmHabitRequest) (*pb.ConfirmHabitResponse, error) {
	if req.HabitId == "" {
		return nil, status.Error(codes.InvalidArgument, "habit_id is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	habitID, err := uuid.Parse(req.HabitId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid habit_id")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	habit, confirmation, err := h.habitService.ConfirmHabit(ctx, habitID, userID, req.Notes)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to confirm habit: %v", err))
	}

	return &pb.ConfirmHabitResponse{
		Habit:        mapHabitToProto(habit),
		Confirmation: mapConfirmationToProto(confirmation),
	}, nil
}

func (h *HabitServiceHandler) GetHabitHistory(ctx context.Context, req *pb.GetHabitHistoryRequest) (*pb.GetHabitHistoryResponse, error) {
	if req.HabitId == "" {
		return nil, status.Error(codes.InvalidArgument, "habit_id is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	habitID, err := uuid.Parse(req.HabitId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid habit_id")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	limit := int32(30)
	if req.Limit != nil {
		limit = *req.Limit
	}

	offset := int32(0)
	if req.Offset != nil {
		offset = *req.Offset
	}

	confirmations, totalCount, err := h.habitService.GetHabitHistory(ctx, habitID, userID, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get habit history: %v", err))
	}

	protoConfirmations := make([]*pb.HabitConfirmation, len(confirmations))
	for i, confirmation := range confirmations {
		protoConfirmations[i] = mapConfirmationToProto(confirmation)
	}

	return &pb.GetHabitHistoryResponse{
		Confirmations: protoConfirmations,
		TotalCount:    totalCount,
	}, nil
}

func (h *HabitServiceHandler) GetHabitStats(ctx context.Context, req *pb.GetHabitStatsRequest) (*pb.GetHabitStatsResponse, error) {
	if req.HabitId == "" {
		return nil, status.Error(codes.InvalidArgument, "habit_id is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	habitID, err := uuid.Parse(req.HabitId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid habit_id")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	stats, err := h.habitService.GetHabitStats(ctx, habitID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get habit stats: %v", err))
	}

	resp := &pb.GetHabitStatsResponse{
		CurrentStreak:      stats.CurrentStreak,
		LongestStreak:      stats.LongestStreak,
		TotalConfirmations: stats.TotalConfirmations,
		CompletionRate:     stats.CompletionRate,
	}

	if stats.FirstConfirmation != nil {
		resp.FirstConfirmation = timestamppb.New(*stats.FirstConfirmation)
	}

	if stats.LastConfirmation != nil {
		resp.LastConfirmation = timestamppb.New(*stats.LastConfirmation)
	}

	return resp, nil
}

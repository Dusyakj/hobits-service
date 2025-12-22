package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"api-gateway/internal/middleware"
	pb "api-gateway/proto/habits/v1"
)

// HabitHandler handles habit-related HTTP requests
type HabitHandler struct {
	habitClient pb.HabitServiceClient
}

// NewHabitHandler creates a new habit handler
func NewHabitHandler(habitClient pb.HabitServiceClient) *HabitHandler {
	return &HabitHandler{
		habitClient: habitClient,
	}
}

// CreateHabit handles habit creation
// @Summary Create a new habit
// @Description Create a new habit with schedule configuration
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{name=string,description=string,color=string,schedule_type=string,interval_days=int,weekly_days=[]int,timezone=string} true "Create habit request"
// @Success 201 {object} object{message=string,habit=object}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/habits/create [post]
func (h *HabitHandler) CreateHabit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user_id from context (set by auth middleware)
	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name         string  `json:"name"`
		Description  *string `json:"description"`
		Color        *string `json:"color"`
		ScheduleType string  `json:"schedule_type"` // "interval" or "weekly"
		IntervalDays *int32  `json:"interval_days"`
		WeeklyDays   []int32 `json:"weekly_days"`
		Timezone     string  `json:"timezone"` // IANA timezone string
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Map schedule type to proto enum and validate schedule fields
	var scheduleType pb.ScheduleType
	var intervalDays *int32
	var weeklyDays []int32

	switch strings.ToLower(req.ScheduleType) {
	case "interval":
		scheduleType = pb.ScheduleType_SCHEDULE_TYPE_INTERVAL
		if req.IntervalDays == nil || *req.IntervalDays <= 0 {
			http.Error(w, "interval_days is required and must be positive for interval schedule", http.StatusBadRequest)
			return
		}
		intervalDays = req.IntervalDays
		// weeklyDays stays nil
	case "weekly":
		scheduleType = pb.ScheduleType_SCHEDULE_TYPE_WEEKLY
		if req.WeeklyDays == nil || len(req.WeeklyDays) == 0 {
			http.Error(w, "weekly_days is required for weekly schedule", http.StatusBadRequest)
			return
		}
		weeklyDays = req.WeeklyDays
		// intervalDays stays nil
	default:
		http.Error(w, "Invalid schedule_type", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.CreateHabitRequest{
		UserId:       userID,
		Name:         req.Name,
		Description:  req.Description,
		Color:        req.Color,
		ScheduleType: scheduleType,
		IntervalDays: intervalDays,
		WeeklyDays:   weeklyDays,
		Timezone:     req.Timezone,
	}

	resp, err := h.habitClient.CreateHabit(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp.Habit)
}

// GetHabit retrieves a single habit by ID
// @Summary Get habit by ID
// @Description Retrieve a single habit by its ID
// @Tags habits
// @Produce json
// @Security BearerAuth
// @Param id query string true "Habit ID"
// @Success 200 {object} object{id=string,name=string,description=string,color=string,schedule_type=string,streak=int}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /api/v1/habits/get [get]
func (h *HabitHandler) GetHabit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract habit ID from query parameter
	habitID := r.URL.Query().Get("id")
	if habitID == "" {
		http.Error(w, "Habit ID is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.GetHabitRequest{
		HabitId: habitID,
		UserId:  userID,
	}

	resp, err := h.habitClient.GetHabit(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Habit)
}

// ListHabits retrieves all habits for the authenticated user
// @Summary List all habits
// @Description Get all habits for the authenticated user
// @Tags habits
// @Produce json
// @Security BearerAuth
// @Param active_only query boolean false "Filter only active habits"
// @Success 200 {object} object{habits=[]object,total_count=int}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/habits/list [get]
func (h *HabitHandler) ListHabits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	activeOnly := r.URL.Query().Get("active_only") == "true"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	activeOnlyPtr := &activeOnly
	grpcReq := &pb.ListHabitsRequest{
		UserId:     userID,
		ActiveOnly: activeOnlyPtr,
	}

	resp, err := h.habitClient.ListHabits(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateHabit updates an existing habit
// @Summary Update habit
// @Description Update an existing habit's properties
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query string true "Habit ID"
// @Param request body object{name=string,description=string,color=string,schedule_type=string,interval_days=int,weekly_days=[]int,timezone=string} true "Update habit request"
// @Success 200 {object} object{message=string,habit=object}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /api/v1/habits/update [post]
func (h *HabitHandler) UpdateHabit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	habitID := r.URL.Query().Get("id")
	if habitID == "" {
		http.Error(w, "Habit ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Name         *string `json:"name"`
		Description  *string `json:"description"`
		Color        *string `json:"color"`
		ScheduleType *string `json:"schedule_type"`
		IntervalDays *int32  `json:"interval_days"`
		WeeklyDays   []int32 `json:"weekly_days"`
		Timezone     *string `json:"timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.UpdateHabitRequest{
		HabitId:     habitID,
		UserId:      userID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Timezone:    req.Timezone,
	}

	// Map schedule type if provided and validate schedule fields
	if req.ScheduleType != nil {
		var scheduleType pb.ScheduleType
		switch strings.ToLower(*req.ScheduleType) {
		case "interval":
			scheduleType = pb.ScheduleType_SCHEDULE_TYPE_INTERVAL
			if req.IntervalDays == nil || *req.IntervalDays <= 0 {
				http.Error(w, "interval_days is required and must be positive for interval schedule", http.StatusBadRequest)
				return
			}
			grpcReq.IntervalDays = req.IntervalDays
			// Don't set WeeklyDays - it will remain nil
		case "weekly":
			scheduleType = pb.ScheduleType_SCHEDULE_TYPE_WEEKLY
			if req.WeeklyDays == nil || len(req.WeeklyDays) == 0 {
				http.Error(w, "weekly_days is required for weekly schedule", http.StatusBadRequest)
				return
			}
			grpcReq.WeeklyDays = req.WeeklyDays
			// Don't set IntervalDays - it will remain nil
		default:
			http.Error(w, "Invalid schedule_type", http.StatusBadRequest)
			return
		}
		grpcReq.ScheduleType = &scheduleType
	} else {
		// If schedule_type is not being updated, allow setting interval_days or weekly_days independently
		grpcReq.IntervalDays = req.IntervalDays
		grpcReq.WeeklyDays = req.WeeklyDays
	}

	resp, err := h.habitClient.UpdateHabit(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Habit)
}

// DeleteHabit soft deletes a habit
// @Summary Delete habit
// @Description Soft delete a habit (mark as inactive)
// @Tags habits
// @Produce json
// @Security BearerAuth
// @Param id query string true "Habit ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /api/v1/habits/delete [post]
func (h *HabitHandler) DeleteHabit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	habitID := r.URL.Query().Get("id")
	if habitID == "" {
		http.Error(w, "Habit ID is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.DeleteHabitRequest{
		HabitId: habitID,
		UserId:  userID,
	}

	_, err := h.habitClient.DeleteHabit(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Habit deleted successfully",
	})
}

// ConfirmHabit confirms habit completion for the current period
// @Summary Confirm habit completion
// @Description Mark habit as completed for the current period and increment streak
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query string true "Habit ID"
// @Param request body object{notes=string} false "Optional notes"
// @Success 200 {object} object{message=string,habit=object,confirmation=object}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /api/v1/habits/confirm [post]
func (h *HabitHandler) ConfirmHabit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	habitID := r.URL.Query().Get("id")
	if habitID == "" {
		http.Error(w, "Habit ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Notes *string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Notes are optional, so ignore decode errors
		req.Notes = nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.ConfirmHabitRequest{
		HabitId: habitID,
		UserId:  userID,
		Notes:   req.Notes,
	}

	resp, err := h.habitClient.ConfirmHabit(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetHabitHistory retrieves confirmation history for a habit
// @Summary Get habit confirmation history
// @Description Retrieve the history of confirmations for a habit with pagination
// @Tags habits
// @Produce json
// @Security BearerAuth
// @Param id query string true "Habit ID"
// @Param limit query int false "Limit (default 30)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} object{confirmations=[]object,total_count=int}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /api/v1/habits/history [get]
func (h *HabitHandler) GetHabitHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	habitID := r.URL.Query().Get("id")
	if habitID == "" {
		http.Error(w, "Habit ID is required", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var limit, offset int32 = 30, 0
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(l)
		}
	}
	if offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
			offset = int32(o)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.GetHabitHistoryRequest{
		HabitId: habitID,
		UserId:  userID,
		Limit:   &limit,
		Offset:  &offset,
	}

	resp, err := h.habitClient.GetHabitHistory(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetHabitStats retrieves statistics for a habit
// @Summary Get habit statistics
// @Description Get statistics including current streak, longest streak, total confirmations, and completion rate
// @Tags habits
// @Produce json
// @Security BearerAuth
// @Param id query string true "Habit ID"
// @Success 200 {object} object{current_streak=int,longest_streak=int,total_confirmations=int,completion_rate=number,first_confirmation=string,last_confirmation=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /api/v1/habits/stats [get]
func (h *HabitHandler) GetHabitStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	habitID := r.URL.Query().Get("id")
	if habitID == "" {
		http.Error(w, "Habit ID is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.GetHabitStatsRequest{
		HabitId: habitID,
		UserId:  userID,
	}

	resp, err := h.habitClient.GetHabitStats(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

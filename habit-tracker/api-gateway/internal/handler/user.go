package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"api-gateway/internal/middleware"
	pb "api-gateway/proto/user/v1"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userClient pb.UserServiceClient
}

// NewUserHandler creates a new user handler
func NewUserHandler(userClient pb.UserServiceClient) *UserHandler {
	return &UserHandler{
		userClient: userClient,
	}
}

// Register handles user registration
// @Summary Register new user
// @Description Create a new user account and receive authentication tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string,username=string,password=string,first_name=string,timezone=string} true "Registration request"
// @Success 201 {object} object{message=string,user_id=string,email=string,username=string,access_token=string,refresh_token=string}
// @Failure 400 {object} object{error=string}
// @Failure 409 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/auth/register [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email     string `json:"email"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FirstName string `json:"first_name,omitempty"`
		Timezone  string `json:"timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call user-service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.RegisterRequest{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		Timezone:  req.Timezone,
	}

	resp, err := h.userClient.Register(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Registration successful. Please check your email to verify your account.",
		"user_id":    resp.User.Id,
		"email":      resp.User.Email,
		"username":   resp.User.Username,
		"email_sent": true,
	})
}

// Login handles user authentication
// @Summary User login
// @Description Authenticate user and receive access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email_or_username=string,password=string} true "Login credentials"
// @Success 200 {object} object{message=string,user_id=string,email=string,username=string,access_token=string,refresh_token=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		EmailOrUsername string `json:"email_or_username"`
		Password        string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call user-service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ipAddr := r.RemoteAddr
	userAgent := r.UserAgent()

	grpcReq := &pb.LoginRequest{
		EmailOrUsername: req.EmailOrUsername,
		Password:        req.Password,
		IpAddress:       &ipAddr,
		UserAgent:       &userAgent,
	}

	resp, err := h.userClient.Login(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Login successful",
		"user_id":       resp.User.Id,
		"email":         resp.User.Email,
		"username":      resp.User.Username,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

// Logout handles user logout
// @Summary User logout
// @Description Invalidate current user session and tokens
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{message=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/auth/logout [post]
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get session ID from context (set by auth middleware)
	sessionID := middleware.GetSessionID(r)
	if sessionID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Call user-service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.LogoutRequest{
		UserId:    userID,
		SessionId: sessionID,
	}

	_, err := h.userClient.Logout(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Logout successful",
	})
}

// GetProfile returns user profile
// @Summary Get user profile
// @Description Retrieve authenticated user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{id=string,email=string,username=string,first_name=string,timezone=string,is_active=bool,created_at=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/users/profile [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Call user-service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.GetUserRequest{
		UserId: userID,
	}

	resp, err := h.userClient.GetUser(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         resp.User.Id,
		"email":      resp.User.Email,
		"username":   resp.User.Username,
		"first_name": resp.User.FirstName,
		"timezone":   resp.User.Timezone,
		"is_active":  resp.User.IsActive,
		"created_at": resp.User.CreatedAt,
	})
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Get new access and refresh tokens using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} object{message=string,access_token=string,refresh_token=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/auth/refresh [post]
func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call user-service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	resp, err := h.userClient.RefreshToken(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Token refreshed successfully",
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

// handleGRPCError converts gRPC errors to HTTP errors
func handleGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var httpStatus int
	switch st.Code() {
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	default:
		httpStatus = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": st.Message(),
	})
}

// VerifyEmail handles email verification
// @Summary Verify email
// @Description Verify user email with token
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "Verification token"
// @Success 200 {object} object{message=string,user_id=string,email=string}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/auth/verify-email [get]
func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Call user-service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.VerifyEmailRequest{
		Token: token,
	}

	resp, err := h.userClient.VerifyEmail(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	if !resp.Success {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		errorMsg := "Verification failed"
		if resp.Error != nil {
			errorMsg = *resp.Error
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": errorMsg,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Email verified successfully",
		"user_id": resp.User.Id,
		"email":   resp.User.Email,
	})
}

// ResendVerificationEmail handles resending verification email
// @Summary Resend verification email
// @Description Resend email verification link
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string} true "Email address"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/auth/resend-verification [post]
func (h *UserHandler) ResendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Call user-service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.ResendVerificationEmailRequest{
		Email: req.Email,
	}

	resp, err := h.userClient.ResendVerificationEmail(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	if !resp.Success {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		errorMsg := "Failed to resend verification email"
		if resp.Error != nil {
			errorMsg = *resp.Error
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": errorMsg,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Verification email sent successfully",
	})
}

// ChangePassword handles password change
// @Summary Change password
// @Description Change user password (requires old password)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{old_password=string,new_password=string} true "Password change request"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/users/change-password [post]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.OldPassword == "" {
		http.Error(w, "Old password is required", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		http.Error(w, "New password is required", http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 8 {
		http.Error(w, "New password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Call user-service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.ChangePasswordRequest{
		UserId:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	resp, err := h.userClient.ChangePassword(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	if !resp.Success {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to change password",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Password changed successfully",
	})
}

// DeactivateAccount handles account deactivation
// @Summary Deactivate account
// @Description Deactivate user account (soft delete)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{message=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/users/deactivate [delete]
func (h *UserHandler) DeactivateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Call user-service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.DeactivateUserRequest{
		UserId: userID,
	}

	resp, err := h.userClient.DeactivateUser(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	if !resp.Success {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		errorMsg := "Failed to deactivate account"
		if resp.Error != nil {
			errorMsg = *resp.Error
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": errorMsg,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Account deactivated successfully",
	})
}

// ForgotPassword handles password reset request
// @Summary Request password reset
// @Description Initiates password reset process by sending email with reset link
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string} true "Email address"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/auth/forgot-password [post]
func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Call user-service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.ForgotPasswordRequest{
		Email: req.Email,
	}

	resp, err := h.userClient.ForgotPassword(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	message := "If the email exists, a password reset link has been sent"
	if resp.Message != nil {
		message = *resp.Message
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": message,
	})
}

// ResetPassword handles password reset with token
// @Summary Reset password with token
// @Description Completes password reset using token from email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{token=string,new_password=string} true "Reset token and new password"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/auth/reset-password [post]
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		http.Error(w, "New password is required", http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Call user-service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcReq := &pb.ResetPasswordRequest{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}

	resp, err := h.userClient.ResetPassword(ctx, grpcReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	if !resp.Success {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		errorMsg := "Invalid or expired reset token"
		if resp.Error != nil {
			errorMsg = *resp.Error
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": errorMsg,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Password reset successfully. Please login with your new password",
	})
}

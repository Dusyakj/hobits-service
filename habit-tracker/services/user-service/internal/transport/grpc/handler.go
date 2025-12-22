package grpc

import (
	"context"
	"fmt"

	"user-service/internal/domain/service"
	"user-service/pkg/validation"
	pb "user-service/proto/user/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServiceHandler implements pb.UserServiceServer
type UserServiceHandler struct {
	pb.UnimplementedUserServiceServer
	userService service.UserService
	authService service.AuthService
}

// NewUserServiceHandler creates a new gRPC handler
func NewUserServiceHandler(userService service.UserService, authService service.AuthService) *UserServiceHandler {
	return &UserServiceHandler{
		userService: userService,
		authService: authService,
	}
}

// Register handles user registration
func (h *UserServiceHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Validate email
	if err := validation.ValidateEmail(req.Email); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Validate username
	if err := validation.ValidateUsername(req.Username); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Validate password
	if err := validation.ValidatePassword(req.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Validate timezone
	if err := validation.ValidateTimezone(req.Timezone); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert to domain
	userCreate := toUserCreate(req)
	ipAddress := parseIPAddress(nil) // TODO: Extract from metadata
	var userAgent *string            // TODO: Extract from metadata

	// Register user
	user, _, err := h.authService.Register(ctx, userCreate, ipAddress, userAgent)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to register: %v", err))
	}

	// tokenPair is nil - user must verify email before logging in
	return &pb.RegisterResponse{
		User:      toProtoUser(user),
		EmailSent: true,
	}, nil
}

// Login handles user authentication
func (h *UserServiceHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Validate request
	if req.EmailOrUsername == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	// Parse IP and user agent
	ipAddress := parseIPAddress(req.IpAddress)
	userAgent := req.UserAgent

	// Login
	user, tokenPair, err := h.authService.Login(ctx, req.EmailOrUsername, req.Password, ipAddress, userAgent)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Convert to proto
	accessToken, refreshToken, accessExpiresAt, refreshExpiresAt := toProtoTokenPair(tokenPair)

	return &pb.LoginResponse{
		User:                  toProtoUser(user),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}

// Logout handles user logout
func (h *UserServiceHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// Parse UUIDs
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session ID")
	}

	// Logout
	if err := h.authService.Logout(ctx, userID, sessionID); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to logout: %v", err))
	}

	return &pb.LogoutResponse{Success: true}, nil
}

// RefreshToken handles token refresh
func (h *UserServiceHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	// Validate request
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "missing refresh token")
	}

	// Refresh token
	tokenPair, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	// Convert to proto
	accessToken, refreshToken, accessExpiresAt, refreshExpiresAt := toProtoTokenPair(tokenPair)

	return &pb.RefreshTokenResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}

// ValidateToken validates access token
func (h *UserServiceHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	// Validate request
	if req.AccessToken == "" {
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: strPtr("missing access token"),
		}, nil
	}

	// Validate token
	userID, sessionID, err := h.authService.ValidateAccessToken(ctx, req.AccessToken)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: strPtr(err.Error()),
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:     true,
		UserId:    strPtr(userID.String()),
		SessionId: strPtr(sessionID.String()),
	}, nil
}

// GetUser retrieves user by ID
func (h *UserServiceHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// Parse UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	// Get user
	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.GetUserResponse{
		User: toProtoUser(user),
	}, nil
}

// GetUserByEmail retrieves user by email
func (h *UserServiceHandler) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.GetUserResponse, error) {
	// Validate request
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "missing email")
	}

	// Get user
	user, err := h.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.GetUserResponse{
		User: toProtoUser(user),
	}, nil
}

// UpdateUser updates user information
func (h *UserServiceHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// Parse UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	// Convert to domain
	userUpdate := toUserUpdate(req)

	// Update user
	user, err := h.userService.UpdateUser(ctx, userID, userUpdate)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update user: %v", err))
	}

	return &pb.UpdateUserResponse{
		User: toProtoUser(user),
	}, nil
}

// ChangePassword changes user password
func (h *UserServiceHandler) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	// Parse UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	// Validate old password
	if req.OldPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "old password is required")
	}

	// Validate new password
	if err := validation.ValidatePassword(req.NewPassword); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Change password
	if err := h.userService.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to change password: %v", err))
	}

	return &pb.ChangePasswordResponse{Success: true}, nil
}

// GetUserSessions retrieves all active sessions
func (h *UserServiceHandler) GetUserSessions(ctx context.Context, req *pb.GetUserSessionsRequest) (*pb.GetUserSessionsResponse, error) {
	// Parse UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	// Get sessions
	sessions, err := h.authService.GetUserSessions(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get sessions: %v", err))
	}

	// Convert to proto
	protoSessions := make([]*pb.Session, len(sessions))
	for i, session := range sessions {
		protoSessions[i] = toProtoSession(session)
	}

	return &pb.GetUserSessionsResponse{
		Sessions: protoSessions,
	}, nil
}

// RevokeSession revokes a specific session
func (h *UserServiceHandler) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest) (*pb.RevokeSessionResponse, error) {
	// Parse UUIDs
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session ID")
	}

	// Revoke session
	if err := h.authService.RevokeSession(ctx, userID, sessionID); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to revoke session: %v", err))
	}

	return &pb.RevokeSessionResponse{Success: true}, nil
}

// RevokeAllSessions revokes all sessions for a user
func (h *UserServiceHandler) RevokeAllSessions(ctx context.Context, req *pb.RevokeAllSessionsRequest) (*pb.RevokeAllSessionsResponse, error) {
	// Parse UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	// Revoke all sessions
	count, err := h.authService.RevokeAllSessions(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to revoke sessions: %v", err))
	}

	return &pb.RevokeAllSessionsResponse{
		RevokedCount: int32(count),
	}, nil
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

// VerifyEmail verifies user email with token
func (h *UserServiceHandler) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	// Validate request
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// Verify email
	user, err := h.authService.VerifyEmail(ctx, req.Token)
	if err != nil {
		errMsg := fmt.Sprintf("failed to verify email: %v", err)
		return &pb.VerifyEmailResponse{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	return &pb.VerifyEmailResponse{
		Success: true,
		User:    toProtoUser(user),
	}, nil
}

// ResendVerificationEmail resends verification email
func (h *UserServiceHandler) ResendVerificationEmail(ctx context.Context, req *pb.ResendVerificationEmailRequest) (*pb.ResendVerificationEmailResponse, error) {
	// Validate request
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	// Resend verification email
	err := h.authService.ResendVerificationEmail(ctx, req.Email)
	if err != nil {
		errMsg := fmt.Sprintf("failed to resend verification email: %v", err)
		return &pb.ResendVerificationEmailResponse{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	return &pb.ResendVerificationEmailResponse{
		Success: true,
	}, nil
}

// DeactivateUser handles user account deactivation
func (h *UserServiceHandler) DeactivateUser(ctx context.Context, req *pb.DeactivateUserRequest) (*pb.DeactivateUserResponse, error) {
	// Parse UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	// Deactivate user
	if err := h.userService.DeactivateUser(ctx, userID); err != nil {
		errMsg := fmt.Sprintf("failed to deactivate user: %v", err)
		return &pb.DeactivateUserResponse{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	return &pb.DeactivateUserResponse{
		Success: true,
	}, nil
}

// ForgotPassword initiates password reset process
func (h *UserServiceHandler) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {
	// Validate email
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	// Initiate password reset
	if err := h.authService.ForgotPassword(ctx, req.Email); err != nil {
		// Don't reveal internal errors to user
		return &pb.ForgotPasswordResponse{
			Success: true,
			Message: strPtr("If the email exists, a password reset link has been sent"),
		}, nil
	}

	return &pb.ForgotPasswordResponse{
		Success: true,
		Message: strPtr("If the email exists, a password reset link has been sent"),
	}, nil
}

// ResetPassword completes password reset with token
func (h *UserServiceHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	// Validate request
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// Validate new password
	if err := validation.ValidatePassword(req.NewPassword); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Reset password
	if err := h.authService.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		return &pb.ResetPasswordResponse{
			Success: false,
			Error:   strPtr("Invalid or expired reset token"),
		}, nil
	}

	return &pb.ResetPasswordResponse{
		Success: true,
	}, nil
}

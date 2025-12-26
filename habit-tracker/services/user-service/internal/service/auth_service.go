package service

import (
	"context"
	"fmt"
	"net"
	"time"

	"user-service/internal/domain/entity"
	"user-service/internal/domain/repository"
	"user-service/internal/domain/service"
	"user-service/internal/infrastructure/kafka"
	"user-service/internal/infrastructure/redis"
	pkgjwt "user-service/pkg/jwt"

	"github.com/google/uuid"
)

// authService implements service.AuthService
type authService struct {
	userService             service.UserService
	sessionRepo             repository.SessionRepository
	sessionStorage          *redis.SessionStorage
	verificationTokenStore  *redis.VerificationTokenStorage
	passwordResetTokenStore *redis.PasswordResetTokenStorage
	tokenManager            *pkgjwt.TokenManager
	kafkaProducer           *kafka.Producer
}

// NewAuthService creates a new auth service
func NewAuthService(
	userService service.UserService,
	sessionRepo repository.SessionRepository,
	sessionStorage *redis.SessionStorage,
	verificationTokenStore *redis.VerificationTokenStorage,
	passwordResetTokenStore *redis.PasswordResetTokenStorage,
	tokenManager *pkgjwt.TokenManager,
	kafkaProducer *kafka.Producer,
) service.AuthService {
	return &authService{
		userService:             userService,
		sessionRepo:             sessionRepo,
		sessionStorage:          sessionStorage,
		verificationTokenStore:  verificationTokenStore,
		passwordResetTokenStore: passwordResetTokenStore,
		tokenManager:            tokenManager,
		kafkaProducer:           kafkaProducer,
	}
}

// Register registers a new user and creates session
func (s *authService) Register(
	ctx context.Context,
	userCreate *entity.UserCreate,
	ipAddress *net.IP,
	userAgent *string,
) (*entity.User, *service.TokenPair, error) {
	user, err := s.userService.CreateUser(ctx, userCreate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	verificationToken, err := s.verificationTokenStore.GenerateToken()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	if err := s.verificationTokenStore.StoreToken(ctx, verificationToken, user.ID.String()); err != nil {
		return nil, nil, fmt.Errorf("failed to store verification token: %w", err)
	}

	firstName := ""
	if user.FirstName != nil {
		firstName = *user.FirstName
	}

	event := &kafka.UserRegisteredEvent{
		EventID:           kafka.NewEventID(),
		UserID:            user.ID.String(),
		Email:             user.Email,
		Username:          user.Username,
		FirstName:         firstName,
		VerificationToken: verificationToken,
		Timezone:          user.Timezone,
		CreatedAt:         user.CreatedAt,
	}

	if err := s.kafkaProducer.PublishUserRegisteredEvent(ctx, event); err != nil {
		fmt.Printf("Warning: failed to publish user registered event: %v\n", err)
	}

	return user, nil, nil
}

// Login authenticates user and creates session
func (s *authService) Login(
	ctx context.Context,
	emailOrUsername, password string,
	ipAddress *net.IP,
	userAgent *string,
) (*entity.User, *service.TokenPair, error) {
	user, err := s.userService.GetUserByEmail(ctx, emailOrUsername)
	if err != nil {
		user, err = s.userService.GetUserByUsername(ctx, emailOrUsername)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid credentials")
		}
	}

	if !user.IsActive {
		return nil, nil, fmt.Errorf("account is deactivated")
	}

	if !user.EmailVerified {
		return nil, nil, fmt.Errorf("email not verified")
	}

	if err := s.userService.ValidatePassword(ctx, user, password); err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	tokenPair, err := s.createSession(ctx, user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return user, tokenPair, nil
}

// Logout invalidates user session
func (s *authService) Logout(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	session, err := s.sessionStorage.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found")
	}

	if session.UserID != userID {
		return fmt.Errorf("unauthorized: session does not belong to user")
	}

	if err := s.sessionStorage.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session from cache: %w", err)
	}

	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// RefreshToken generates new token pair using refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*service.TokenPair, error) {
	claims, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	session, err := s.sessionStorage.Get(ctx, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found or expired")
	}

	accessToken, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(session.UserID, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.tokenManager.GenerateRefreshToken(session.UserID, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.sessionStorage.UpdateLastActivity(ctx, session.ID); err != nil {
		// todo: Log error but don't fail
	}

	return &service.TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}

// ValidateAccessToken validates access token and returns user ID and session ID
func (s *authService) ValidateAccessToken(ctx context.Context, accessToken string) (uuid.UUID, uuid.UUID, error) {
	claims, err := s.tokenManager.ValidateAccessToken(accessToken)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid access token: %w", err)
	}

	exists, err := s.sessionStorage.Exists(ctx, claims.SessionID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("failed to check session: %w", err)
	}
	if !exists {
		return uuid.Nil, uuid.Nil, fmt.Errorf("session not found or expired")
	}

	if err := s.sessionStorage.UpdateLastActivity(ctx, claims.SessionID); err != nil {
		// TODO: Log error but don't fail
	}

	return claims.UserID, claims.SessionID, nil
}

// GetUserSessions retrieves all active sessions for a user
func (s *authService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error) {
	// Get from Redis first (active sessions)
	sessions, err := s.sessionStorage.GetByUserID(ctx, userID)
	if err != nil {
		// Fallback to PostgreSQL
		sessions, err = s.sessionRepo.GetActiveByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user sessions: %w", err)
		}
	}

	return sessions, nil
}

// RevokeSession revokes a specific session
func (s *authService) RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	session, err := s.sessionStorage.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found")
	}

	if session.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	if err := s.sessionStorage.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session from cache: %w", err)
	}

	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// RevokeAllSessions revokes all sessions for a user
func (s *authService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) (int, error) {
	sessions, err := s.sessionStorage.GetByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get sessions: %w", err)
	}

	count := len(sessions)

	if err := s.sessionStorage.DeleteByUserID(ctx, userID); err != nil {
		return 0, fmt.Errorf("failed to delete sessions from cache: %w", err)
	}

	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return 0, fmt.Errorf("failed to delete sessions: %w", err)
	}

	return count, nil
}

// createSession creates a new session and generates tokens
func (s *authService) createSession(
	ctx context.Context,
	userID uuid.UUID,
	ipAddress *net.IP,
	userAgent *string,
) (*service.TokenPair, error) {
	sessionID := uuid.New()

	accessToken, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.tokenManager.GenerateRefreshToken(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	tokenHash := pkgjwt.HashToken(refreshToken)

	now := time.Now()
	session := &entity.Session{
		ID:             sessionID,
		UserID:         userID,
		TokenHash:      tokenHash,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		ExpiresAt:      refreshExpiresAt,
		CreatedAt:      now,
		LastActivityAt: now,
	}

	if err := s.sessionStorage.Set(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session to cache: %w", err)
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		// todo: Log error but don't fail - Redis is primary storage
	}

	return &service.TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}

// VerifyEmail verifies user email with token
func (s *authService) VerifyEmail(ctx context.Context, token string) (*entity.User, error) {
	userIDStr, err := s.verificationTokenStore.GetUserIDByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired verification token: %w", err)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.EmailVerified {
		return user, nil
	}

	emailVerified := true
	updateData := &entity.UserUpdate{
		EmailVerified: &emailVerified,
	}

	updatedUser, err := s.userService.UpdateUser(ctx, userID, updateData)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if err := s.verificationTokenStore.DeleteToken(ctx, token); err != nil {
		fmt.Printf("Warning: failed to delete verification token: %v\n", err)
	}

	return updatedUser, nil
}

// ResendVerificationEmail resends verification email
func (s *authService) ResendVerificationEmail(ctx context.Context, email string) error {
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	verificationToken, err := s.verificationTokenStore.GenerateToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	if err := s.verificationTokenStore.StoreToken(ctx, verificationToken, user.ID.String()); err != nil {
		return fmt.Errorf("failed to store verification token: %w", err)
	}

	firstName := ""
	if user.FirstName != nil {
		firstName = *user.FirstName
	}

	event := &kafka.UserRegisteredEvent{
		EventID:           kafka.NewEventID(),
		UserID:            user.ID.String(),
		Email:             user.Email,
		Username:          user.Username,
		FirstName:         firstName,
		VerificationToken: verificationToken,
		Timezone:          user.Timezone,
		CreatedAt:         time.Now(),
	}

	if err := s.kafkaProducer.PublishUserRegisteredEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to publish verification event: %w", err)
	}

	return nil
}

// ForgotPassword initiates password reset process
func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists - always return success
		return nil
	}

	resetToken, err := s.passwordResetTokenStore.GenerateToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	if err := s.passwordResetTokenStore.StoreToken(ctx, resetToken, user.ID.String(), user.Email); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	event := &kafka.PasswordResetRequestedEvent{
		EventID:     kafka.NewEventID(),
		UserID:      user.ID.String(),
		Email:       user.Email,
		ResetToken:  resetToken,
		RequestedAt: time.Now(),
	}

	if err := s.kafkaProducer.PublishPasswordResetRequestedEvent(ctx, event); err != nil {
		fmt.Printf("Warning: failed to publish password reset requested event: %v\n", err)
	}

	return nil
}

// ResetPassword completes password reset with token
func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	userIDStr, email, err := s.passwordResetTokenStore.GetTokenData(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if err := s.userService.UpdatePassword(ctx, userID, newPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := s.passwordResetTokenStore.DeleteToken(ctx, token); err != nil {
		fmt.Printf("Warning: failed to delete reset token: %v\n", err)
	}

	if _, err := s.RevokeAllSessions(ctx, userID); err != nil {
		fmt.Printf("Warning: failed to revoke sessions: %v\n", err)
	}

	event := &kafka.PasswordChangedEvent{
		EventID:   kafka.NewEventID(),
		UserID:    user.ID.String(),
		Email:     email,
		ChangedAt: time.Now(),
		WasReset:  true,
	}

	if err := s.kafkaProducer.PublishPasswordChangedEvent(ctx, event); err != nil {
		fmt.Printf("Warning: failed to publish password changed event: %v\n", err)
	}

	return nil
}

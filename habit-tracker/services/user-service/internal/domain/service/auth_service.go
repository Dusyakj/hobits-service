package service

import (
	"context"
	"net"
	"time"

	"user-service/internal/domain/entity"

	"github.com/google/uuid"
)

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

// AuthService defines business logic for authentication
type AuthService interface {
	// Register registers a new user and creates session
	Register(ctx context.Context, userCreate *entity.UserCreate, ipAddress *net.IP, userAgent *string) (*entity.User, *TokenPair, error)

	// Login authenticates user and creates session
	Login(ctx context.Context, emailOrUsername, password string, ipAddress *net.IP, userAgent *string) (*entity.User, *TokenPair, error)

	// Logout invalidates user session
	Logout(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error

	// RefreshToken generates new token pair using refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)

	// ValidateAccessToken validates access token and returns user ID and session ID
	ValidateAccessToken(ctx context.Context, accessToken string) (uuid.UUID, uuid.UUID, error)

	// GetUserSessions retrieves all active sessions for a user
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error)

	// RevokeSession revokes a specific session
	RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error

	// RevokeAllSessions revokes all sessions for a user
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) (int, error)

	// VerifyEmail verifies user email with token
	VerifyEmail(ctx context.Context, token string) (*entity.User, error)

	// ResendVerificationEmail resends verification email
	ResendVerificationEmail(ctx context.Context, email string) error

	// ForgotPassword initiates password reset process
	ForgotPassword(ctx context.Context, email string) error

	// ResetPassword completes password reset with token
	ResetPassword(ctx context.Context, token, newPassword string) error
}

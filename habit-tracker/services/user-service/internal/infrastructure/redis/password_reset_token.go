package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	passwordResetTokenPrefix = "password_reset:token:"
	passwordResetTokenTTL    = 1 * time.Hour
)

// PasswordResetTokenStorage handles password reset token storage in Redis
type PasswordResetTokenStorage struct {
	client *redis.Client
}

// NewPasswordResetTokenStorage creates a new password reset token storage
func NewPasswordResetTokenStorage(client *redis.Client) *PasswordResetTokenStorage {
	return &PasswordResetTokenStorage{
		client: client,
	}
}

// GenerateToken generates a new password reset token
func (s *PasswordResetTokenStorage) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// StoreToken stores a password reset token with user ID and email
func (s *PasswordResetTokenStorage) StoreToken(ctx context.Context, token, userID, email string) error {
	key := passwordResetTokenPrefix + token
	// Store both user ID and email as JSON-like string
	value := fmt.Sprintf("%s:%s", userID, email)
	err := s.client.Set(ctx, key, value, passwordResetTokenTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to store password reset token: %w", err)
	}
	return nil
}

// GetTokenData retrieves user ID and email by password reset token
func (s *PasswordResetTokenStorage) GetTokenData(ctx context.Context, token string) (userID, email string, err error) {
	key := passwordResetTokenPrefix + token
	value, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", "", fmt.Errorf("password reset token not found or expired")
	}
	if err != nil {
		return "", "", fmt.Errorf("failed to get password reset token: %w", err)
	}
	
	// Parse value (format: "userID:email")
	var uid, em string
	if _, err := fmt.Sscanf(value, "%s:%s", &uid, &em); err != nil {
		return "", "", fmt.Errorf("invalid token data format")
	}
	
	return uid, em, nil
}

// DeleteToken deletes a password reset token
func (s *PasswordResetTokenStorage) DeleteToken(ctx context.Context, token string) error {
	key := passwordResetTokenPrefix + token
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete password reset token: %w", err)
	}
	return nil
}

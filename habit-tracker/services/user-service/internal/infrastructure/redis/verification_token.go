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
	verificationTokenPrefix = "verification:token:"
	verificationTokenTTL    = 24 * time.Hour
)

// VerificationTokenStorage handles verification token storage in Redis
type VerificationTokenStorage struct {
	client *redis.Client
}

// NewVerificationTokenStorage creates a new verification token storage
func NewVerificationTokenStorage(client *redis.Client) *VerificationTokenStorage {
	return &VerificationTokenStorage{
		client: client,
	}
}

// GenerateToken generates a new verification token
func (s *VerificationTokenStorage) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// StoreToken stores a verification token with user ID
func (s *VerificationTokenStorage) StoreToken(ctx context.Context, token, userID string) error {
	key := verificationTokenPrefix + token
	err := s.client.Set(ctx, key, userID, verificationTokenTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to store verification token: %w", err)
	}
	return nil
}

// GetUserIDByToken retrieves user ID by verification token
func (s *VerificationTokenStorage) GetUserIDByToken(ctx context.Context, token string) (string, error) {
	key := verificationTokenPrefix + token
	userID, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("verification token not found or expired")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get verification token: %w", err)
	}
	return userID, nil
}

// DeleteToken deletes a verification token
func (s *VerificationTokenStorage) DeleteToken(ctx context.Context, token string) error {
	key := verificationTokenPrefix + token
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete verification token: %w", err)
	}
	return nil
}

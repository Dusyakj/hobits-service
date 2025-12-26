package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"user-service/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SessionStorage handles session storage in Redis
type SessionStorage struct {
	client     *redis.Client
	sessionTTL time.Duration
}

// NewSessionStorage creates a new session storage
func NewSessionStorage(client *redis.Client, sessionTTL time.Duration) *SessionStorage {
	return &SessionStorage{
		client:     client,
		sessionTTL: sessionTTL,
	}
}

// sessionKey generates Redis key for session
func (s *SessionStorage) sessionKey(sessionID uuid.UUID) string {
	return fmt.Sprintf("session:%s", sessionID.String())
}

// tokenHashKey generates Redis key for token hash lookup
func (s *SessionStorage) tokenHashKey(tokenHash string) string {
	return fmt.Sprintf("token:%s", tokenHash)
}

// userSessionsKey generates Redis key for user sessions set
func (s *SessionStorage) userSessionsKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s:sessions", userID.String())
}

// Set stores a session in Redis
func (s *SessionStorage) Set(ctx context.Context, session *entity.Session) error {
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("session already expired")
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	sessionKey := s.sessionKey(session.ID)
	if err := s.client.Set(ctx, sessionKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	tokenHashKey := s.tokenHashKey(session.TokenHash)
	if err := s.client.Set(ctx, tokenHashKey, session.ID.String(), ttl).Err(); err != nil {
		return fmt.Errorf("failed to store token hash mapping: %w", err)
	}

	userSessionsKey := s.userSessionsKey(session.UserID)
	if err := s.client.SAdd(ctx, userSessionsKey, session.ID.String()).Err(); err != nil {
		return fmt.Errorf("failed to add session to user set: %w", err)
	}

	if err := s.client.Expire(ctx, userSessionsKey, s.sessionTTL+24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set expiration on user sessions: %w", err)
	}

	return nil
}

// Get retrieves a session by ID
func (s *SessionStorage) Get(ctx context.Context, sessionID uuid.UUID) (*entity.Session, error) {
	sessionKey := s.sessionKey(sessionID)
	data, err := s.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session entity.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// GetByTokenHash retrieves a session by token hash
func (s *SessionStorage) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error) {
	// Get session ID from token hash
	tokenHashKey := s.tokenHashKey(tokenHash)
	sessionIDStr, err := s.client.Get(ctx, tokenHashKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found for token")
		}
		return nil, fmt.Errorf("failed to get session ID from token: %w", err)
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	return s.Get(ctx, sessionID)
}

// GetByUserID retrieves all active sessions for a user
func (s *SessionStorage) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error) {
	userSessionsKey := s.userSessionsKey(userID)
	sessionIDs, err := s.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	sessions := make([]*entity.Session, 0, len(sessionIDs))
	for _, sessionIDStr := range sessionIDs {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			continue // Skip invalid IDs
		}

		session, err := s.Get(ctx, sessionID)
		if err != nil {
			// Session expired or deleted, remove from set
			s.client.SRem(ctx, userSessionsKey, sessionIDStr)
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// UpdateLastActivity updates the last activity timestamp and refreshes TTL
func (s *SessionStorage) UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.UpdateActivity()

	// Re-store with updated timestamp
	return s.Set(ctx, session)
}

// Delete removes a session from Redis
func (s *SessionStorage) Delete(ctx context.Context, sessionID uuid.UUID) error {
	// Get session first to clean up related keys
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	// Delete session data
	sessionKey := s.sessionKey(sessionID)
	if err := s.client.Del(ctx, sessionKey).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Delete token hash mapping
	tokenHashKey := s.tokenHashKey(session.TokenHash)
	if err := s.client.Del(ctx, tokenHashKey).Err(); err != nil {
		return fmt.Errorf("failed to delete token hash: %w", err)
	}

	// Remove from user sessions set
	userSessionsKey := s.userSessionsKey(session.UserID)
	if err := s.client.SRem(ctx, userSessionsKey, sessionID.String()).Err(); err != nil {
		return fmt.Errorf("failed to remove session from user set: %w", err)
	}

	return nil
}

// DeleteByUserID removes all sessions for a user
func (s *SessionStorage) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	// Get all user sessions
	sessions, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Delete each session
	for _, session := range sessions {
		if err := s.Delete(ctx, session.ID); err != nil {
			// Log error but continue
			continue
		}
	}

	// Delete user sessions set
	userSessionsKey := s.userSessionsKey(userID)
	if err := s.client.Del(ctx, userSessionsKey).Err(); err != nil {
		return fmt.Errorf("failed to delete user sessions set: %w", err)
	}

	return nil
}

// Exists checks if a session exists
func (s *SessionStorage) Exists(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	sessionKey := s.sessionKey(sessionID)
	result, err := s.client.Exists(ctx, sessionKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return result > 0, nil
}

// CountActiveByUserID counts active sessions for a user
func (s *SessionStorage) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	userSessionsKey := s.userSessionsKey(userID)
	count, err := s.client.SCard(ctx, userSessionsKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count user sessions: %w", err)
	}

	return int(count), nil
}

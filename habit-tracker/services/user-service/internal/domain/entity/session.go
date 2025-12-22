package entity

import (
	"net"
	"time"

	"github.com/google/uuid"
)

// Session represents a user session
type Session struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	TokenHash      string    `json:"-" db:"token_hash"` // Never expose in JSON
	IPAddress      *net.IP   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      *string   `json:"user_agent,omitempty" db:"user_agent"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	LastActivityAt time.Time `json:"last_activity_at" db:"last_activity_at"`
}

// SessionCreate represents data needed to create a new session
type SessionCreate struct {
	UserID    uuid.UUID
	TokenHash string
	IPAddress *net.IP
	UserAgent *string
	ExpiresAt time.Time
}

// SessionResponse represents session data for API responses
type SessionResponse struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	IPAddress      *string   `json:"ip_address,omitempty"`
	UserAgent      *string   `json:"user_agent,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
}

// ToResponse converts Session to SessionResponse
func (s *Session) ToResponse() *SessionResponse {
	resp := &SessionResponse{
		ID:             s.ID,
		UserID:         s.UserID,
		UserAgent:      s.UserAgent,
		ExpiresAt:      s.ExpiresAt,
		CreatedAt:      s.CreatedAt,
		LastActivityAt: s.LastActivityAt,
	}

	if s.IPAddress != nil {
		ipStr := s.IPAddress.String()
		resp.IPAddress = &ipStr
	}

	return resp
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsActive checks if the session is active (not expired)
func (s *Session) IsActive() bool {
	return !s.IsExpired()
}

// UpdateActivity updates the last activity timestamp
func (s *Session) UpdateActivity() {
	s.LastActivityAt = time.Now()
}

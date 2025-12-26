package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	pb "api-gateway/proto/user/v1"
)

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	SessionIDKey contextKey = "sessionID"
)

// AuthMiddleware wraps gRPC client for token validation
type AuthMiddleware struct {
	userClient pb.UserServiceClient
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(userClient pb.UserServiceClient) *AuthMiddleware {
	return &AuthMiddleware{
		userClient: userClient,
	}
}

// Auth validates JWT token from Authorization header
func (m *AuthMiddleware) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token == "" {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		validateReq := &pb.ValidateTokenRequest{
			AccessToken: token,
		}

		resp, err := m.userClient.ValidateToken(ctx, validateReq)
		if err != nil {
			http.Error(w, "Failed to validate token", http.StatusUnauthorized)
			return
		}

		if !resp.Valid {
			errorMsg := "Invalid token"
			if resp.Error != nil {
				errorMsg = *resp.Error
			}
			http.Error(w, errorMsg, http.StatusUnauthorized)
			return
		}

		if resp.UserId == nil {
			http.Error(w, "Missing user ID in token", http.StatusUnauthorized)
			return
		}

		if resp.SessionId == nil {
			http.Error(w, "Missing session ID in token", http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(r.Context(), UserIDKey, *resp.UserId)
		ctx = context.WithValue(ctx, SessionIDKey, *resp.SessionId)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserID extracts user ID from request context
func GetUserID(r *http.Request) string {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// GetSessionID extracts session ID from request context
func GetSessionID(r *http.Request) string {
	sessionID, ok := r.Context().Value(SessionIDKey).(string)
	if !ok {
		return ""
	}
	return sessionID
}

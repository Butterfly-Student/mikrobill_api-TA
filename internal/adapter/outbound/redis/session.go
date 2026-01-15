package redis_outbound_adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SessionData represents the session information stored in Redis
type SessionData struct {
	UserID       uuid.UUID  `json:"user_id"`
	TenantID     *uuid.UUID `json:"tenant_id,omitempty"`
	UserRole     string     `json:"user_role"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	IsSuperAdmin bool       `json:"is_superadmin"`
	IPAddress    string     `json:"ip_address,omitempty"`
	UserAgent    string     `json:"user_agent,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
}

// SessionAdapter manages session operations in Redis
type SessionAdapter struct {
	client *redis.Client
}

// NewSessionAdapter creates a new session adapter
func NewSessionAdapter(client *redis.Client) *SessionAdapter {
	return &SessionAdapter{
		client: client,
	}
}

// sessionKey generates the Redis key for a session
func (s *SessionAdapter) sessionKey(tokenHash string) string {
	return fmt.Sprintf("session:%s", tokenHash)
}

// userSessionsKey generates the Redis key for user's active sessions
func (s *SessionAdapter) userSessionsKey(userID uuid.UUID) string {
	return fmt.Sprintf("user_sessions:%s", userID.String())
}

// StoreSession stores a session in Redis
func (s *SessionAdapter) StoreSession(ctx context.Context, tokenHash string, session *SessionData) error {
	key := s.sessionKey(tokenHash)

	// Serialize session data
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Calculate TTL
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("session already expired")
	}

	// Store in Redis with expiration
	if err := s.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	// Add to user's active sessions set
	userKey := s.userSessionsKey(session.UserID)
	if err := s.client.SAdd(ctx, userKey, tokenHash).Err(); err != nil {
		return fmt.Errorf("failed to add session to user set: %w", err)
	}

	// Set expiration on user sessions set
	if err := s.client.Expire(ctx, userKey, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set expiration on user sessions: %w", err)
	}

	return nil
}

// GetSession retrieves a session from Redis
func (s *SessionAdapter) GetSession(ctx context.Context, tokenHash string) (*SessionData, error) {
	key := s.sessionKey(tokenHash)

	data, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		_ = s.DeleteSession(ctx, tokenHash)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// DeleteSession removes a session from Redis
func (s *SessionAdapter) DeleteSession(ctx context.Context, tokenHash string) error {
	// Get session first to remove from user's set
	session, err := s.GetSession(ctx, tokenHash)
	if err == nil && session != nil {
		userKey := s.userSessionsKey(session.UserID)
		_ = s.client.SRem(ctx, userKey, tokenHash).Err()
	}

	key := s.sessionKey(tokenHash)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// RefreshSession extends the session expiration
func (s *SessionAdapter) RefreshSession(ctx context.Context, tokenHash string, newExpiration time.Time) error {
	session, err := s.GetSession(ctx, tokenHash)
	if err != nil {
		return err
	}

	session.ExpiresAt = newExpiration

	// Update in Redis
	return s.StoreSession(ctx, tokenHash, session)
}

// GetUserSessions retrieves all active sessions for a user
func (s *SessionAdapter) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*SessionData, error) {
	userKey := s.userSessionsKey(userID)

	// Get all session token hashes
	tokenHashes, err := s.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	sessions := make([]*SessionData, 0, len(tokenHashes))
	for _, hash := range tokenHashes {
		session, err := s.GetSession(ctx, hash)
		if err == nil {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// RevokeUserSessions revokes all sessions for a user
func (s *SessionAdapter) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	userKey := s.userSessionsKey(userID)

	// Get all session token hashes
	tokenHashes, err := s.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Delete all sessions
	for _, hash := range tokenHashes {
		_ = s.DeleteSession(ctx, hash)
	}

	// Delete user sessions set
	if err := s.client.Del(ctx, userKey).Err(); err != nil {
		return fmt.Errorf("failed to delete user sessions set: %w", err)
	}

	return nil
}

// CleanExpiredSessions removes expired sessions (should be run periodically)
func (s *SessionAdapter) CleanExpiredSessions(ctx context.Context) error {
	// This is mainly a safety net, as Redis TTL handles most cleanup
	// But we can implement a scan-based cleanup for orphaned user session sets

	iter := s.client.Scan(ctx, 0, "user_sessions:*", 0).Iterator()
	for iter.Next(ctx) {
		userKey := iter.Val()

		// Get all sessions for this user
		tokenHashes, err := s.client.SMembers(ctx, userKey).Result()
		if err != nil {
			continue
		}

		validSessions := 0
		for _, hash := range tokenHashes {
			_, err := s.GetSession(ctx, hash)
			if err == nil {
				validSessions++
			} else {
				// Remove invalid session from set
				_ = s.client.SRem(ctx, userKey, hash).Err()
			}
		}

		// If no valid sessions, delete the set
		if validSessions == 0 {
			_ = s.client.Del(ctx, userKey).Err()
		}
	}

	return iter.Err()
}

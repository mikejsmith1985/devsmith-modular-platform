package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore manages session storage in Redis
type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

// Session represents a user session
type Session struct {
	SessionID      string                 `json:"session_id"`
	UserID         int                    `json:"user_id"`
	GitHubUsername string                 `json:"github_username"`
	GitHubToken    string                 `json:"github_token"`
	CreatedAt      time.Time              `json:"created_at"`
	LastAccessedAt time.Time              `json:"last_accessed_at"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// NewRedisStore creates a new Redis session store
func NewRedisStore(addr string, ttl time.Duration) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "", // No password set
		DB:           0,  // Use default DB
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisStore{client: client, ttl: ttl}, nil
}

// GenerateSessionID creates a cryptographically secure session ID
func GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// Create stores a new session in Redis
func (s *RedisStore) Create(ctx context.Context, session *Session) (string, error) {
	now := time.Now()
	session.CreatedAt = now
	session.LastAccessedAt = now

	if session.SessionID == "" {
		sessionID, err := GenerateSessionID()
		if err != nil {
			return "", err
		}
		session.SessionID = sessionID
	}

	data, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("marshal session: %w", err)
	}

	key := fmt.Sprintf("session:%s", session.SessionID)
	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return "", fmt.Errorf("redis set: %w", err)
	}

	return session.SessionID, nil
}

// Get retrieves a session from Redis
func (s *RedisStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Session not found
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	// Update last accessed time
	session.LastAccessedAt = time.Now()
	if err := s.Update(ctx, &session); err != nil {
		// Log but don't fail - session still valid
	}

	return &session, nil
}

// Update refreshes a session in Redis
func (s *RedisStore) Update(ctx context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	key := fmt.Sprintf("session:%s", session.SessionID)
	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

// Delete removes a session from Redis
func (s *RedisStore) Delete(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// Exists checks if a session exists without retrieving it
func (s *RedisStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists: %w", err)
	}
	return result > 0, nil
}

// RefreshTTL extends the expiration time of a session
func (s *RedisStore) RefreshTTL(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	if err := s.client.Expire(ctx, key, s.ttl).Err(); err != nil {
		return fmt.Errorf("redis expire: %w", err)
	}
	return nil
}

// StoreOAuthState stores an OAuth state parameter in Redis with expiration
func (s *RedisStore) StoreOAuthState(ctx context.Context, state string, ttl time.Duration) error {
	key := fmt.Sprintf("oauth_state:%s", state)
	if err := s.client.Set(ctx, key, "valid", ttl).Err(); err != nil {
		return fmt.Errorf("redis set oauth state: %w", err)
	}
	return nil
}

// ValidateOAuthState checks if an OAuth state exists in Redis and deletes it (one-time use)
func (s *RedisStore) ValidateOAuthState(ctx context.Context, state string) (bool, error) {
	key := fmt.Sprintf("oauth_state:%s", state)

	// Check if state exists
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return false, nil // State doesn't exist or error occurred
	}

	if val != "valid" {
		return false, nil
	}

	// Delete state (one-time use for CSRF protection)
	s.client.Del(ctx, key)
	return true, nil
}

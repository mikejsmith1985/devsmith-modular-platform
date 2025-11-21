package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedisStore_Create verifies session creation
func TestRedisStore_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	store, err := NewRedisStore("localhost:6379", 7*24*time.Hour)
	require.NoError(t, err)
	defer store.Close()

	session := &Session{
		UserID:         123,
		GitHubUsername: "testuser",
			GitHubToken:    "gh*_test123", // Test token (masked prefix)
		Metadata:       map[string]interface{}{"ip": "127.0.0.1"},
	}

	sessionID, err := store.Create(context.Background(), session)
	require.NoError(t, err)
	assert.NotEmpty(t, sessionID)
	assert.Equal(t, sessionID, session.SessionID)
}

// TestRedisStore_GetAndUpdate verifies session retrieval and updates
func TestRedisStore_GetAndUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	store, err := NewRedisStore("localhost:6379", 7*24*time.Hour)
	require.NoError(t, err)
	defer store.Close()

	// Create session
	original := &Session{
		UserID:         456,
		GitHubUsername: "anotheruser",
			GitHubToken:    "gh*_test456", // Test token (masked prefix)
	}

	sessionID, err := store.Create(context.Background(), original)
	require.NoError(t, err)

	// Retrieve session
	retrieved, err := store.Get(context.Background(), sessionID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, original.UserID, retrieved.UserID)
	assert.Equal(t, original.GitHubUsername, retrieved.GitHubUsername)

	// Update session
	retrieved.Metadata = map[string]interface{}{"updated": true}
	err = store.Update(context.Background(), retrieved)
	require.NoError(t, err)

	// Verify update
	updated, err := store.Get(context.Background(), sessionID)
	require.NoError(t, err)
	assert.Equal(t, true, updated.Metadata["updated"])
}

// TestRedisStore_Delete verifies session deletion
func TestRedisStore_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	store, err := NewRedisStore("localhost:6379", 7*24*time.Hour)
	require.NoError(t, err)
	defer store.Close()

	// Create session
	session := &Session{
		UserID:         789,
		GitHubUsername: "deleteuser",
	}

	sessionID, err := store.Create(context.Background(), session)
	require.NoError(t, err)

	// Delete session
	err = store.Delete(context.Background(), sessionID)
	require.NoError(t, err)

	// Verify deletion
	retrieved, err := store.Get(context.Background(), sessionID)
	require.NoError(t, err)
	assert.Nil(t, retrieved, "Session should be deleted")
}

// TestRedisStore_NonexistentSession verifies behavior for missing sessions
func TestRedisStore_NonexistentSession(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	store, err := NewRedisStore("localhost:6379", 7*24*time.Hour)
	require.NoError(t, err)
	defer store.Close()

	retrieved, err := store.Get(context.Background(), "nonexistent-session-id")
	require.NoError(t, err)
	assert.Nil(t, retrieved, "Should return nil for nonexistent session")
}

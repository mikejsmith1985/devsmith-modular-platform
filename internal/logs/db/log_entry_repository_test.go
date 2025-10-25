package db

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

func TestLogEntry_Struct(t *testing.T) {
	entry := &models.LogEntry{
		ID:      1,
		UserID:  42,
		Service: "portal",
		Level:   "info",
		Message: "Test log message",
	}

	assert.Equal(t, int64(1), entry.ID)
	assert.Equal(t, int64(42), entry.UserID)
	assert.Equal(t, "portal", entry.Service)
	assert.Equal(t, "info", entry.Level)
	assert.Equal(t, "Test log message", entry.Message)
}

func TestNewLogEntryRepository_Creation(t *testing.T) {
	repo := NewLogEntryRepository(nil)
	assert.NotNil(t, repo)
}

func TestLogEntryRepository_GetMetadataValue_ValidMetadata(t *testing.T) {
	repo := NewLogEntryRepository(nil)

	metadata := map[string]interface{}{
		"ip":        "192.168.1.1",
		"timestamp": 1234567890,
		"error":     "connection refused",
	}
	metadataJSON, err := json.Marshal(metadata)
	require.NoError(t, err)

	ip, err := repo.GetMetadataValue(metadataJSON, "ip")
	require.NoError(t, err)
	assert.Equal(t, "192.168.1.1", ip)

	timestamp, err := repo.GetMetadataValue(metadataJSON, "timestamp")
	require.NoError(t, err)
	assert.Equal(t, float64(1234567890), timestamp)

	errMsg, err := repo.GetMetadataValue(metadataJSON, "error")
	require.NoError(t, err)
	assert.Equal(t, "connection refused", errMsg)
}

func TestLogEntryRepository_GetMetadataValue_NilMetadata(t *testing.T) {
	repo := NewLogEntryRepository(nil)

	value, err := repo.GetMetadataValue(nil, "key")
	require.NoError(t, err)
	assert.Nil(t, value)
}

func TestLogEntryRepository_GetMetadataValue_EmptyMetadata(t *testing.T) {
	repo := NewLogEntryRepository(nil)

	value, err := repo.GetMetadataValue([]byte(""), "key")
	require.NoError(t, err)
	assert.Nil(t, value)
}

func TestLogEntryRepository_GetMetadataValue_MissingKey(t *testing.T) {
	repo := NewLogEntryRepository(nil)

	metadata := map[string]interface{}{
		"existing": "value",
	}
	metadataJSON, err := json.Marshal(metadata)
	require.NoError(t, err)

	value, err := repo.GetMetadataValue(metadataJSON, "missing")
	require.NoError(t, err)
	assert.Nil(t, value)
}

func TestLogEntryRepository_GetMetadataValue_InvalidJSON(t *testing.T) {
	repo := NewLogEntryRepository(nil)

	invalidJSON := []byte(`{invalid json}`)

	_, err := repo.GetMetadataValue(invalidJSON, "key")
	require.Error(t, err)
}

func TestLogEntry_ValidLogLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		entry := &models.LogEntry{
			Level: level,
		}
		assert.Equal(t, level, entry.Level)
	}
}

func TestLogEntry_ValidServices(t *testing.T) {
	services := []string{"portal", "review", "logging", "analytics", "build"}

	for _, service := range services {
		entry := &models.LogEntry{
			Service: service,
		}
		assert.Equal(t, service, entry.Service)
	}
}

func TestLogEntry_MetadataHandling(t *testing.T) {
	complexMetadata := map[string]interface{}{
		"user_ip":     "192.168.1.100",
		"user_agent":  "Mozilla/5.0",
		"request_id":  "req-123456",
		"duration_ms": 245,
		"tags":        []string{"auth", "login", "success"},
		"nested": map[string]interface{}{
			"error": "none",
			"code":  200,
		},
	}

	metadataJSON, err := json.Marshal(complexMetadata)
	require.NoError(t, err)

	entry := &models.LogEntry{
		ID:       1,
		UserID:   10,
		Service:  "portal",
		Level:    "info",
		Message:  "User authentication successful",
		Metadata: metadataJSON,
	}

	assert.Equal(t, int64(1), entry.ID)
	assert.Equal(t, int64(10), entry.UserID)
	assert.Equal(t, "portal", entry.Service)
	assert.Equal(t, "info", entry.Level)
	assert.Equal(t, "User authentication successful", entry.Message)
	assert.NotNil(t, entry.Metadata)
	assert.Len(t, entry.Metadata, len(metadataJSON))
}

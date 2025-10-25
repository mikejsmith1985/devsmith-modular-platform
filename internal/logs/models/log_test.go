package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogEntry_BasicStruct(t *testing.T) {
	now := time.Now()
	metadata := []byte(`{"key":"value"}`)
	entry := LogEntry{
		CreatedAt: now,
		Service:   "portal",
		Level:     "INFO",
		Message:   "User logged in",
		Metadata:  metadata,
		ID:        1,
		UserID:    123,
	}

	assert.Equal(t, now, entry.CreatedAt)
	assert.Equal(t, "portal", entry.Service)
	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "User logged in", entry.Message)
	assert.Equal(t, metadata, entry.Metadata)
	assert.Equal(t, int64(1), entry.ID)
	assert.Equal(t, int64(123), entry.UserID)
}

func TestLogEntry_ZeroValues(t *testing.T) {
	entry := LogEntry{}

	assert.Equal(t, time.Time{}, entry.CreatedAt)
	assert.Equal(t, "", entry.Service)
	assert.Equal(t, "", entry.Level)
	assert.Equal(t, "", entry.Message)
	assert.Len(t, entry.Metadata, 0)
	assert.Equal(t, int64(0), entry.ID)
	assert.Equal(t, int64(0), entry.UserID)
}

func TestLogEntry_DifferentLevels(t *testing.T) {
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	for _, level := range levels {
		entry := LogEntry{
			Level: level,
		}
		assert.Equal(t, level, entry.Level)
	}
}

func TestLogEntry_DifferentServices(t *testing.T) {
	services := []string{"portal", "review", "logs", "analytics"}

	for _, service := range services {
		entry := LogEntry{
			Service: service,
		}
		assert.Equal(t, service, entry.Service)
	}
}

func TestLogEntry_Timestamps(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Second)

	entry1 := LogEntry{CreatedAt: now}
	entry2 := LogEntry{CreatedAt: later}

	assert.True(t, entry2.CreatedAt.After(entry1.CreatedAt))
}

func TestLogEntry_FieldIndependence(t *testing.T) {
	entry := LogEntry{}

	entry.Service = "test"
	assert.Equal(t, "test", entry.Service)
	assert.Equal(t, "", entry.Level)

	entry.Level = "ERROR"
	assert.Equal(t, "test", entry.Service)
	assert.Equal(t, "ERROR", entry.Level)

	entry.ID = 999
	assert.Equal(t, "test", entry.Service)
	assert.Equal(t, "ERROR", entry.Level)
	assert.Equal(t, int64(999), entry.ID)
}

func TestLogEntry_Metadata(t *testing.T) {
	metadata := []byte(`{"user_agent":"Mozilla/5.0"}`)
	entry := LogEntry{
		Metadata: metadata,
	}

	assert.NotNil(t, entry.Metadata)
	assert.Greater(t, len(entry.Metadata), 0)
}

func TestLogEntry_LargeID(t *testing.T) {
	entry := LogEntry{
		ID:     9223372036854775807, // max int64
		UserID: 9223372036854775807,
	}

	assert.Equal(t, int64(9223372036854775807), entry.ID)
	assert.Equal(t, int64(9223372036854775807), entry.UserID)
}

func TestLogEntry_MultipleInstances(t *testing.T) {
	entry1 := LogEntry{ID: 1, Service: "service1"}
	entry2 := LogEntry{ID: 2, Service: "service2"}
	entry3 := LogEntry{ID: 3, Service: "service3"}

	assert.NotEqual(t, entry1.ID, entry2.ID)
	assert.NotEqual(t, entry2.ID, entry3.ID)
	assert.NotEqual(t, entry1.Service, entry2.Service)
}

func TestLogEntry_EmptyMessage(t *testing.T) {
	entry := LogEntry{
		Message: "",
		Level:   "INFO",
	}

	assert.Empty(t, entry.Message)
	assert.Equal(t, "INFO", entry.Level)
}

func TestLogEntry_LongMessage(t *testing.T) {
	longMsg := "This is a very long message that contains a lot of text to test the message field"
	entry := LogEntry{
		Message: longMsg,
	}

	assert.Equal(t, longMsg, entry.Message)
	assert.Greater(t, len(entry.Message), 50)
}

func TestLogEntry_ServiceNames(t *testing.T) {
	testCases := []string{
		"portal",
		"review",
		"logs",
		"analytics",
		"custom-service",
	}

	for _, service := range testCases {
		entry := LogEntry{Service: service}
		assert.Equal(t, service, entry.Service)
	}
}

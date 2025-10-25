// Package models defines the data structures used in the logs service.
package models

import "time"

// LogEntry represents a log entry in the system.
// Fields are ordered to optimize memory alignment.
type LogEntry struct {
	CreatedAt time.Time `json:"created_at"` // 16 bytes
	Service   string    `json:"service"`    // 16 bytes
	Level     string    `json:"level"`      // 16 bytes
	Message   string    `json:"message"`    // 16 bytes
	Metadata  []byte    `json:"metadata"`   // 8 bytes
	Tags      []string  `json:"tags"`       // 24 bytes (slice)
	ID        int64     `json:"id"`         // 8 bytes
	UserID    int64     `json:"user_id"`    // 8 bytes
}

// Package models defines the data structures used in the logs service.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// CorrelationContext stores request context for tracing across services
type CorrelationContext struct {
	UserID        *int      `json:"user_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CorrelationID string    `json:"correlation_id"`   // Unique per request
	TraceID       string    `json:"trace_id"`         // OpenTelemetry trace ID
	SpanID        string    `json:"span_id"`          // OpenTelemetry span ID
	RequestID     string    `json:"request_id"`       // HTTP request ID
	Service       string    `json:"service"`          // Service that generated log
	Hostname      string    `json:"hostname"`         // Server hostname
	Environment   string    `json:"environment"`      // dev, staging, prod
	Version       string    `json:"version"`          // Service version
	Method        string    `json:"method,omitempty"` // HTTP method
	Path          string    `json:"path,omitempty"`   // HTTP path
	RemoteAddr    string    `json:"remote_addr,omitempty"`
	SessionID     string    `json:"session_id,omitempty"`
}

// Value implements driver.Valuer for database storage
func (cc *CorrelationContext) Value() (driver.Value, error) {
	return json.Marshal(cc)
}

// Scan implements sql.Scanner for database retrieval
func (cc *CorrelationContext) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(bytes, &cc)
}

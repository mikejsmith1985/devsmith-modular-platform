package logger

import (
	"encoding/json"
	"time"
)

// LogEntry represents a single log entry to be sent to the logs service.
type LogEntry struct {
	CreatedAt time.Time              `json:"created_at"`
	Service   string                 `json:"service"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
	Tags      []string               `json:"tags"`
}

// MarshalJSON converts LogEntry to JSON for API transmission.
func (e *LogEntry) MarshalJSON() ([]byte, error) {
	type Alias LogEntry
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
	}{
		Alias:     (*Alias)(e),
		CreatedAt: e.CreatedAt.UTC().Format(time.RFC3339Nano),
	})
}

// LogRequest represents a batch of logs to send to the logging service.
type LogRequest struct {
	Logs []*LogEntry `json:"logs"`
}

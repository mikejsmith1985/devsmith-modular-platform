package search

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExportFormat defines the export format for search results.
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
)

// Exporter handles exporting search results in various formats.
type Exporter struct {
	format ExportFormat
}

// NewExporter creates a new exporter for the given format.
func NewExporter(format ExportFormat) *Exporter {
	return &Exporter{format: format}
}

// Export converts log entries to the specified format.
func (e *Exporter) Export(logs []*LogEntry) ([]byte, error) {
	switch e.format {
	case ExportFormatCSV:
		return e.exportCSV(logs)
	case ExportFormatJSON:
		return e.exportJSON(logs)
	default:
		return nil, ErrInvalidFormat
	}
}

// exportCSV exports logs to CSV format.
func (e *Exporter) exportCSV(logs []*LogEntry) ([]byte, error) {
	// Placeholder - will be implemented in GREEN phase
	return []byte{}, nil
}

// exportJSON exports logs to JSON format.
func (e *Exporter) exportJSON(logs []*LogEntry) ([]byte, error) {
	// Placeholder - will be implemented in GREEN phase
	return []byte{}, nil
}

// Error variables
var (
	ErrInvalidFormat = NewError("invalid_format", "unsupported export format")
)

// TestExport_CSVFormat tests CSV export functionality.
func TestExport_CSVFormat(t *testing.T) {
	logs := []*LogEntry{
		{
			ID:        1,
			Service:   "auth",
			Level:     "error",
			Message:   "authentication failed",
			Timestamp: time.Date(2025, 10, 25, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:        2,
			Service:   "portal",
			Level:     "warn",
			Message:   "slow database query",
			Timestamp: time.Date(2025, 10, 25, 12, 1, 0, 0, time.UTC),
		},
	}

	exporter := NewExporter(ExportFormatCSV)
	data, err := exporter.Export(logs)

	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify CSV format: should have header row + 2 data rows
	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(records), 2) // At least header + 1 row
}

// TestExport_JSONFormat tests JSON export functionality.
func TestExport_JSONFormat(t *testing.T) {
	logs := []*LogEntry{
		{
			ID:        1,
			Service:   "auth",
			Level:     "error",
			Message:   "authentication failed",
			Timestamp: time.Date(2025, 10, 25, 12, 0, 0, 0, time.UTC),
		},
	}

	exporter := NewExporter(ExportFormatJSON)
	data, err := exporter.Export(logs)

	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify JSON format
	var result interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
}

// TestExport_CSVHeaders tests that CSV headers are correct.
func TestExport_CSVHeaders(t *testing.T) {
	logs := []*LogEntry{}

	exporter := NewExporter(ExportFormatCSV)
	data, err := exporter.Export(logs)

	require.NoError(t, err)
	assert.NotEmpty(t, data)

	reader := csv.NewReader(bytes.NewReader(data))
	header, err := reader.Read()
	require.NoError(t, err)

	// Should have columns: id, service, level, message, timestamp
	assert.GreaterOrEqual(t, len(header), 5)
}

// TestExport_JSONStructure tests that JSON has expected structure.
func TestExport_JSONStructure(t *testing.T) {
	logs := []*LogEntry{
		{
			ID:        1,
			Service:   "auth",
			Level:     "error",
			Message:   "failed",
			Timestamp: time.Date(2025, 10, 25, 12, 0, 0, 0, time.UTC),
		},
	}

	exporter := NewExporter(ExportFormatJSON)
	data, err := exporter.Export(logs)

	require.NoError(t, err)

	var result []map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Len(t, result, 1)
	assert.Contains(t, result[0], "id")
	assert.Contains(t, result[0], "service")
	assert.Contains(t, result[0], "level")
	assert.Contains(t, result[0], "message")
}

// TestExport_CSVEscaping tests that CSV properly escapes special characters.
func TestExport_CSVEscaping(t *testing.T) {
	logs := []*LogEntry{
		{
			ID:        1,
			Service:   "auth",
			Level:     "error",
			Message:   "message with \"quotes\" and, commas",
			Timestamp: time.Date(2025, 10, 25, 12, 0, 0, 0, time.UTC),
		},
	}

	exporter := NewExporter(ExportFormatCSV)
	data, err := exporter.Export(logs)

	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Should be able to parse without errors
	reader := csv.NewReader(bytes.NewReader(data))
	_, err = reader.ReadAll()
	require.NoError(t, err)
}

// TestExport_JSONNullHandling tests that JSON properly handles null/empty values.
func TestExport_JSONNullHandling(t *testing.T) {
	logs := []*LogEntry{
		{
			ID:        1,
			Service:   "",
			Level:     "error",
			Message:   "",
			Timestamp: time.Date(2025, 10, 25, 12, 0, 0, 0, time.UTC),
		},
	}

	exporter := NewExporter(ExportFormatJSON)
	data, err := exporter.Export(logs)

	require.NoError(t, err)

	var result []map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Len(t, result, 1)
}

// TestExport_LargeDataset tests exporting a large dataset.
func TestExport_LargeDataset(t *testing.T) {
	// Create 1000 log entries
	logs := make([]*LogEntry, 1000)
	for i := 0; i < 1000; i++ {
		logs[i] = &LogEntry{
			ID:        int64(i + 1),
			Service:   "service",
			Level:     "info",
			Message:   "message",
			Timestamp: time.Now(),
		}
	}

	exporter := NewExporter(ExportFormatCSV)
	data, err := exporter.Export(logs)

	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Should be readable as valid CSV
	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Should have header + 1000 data rows
	assert.GreaterOrEqual(t, len(records), 1000)
}

// TestExport_InvalidFormat tests handling of invalid export format.
func TestExport_InvalidFormat(t *testing.T) {
	logs := []*LogEntry{}
	exporter := NewExporter(ExportFormat("xml")) // Unsupported format

	_, err := exporter.Export(logs)
	assert.Error(t, err)
}

// TestExport_EmptyResults tests exporting empty result set.
func TestExport_EmptyResults(t *testing.T) {
	logs := []*LogEntry{}

	// CSV with empty results
	exporter := NewExporter(ExportFormatCSV)
	data, err := exporter.Export(logs)
	require.NoError(t, err)
	assert.NotEmpty(t, data) // Should at least have headers

	// JSON with empty results
	exporter = NewExporter(ExportFormatJSON)
	data, err = exporter.Export(logs)
	require.NoError(t, err)
	assert.NotEmpty(t, data) // Should have empty array

	var result []interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestExport_TimestampFormatting tests that timestamps are formatted correctly.
func TestExport_TimestampFormatting(t *testing.T) {
	timestamp := time.Date(2025, 10, 25, 12, 30, 45, 0, time.UTC)
	logs := []*LogEntry{
		{
			ID:        1,
			Service:   "auth",
			Level:     "error",
			Message:   "test",
			Timestamp: timestamp,
		},
	}

	exporter := NewExporter(ExportFormatJSON)
	data, err := exporter.Export(logs)

	require.NoError(t, err)

	var result []map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Contains(t, result[0], "timestamp")
	// Timestamp should be formatted as RFC3339 or similar
	assert.NotEmpty(t, result[0]["timestamp"])
}

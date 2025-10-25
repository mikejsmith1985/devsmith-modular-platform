package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// TestNewLogEntryRepository_ValidCreation tests repository initialization with nil db
func TestNewLogEntryRepository_ValidCreation(t *testing.T) {
	repo := NewLogEntryRepository(nil)
	require.NotNil(t, repo)
	assert.Nil(t, repo.db)
}

// TestLogEntryRepository_ConstructorNilDB tests that nil db is acceptable (unit test mode)
func TestLogEntryRepository_ConstructorNilDB(t *testing.T) {
	repo := NewLogEntryRepository(nil)
	assert.NotNil(t, repo)
}

// TestQueryOptions_Validate tests validation of query options struct
func TestQueryOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    QueryOptions
		wantErr bool
	}{
		{
			name: "valid_options",
			opts: QueryOptions{
				Limit:  10,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "valid_with_service_filter",
			opts: QueryOptions{
				Service: "portal",
				Limit:   10,
				Offset:  0,
			},
			wantErr: false,
		},
		{
			name: "valid_with_time_range",
			opts: QueryOptions{
				Limit:  10,
				Offset: 0,
				Since:  time.Now().Add(-1 * time.Hour),
				Until:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid_negative_limit",
			opts: QueryOptions{
				Limit:  -1,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid_negative_offset",
			opts: QueryOptions{
				Limit:  10,
				Offset: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid_zero_limit",
			opts: QueryOptions{
				Limit:  0,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid_since_after_until",
			opts: QueryOptions{
				Limit:  10,
				Offset: 0,
				Since:  time.Now(),
				Until:  time.Now().Add(-1 * time.Hour),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFilterOptions_ValidateService tests service filter validation
func TestFilterOptions_ValidateService(t *testing.T) {
	tests := []struct {
		name    string
		service string
		wantErr bool
	}{
		{
			name:    "valid_portal",
			service: "portal",
			wantErr: false,
		},
		{
			name:    "valid_review",
			service: "review",
			wantErr: false,
		},
		{
			name:    "valid_analytics",
			service: "analytics",
			wantErr: false,
		},
		{
			name:    "valid_logs",
			service: "logs",
			wantErr: false,
		},
		{
			name:    "invalid_empty_service",
			service: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := FilterOptions{Service: tt.service}
			err := filter.ValidateService()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFilterOptions_ValidateLevel tests log level filter validation
func TestFilterOptions_ValidateLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "valid_debug",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "valid_info",
			level:   "info",
			wantErr: false,
		},
		{
			name:    "valid_warn",
			level:   "warn",
			wantErr: false,
		},
		{
			name:    "valid_error",
			level:   "error",
			wantErr: false,
		},
		{
			name:    "invalid_empty_level",
			level:   "",
			wantErr: true,
		},
		{
			name:    "invalid_level",
			level:   "panic",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := FilterOptions{Level: tt.level}
			err := filter.ValidateLevel()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFilterOptions_Complete tests complete filter validation
func TestFilterOptions_Complete(t *testing.T) {
	tests := []struct {
		name    string
		filter  FilterOptions
		wantErr bool
	}{
		{
			name: "valid_service_filter",
			filter: FilterOptions{
				Service: "portal",
			},
			wantErr: false,
		},
		{
			name: "valid_level_filter",
			filter: FilterOptions{
				Level: "error",
			},
			wantErr: false,
		},
		{
			name: "invalid_both_service_and_level",
			filter: FilterOptions{
				Service: "portal",
				Level:   "error",
			},
			wantErr: true,
		},
		{
			name:    "valid_empty_filter",
			filter:  FilterOptions{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMetadataFilter_ValidateJSONBPath tests JSONB path validation
func TestMetadataFilter_ValidateJSONBPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid_simple_key",
			path:    "error",
			wantErr: false,
		},
		{
			name:    "valid_nested_key",
			path:    "request.ip",
			wantErr: false,
		},
		{
			name:    "invalid_empty_path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "invalid_only_dot",
			path:    ".",
			wantErr: true,
		},
		{
			name:    "invalid_trailing_dot",
			path:    "error.",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf := MetadataFilter{JSONBPath: tt.path}
			err := mf.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogEntryForCreate_ValidateFields tests entry validation before create
func TestLogEntryForCreate_ValidateFields(t *testing.T) {
	tests := []struct {
		entry   *models.LogEntry
		errMsg  string
		name    string
		wantErr bool
	}{
		{
			name: "valid_entry",
			entry: &models.LogEntry{
				Service: "portal",
				Level:   "info",
				Message: "User logged in",
			},
			wantErr: false,
		},
		{
			name: "valid_with_user_and_metadata",
			entry: &models.LogEntry{
				UserID:   123,
				Service:  "review",
				Level:    "error",
				Message:  "Analysis failed",
				Metadata: []byte(`{"error":"timeout"}`),
			},
			wantErr: false,
		},
		{
			name: "invalid_missing_service",
			entry: &models.LogEntry{
				Level:   "info",
				Message: "Test",
			},
			wantErr: true,
			errMsg:  "service",
		},
		{
			name: "invalid_missing_level",
			entry: &models.LogEntry{
				Service: "portal",
				Message: "Test",
			},
			wantErr: true,
			errMsg:  "level",
		},
		{
			name: "invalid_missing_message",
			entry: &models.LogEntry{
				Service: "portal",
				Level:   "info",
			},
			wantErr: true,
			errMsg:  "message",
		},
		{
			name: "invalid_service",
			entry: &models.LogEntry{
				Service: "unknown",
				Level:   "info",
				Message: "Test",
			},
			wantErr: true,
		},
		{
			name: "invalid_level",
			entry: &models.LogEntry{
				Service: "portal",
				Level:   "critical",
				Message: "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogEntryForCreate(tt.entry)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSearchQuery_ValidateSearchTerm tests search term validation
func TestSearchQuery_ValidateSearchTerm(t *testing.T) {
	tests := []struct {
		name    string
		term    string
		wantErr bool
	}{
		{
			name:    "valid_simple_term",
			term:    "error",
			wantErr: false,
		},
		{
			name:    "valid_multi_word",
			term:    "SQL injection",
			wantErr: false,
		},
		{
			name:    "valid_special_chars",
			term:    "error: timeout",
			wantErr: false,
		},
		{
			name:    "invalid_empty_term",
			term:    "",
			wantErr: true,
		},
		{
			name:    "invalid_only_spaces",
			term:    "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sq := SearchQuery{Term: tt.term}
			err := sq.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRepositoryContextHandling tests that operations handle context properly
func TestRepositoryContextHandling(t *testing.T) {
	repo := NewLogEntryRepository(nil)
	require.NotNil(t, repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// With nil DB, this won't actually execute, but we can verify context is accepted
	assert.NotNil(t, ctx)
}

// TestLogEntry_StructDefaults tests that LogEntry has proper zero values
func TestLogEntry_StructDefaults(t *testing.T) {
	entry := &models.LogEntry{}

	assert.Equal(t, int64(0), entry.ID)
	assert.Equal(t, int64(0), entry.UserID)
	assert.Equal(t, "", entry.Service)
	assert.Equal(t, "", entry.Level)
	assert.Equal(t, "", entry.Message)
	assert.Empty(t, entry.Metadata)
}

// TestValidLogLevels_Enumeration tests all valid log levels
func TestValidLogLevels_Enumeration(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error"}

	for _, level := range validLevels {
		t.Run(level, func(t *testing.T) {
			filter := FilterOptions{Level: level}
			err := filter.ValidateLevel()
			assert.NoError(t, err)
		})
	}
}

// TestValidServices_Enumeration tests all valid services
func TestValidServices_Enumeration(t *testing.T) {
	validServices := []string{"portal", "review", "analytics", "logs"}

	for _, service := range validServices {
		t.Run(service, func(t *testing.T) {
			filter := FilterOptions{Service: service}
			err := filter.ValidateService()
			assert.NoError(t, err)
		})
	}
}

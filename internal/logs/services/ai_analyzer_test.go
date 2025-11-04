package logs_services

import (
	"context"
	"testing"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/ai"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAIProvider mocks the AI provider for testing
type MockAIProvider struct {
	mock.Mock
}

func (m *MockAIProvider) Generate(ctx context.Context, req *ai.Request) (*ai.Response, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ai.Response), args.Error(1)
}

func (m *MockAIProvider) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAIProvider) GetModelInfo() *ai.ModelInfo {
	args := m.Called()
	return args.Get(0).(*ai.ModelInfo)
}

// TestNewAIAnalyzer tests analyzer creation
func TestNewAIAnalyzer(t *testing.T) {
	mockProvider := new(MockAIProvider)

	analyzer := NewAIAnalyzer(mockProvider)

	assert.NotNil(t, analyzer)
	assert.NotNil(t, analyzer.cache)
}

// TestAnalyze_DatabaseConnectionError tests analysis of database connection errors
func TestAnalyze_DatabaseConnectionError(t *testing.T) {
	mockProvider := new(MockAIProvider)
	analyzer := NewAIAnalyzer(mockProvider)

	logEntries := []logs_models.LogEntry{
		{
			ID:        1,
			Service:   "portal",
			Level:     "error",
			Message:   "connection refused to database at localhost:5432",
			Metadata:  []byte(`{"correlation_id": "req-123"}`),
			CreatedAt: time.Now(),
		},
	}

	expectedAnalysis := &AnalysisResult{
		RootCause:    "PostgreSQL database connection refused - server may be down or network unreachable",
		SuggestedFix: "1. Check if PostgreSQL is running: `systemctl status postgresql`\n2. Verify connection string in environment variables\n3. Check network connectivity to database host",
		Severity:     5,
		RelatedLogs:  []string{"req-123"},
		FixSteps: []string{
			"Verify PostgreSQL service is running",
			"Check database connection string configuration",
			"Test network connectivity to database host",
		},
	}

	mockProvider.On("Generate", mock.Anything, mock.MatchedBy(func(req *ai.Request) bool {
		return req.Prompt != "" && req.Model != ""
	})).Return(&ai.Response{
		Content:      `{"root_cause":"PostgreSQL database connection refused - server may be down or network unreachable","suggested_fix":"1. Check if PostgreSQL is running: ` + "`systemctl status postgresql`" + `\n2. Verify connection string in environment variables\n3. Check network connectivity to database host","severity":5,"related_logs":["req-123"],"fix_steps":["Verify PostgreSQL service is running","Check database connection string configuration","Test network connectivity to database host"]}`,
		Model:        "qwen2.5-coder:7b",
		FinishReason: "complete",
		ResponseTime: 2 * time.Second,
	}, nil)

	req := AnalysisRequest{
		LogEntries: logEntries,
		Context:    "error",
	}

	result, err := analyzer.Analyze(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedAnalysis.RootCause, result.RootCause)
	assert.Equal(t, expectedAnalysis.Severity, result.Severity)
	assert.Contains(t, result.SuggestedFix, "PostgreSQL")
	assert.Len(t, result.FixSteps, 3)

	mockProvider.AssertExpectations(t)
}

// TestAnalyze_AuthenticationFailure tests analysis of auth errors
func TestAnalyze_AuthenticationFailure(t *testing.T) {
	mockProvider := new(MockAIProvider)
	analyzer := NewAIAnalyzer(mockProvider)

	logEntries := []logs_models.LogEntry{
		{
			ID:        2,
			Service:   "review",
			Level:     "warn",
			Message:   "authentication failed: invalid JWT token",
			Metadata:  []byte(`{"correlation_id": "req-456", "user_id": 123}`),
			CreatedAt: time.Now(),
		},
	}

	mockProvider.On("Generate", mock.Anything, mock.Anything).Return(&ai.Response{
		Content:      `{"root_cause":"JWT token validation failed - token may be expired or malformed","suggested_fix":"1. Check token expiration time\n2. Verify JWT secret configuration\n3. Ensure token format is correct","severity":3,"related_logs":["req-456"],"fix_steps":["Check JWT token expiration","Verify JWT_SECRET environment variable","Review token generation logic"]}`,
		Model:        "qwen2.5-coder:7b",
		FinishReason: "complete",
		ResponseTime: 1500 * time.Millisecond,
	}, nil)

	req := AnalysisRequest{
		LogEntries: logEntries,
		Context:    "warn",
	}

	result, err := analyzer.Analyze(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.RootCause, "JWT")
	assert.Equal(t, 3, result.Severity)
	assert.Len(t, result.FixSteps, 3)

	mockProvider.AssertExpectations(t)
}

// TestAnalyze_CacheHit tests that cached analyses are returned
func TestAnalyze_CacheHit(t *testing.T) {
	mockProvider := new(MockAIProvider)
	analyzer := NewAIAnalyzer(mockProvider)

	logEntries := []logs_models.LogEntry{
		{
			ID:        3,
			Service:   "logs",
			Level:     "error",
			Message:   "connection timeout after 30s",
			Metadata:  []byte(`{}`),
			CreatedAt: time.Now(),
		},
	}

	// First call - should hit AI provider
	mockProvider.On("Generate", mock.Anything, mock.Anything).Return(&ai.Response{
		Content:      `{"root_cause":"Network timeout connecting to external service","suggested_fix":"Increase timeout value or check network connectivity","severity":4,"related_logs":[],"fix_steps":["Check network latency","Increase timeout configuration","Verify service endpoint is reachable"]}`,
		Model:        "qwen2.5-coder:7b",
		FinishReason: "complete",
		ResponseTime: 2 * time.Second,
	}, nil).Once()

	req := AnalysisRequest{
		LogEntries: logEntries,
		Context:    "error",
	}

	// First analysis
	result1, err := analyzer.Analyze(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result1)

	// Second analysis - should use cache, no additional AI call
	result2, err := analyzer.Analyze(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result2)
	assert.Equal(t, result1.RootCause, result2.RootCause)

	// Verify mock was called only once
	mockProvider.AssertNumberOfCalls(t, "Generate", 1)
}

// TestAnalyze_MultipleLogEntries tests analysis of multiple related logs
func TestAnalyze_MultipleLogEntries(t *testing.T) {
	mockProvider := new(MockAIProvider)
	analyzer := NewAIAnalyzer(mockProvider)

	logEntries := []logs_models.LogEntry{
		{
			ID:        4,
			Service:   "review",
			Level:     "error",
			Message:   "failed to connect to Ollama at http://localhost:11434",
			Metadata:  []byte(`{"correlation_id": "req-789"}`),
			CreatedAt: time.Now().Add(-2 * time.Second),
		},
		{
			ID:        5,
			Service:   "review",
			Level:     "error",
			Message:   "analysis timeout after 30 seconds",
			Metadata:  []byte(`{"correlation_id": "req-789"}`),
			CreatedAt: time.Now(),
		},
	}

	mockProvider.On("Generate", mock.Anything, mock.Anything).Return(&ai.Response{
		Content:      `{"root_cause":"Ollama service is not running or unreachable, causing analysis timeouts","suggested_fix":"Start Ollama service: systemctl start ollama or docker-compose up ollama","severity":5,"related_logs":["req-789"],"fix_steps":["Check if Ollama service is running","Verify OLLAMA_ENDPOINT configuration","Restart Ollama service if needed"]}`,
		Model:        "qwen2.5-coder:7b",
		FinishReason: "complete",
		ResponseTime: 2 * time.Second,
	}, nil)

	req := AnalysisRequest{
		LogEntries: logEntries,
		Context:    "error",
	}

	result, err := analyzer.Analyze(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.RootCause, "Ollama")
	assert.Equal(t, 5, result.Severity)
	assert.Contains(t, result.RelatedLogs, "req-789")

	mockProvider.AssertExpectations(t)
}

// TestAnalyze_InvalidJSON tests handling of invalid JSON responses
func TestAnalyze_InvalidJSON(t *testing.T) {
	mockProvider := new(MockAIProvider)
	analyzer := NewAIAnalyzer(mockProvider)

	logEntries := []logs_models.LogEntry{
		{
			ID:        6,
			Service:   "portal",
			Level:     "error",
			Message:   "some error",
			Metadata:  []byte(`{}`),
			CreatedAt: time.Now(),
		},
	}

	// AI returns invalid JSON
	mockProvider.On("Generate", mock.Anything, mock.Anything).Return(&ai.Response{
		Content:      `This is not valid JSON`,
		Model:        "qwen2.5-coder:7b",
		FinishReason: "complete",
		ResponseTime: 1 * time.Second,
	}, nil)

	req := AnalysisRequest{
		LogEntries: logEntries,
		Context:    "error",
	}

	result, err := analyzer.Analyze(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "parse")

	mockProvider.AssertExpectations(t)
}

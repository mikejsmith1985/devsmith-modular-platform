package review_services

import (
    "context"
    "testing"

    "github.com/mikejsmith1985/devsmith-modular-platform/internal/testutils"
    "github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// simple mock logger that satisfies logger.Interface with no-op methods
type nopLogger struct{}
func (n *nopLogger) Info(msg string, keyvals ...interface{})  {}
func (n *nopLogger) Debug(msg string, keyvals ...interface{}) {}
func (n *nopLogger) Warn(msg string, keyvals ...interface{})  {}
func (n *nopLogger) Error(msg string, keyvals ...interface{}) {}
func (n *nopLogger) Fatal(msg string, keyvals ...interface{}) {}
func (n *nopLogger) Panic(msg string, keyvals ...interface{}) {}
func (n *nopLogger) WithContext(ctx context.Context) logger.Interface   { return n }
func (n *nopLogger) WithFields(keyvals ...interface{}) logger.Interface  { return n }
func (n *nopLogger) Flush(ctx context.Context) error             { return nil }
func (n *nopLogger) Close() error                                { return nil }

// Mock Ollama that returns a fixed response
type mockOllama struct{
    resp string
    err error
}
func (m *mockOllama) Generate(ctx context.Context, prompt string) (string, error) {
    return m.resp, m.err
}

func TestAttemptJSONRepair_Success(t *testing.T) {
    repairedJSON := `{"summary":"ok","line_explanations":[]}`
    mock := &mockOllama{resp: repairedJSON, err: nil}
    svc := NewDetailedService(mock, &testutils.MockAnalysisRepository{}, &nopLogger{})

    got, err := svc.attemptJSONRepair(context.Background(), "garbage output")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got != repairedJSON {
        t.Fatalf("expected repaired JSON %s, got %s", repairedJSON, got)
    }
}

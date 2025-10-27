package testutils

import (
	"context"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// MockLogger is a no-op logger for use in tests.
type MockLogger struct{}

// Info is a no-op implementation of logger.Interface.Info.
func (m *MockLogger) Info(msg string, keyvals ...interface{})            {}
// Debug is a no-op implementation of logger.Interface.Debug.
func (m *MockLogger) Debug(msg string, keyvals ...interface{})           {}
// Warn is a no-op implementation of logger.Interface.Warn.
func (m *MockLogger) Warn(msg string, keyvals ...interface{})            {}
// Error is a no-op implementation of logger.Interface.Error.
func (m *MockLogger) Error(msg string, keyvals ...interface{})           {}
// Fatal is a no-op implementation of logger.Interface.Fatal.
func (m *MockLogger) Fatal(msg string, keyvals ...interface{})           {}
// Panic is a no-op implementation of logger.Interface.Panic.
func (m *MockLogger) Panic(msg string, keyvals ...interface{})           {}
// WithContext returns the MockLogger itself (no-op).
func (m *MockLogger) WithContext(ctx context.Context) logger.Interface   { return m }
// WithFields returns the MockLogger itself (no-op).
func (m *MockLogger) WithFields(keyvals ...interface{}) logger.Interface { return m }
// Flush is a no-op implementation of logger.Interface.Flush.
func (m *MockLogger) Flush(ctx context.Context) error                    { return nil }
// Close is a no-op implementation of logger.Interface.Close.
func (m *MockLogger) Close() error                                       { return nil }

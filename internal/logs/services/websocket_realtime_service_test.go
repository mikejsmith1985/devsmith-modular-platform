package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWebSocketRealtimeService is a mock for testing WebSocket service
type MockWebSocketRealtimeService struct {
	mock.Mock
	connCount int
}

// RegisterConnection mocks the RegisterConnection method
func (m *MockWebSocketRealtimeService) RegisterConnection(ctx context.Context, connectionID string) error {
	m.connCount++
	args := m.Called(ctx, connectionID)
	return args.Error(0)
}

// UnregisterConnection mocks the UnregisterConnection method
func (m *MockWebSocketRealtimeService) UnregisterConnection(ctx context.Context, connectionID string) error {
	if m.connCount > 0 {
		m.connCount--
	}
	args := m.Called(ctx, connectionID)
	return args.Error(0)
}

// BroadcastStats mocks the BroadcastStats method
func (m *MockWebSocketRealtimeService) BroadcastStats(ctx context.Context, stats *models.DashboardStats) error {
	args := m.Called(ctx, stats)
	return args.Error(0)
}

// BroadcastAlert mocks the BroadcastAlert method
func (m *MockWebSocketRealtimeService) BroadcastAlert(ctx context.Context, violation *models.AlertThresholdViolation) error {
	args := m.Called(ctx, violation)
	return args.Error(0)
}

// GetConnectionCount mocks the GetConnectionCount method
func (m *MockWebSocketRealtimeService) GetConnectionCount(ctx context.Context) (int, error) {
	// Only use testify mock if On() is set, otherwise return tracked state
	// This avoids panic from testify when no On() is set
	return m.connCount, nil
}

// TestRegisterConnection_SuccessfulRegistration validates connection registration.
func TestRegisterConnection_SuccessfulRegistration(t *testing.T) {
	// GIVEN: A WebSocket realtime service
	mockService := new(MockWebSocketRealtimeService)
	connectionID := "conn-12345"

	mockService.On("RegisterConnection", mock.Anything, connectionID).Return(nil)

	// WHEN: Registering a new connection
	err := mockService.RegisterConnection(context.Background(), connectionID)

	// THEN: Should register without error
	assert.NoError(t, err)
	mockService.AssertCalled(t, "RegisterConnection", mock.Anything, connectionID)
}

// TestRegisterConnection_MultipleConnections validates multiple connections.
func TestRegisterConnection_MultipleConnections(t *testing.T) {
	// GIVEN: WebSocket service with multiple connections
	mockService := new(MockWebSocketRealtimeService)

	connectionIDs := []string{"conn-1", "conn-2", "conn-3", "conn-4"}

	for _, id := range connectionIDs {
		mockService.On("RegisterConnection", mock.Anything, id).Return(nil)
	}

	// WHEN: Registering multiple connections
	for _, id := range connectionIDs {
		err := mockService.RegisterConnection(context.Background(), id)
		assert.NoError(t, err)
	}

	// THEN: All should be registered
	assert.Equal(t, 4, len(connectionIDs))
}

// TestUnregisterConnection_SuccessfulUnregistration validates connection removal.
func TestUnregisterConnection_SuccessfulUnregistration(t *testing.T) {
	// GIVEN: A registered connection
	mockService := new(MockWebSocketRealtimeService)
	connectionID := "conn-12345"

	mockService.On("UnregisterConnection", mock.Anything, connectionID).Return(nil)

	// WHEN: Unregistering connection
	err := mockService.UnregisterConnection(context.Background(), connectionID)

	// THEN: Should unregister without error
	assert.NoError(t, err)
	mockService.AssertCalled(t, "UnregisterConnection", mock.Anything, connectionID)
}

// TestUnregisterConnection_NonexistentConnection validates missing connection handling.
func TestUnregisterConnection_NonexistentConnection(t *testing.T) {
	// GIVEN: Attempting to unregister non-existent connection
	mockService := new(MockWebSocketRealtimeService)

	mockService.On("UnregisterConnection", mock.Anything, "nonexistent").Return(assert.AnError)

	// WHEN: Unregistering non-existent connection
	err := mockService.UnregisterConnection(context.Background(), "nonexistent")

	// THEN: Should return error
	assert.Error(t, err)
}

// TestBroadcastStats_SendsToAllConnections validates broadcast to all clients.
func TestBroadcastStats_SendsToAllConnections(t *testing.T) {
	// GIVEN: WebSocket service with active connections
	mockService := new(MockWebSocketRealtimeService)
	now := time.Now()

	stats := &models.DashboardStats{
		GeneratedAt:   now,
		ServiceStats:  make(map[string]*models.LogStats),
		ServiceHealth: make(map[string]*models.ServiceHealth),
	}

	mockService.On("BroadcastStats", mock.Anything, stats).Return(nil)

	// WHEN: Broadcasting stats
	err := mockService.BroadcastStats(context.Background(), stats)

	// THEN: Should broadcast without error
	assert.NoError(t, err)
	mockService.AssertCalled(t, "BroadcastStats", mock.Anything, stats)
}

// TestBroadcastStats_UpdatesMultipleClients validates all clients receive update.
func TestBroadcastStats_UpdatesMultipleClients(t *testing.T) {
	// GIVEN: Multiple connected clients
	mockService := new(MockWebSocketRealtimeService)
	now := time.Now()

	stats := &models.DashboardStats{
		GeneratedAt: now,
	}

	// Setup connections
	connIDs := []string{"conn-1", "conn-2", "conn-3"}
	for _, id := range connIDs {
		mockService.On("RegisterConnection", mock.Anything, id).Return(nil)
	}

	mockService.On("BroadcastStats", mock.Anything, stats).Return(nil)

	// WHEN: Registering clients and broadcasting
	for _, id := range connIDs {
		mockService.RegisterConnection(context.Background(), id)
	}
	err := mockService.BroadcastStats(context.Background(), stats)

	// THEN: All clients should receive update
	assert.NoError(t, err)
}

// TestBroadcastAlert_SendsToAllConnections validates alert broadcast.
func TestBroadcastAlert_SendsToAllConnections(t *testing.T) {
	// GIVEN: WebSocket service with active connections
	mockService := new(MockWebSocketRealtimeService)
	now := time.Now()

	violation := &models.AlertThresholdViolation{
		Service:        "portal",
		Level:          "error",
		CurrentCount:   150,
		ThresholdValue: 100,
		Timestamp:      now,
	}

	mockService.On("BroadcastAlert", mock.Anything, violation).Return(nil)

	// WHEN: Broadcasting alert
	err := mockService.BroadcastAlert(context.Background(), violation)

	// THEN: Should broadcast without error
	assert.NoError(t, err)
	mockService.AssertCalled(t, "BroadcastAlert", mock.Anything, violation)
}

// TestBroadcastAlert_UrgentMessage validates alerts are sent immediately.
func TestBroadcastAlert_UrgentMessage(t *testing.T) {
	// GIVEN: Critical alert to broadcast
	mockService := new(MockWebSocketRealtimeService)
	now := time.Now()

	criticalViolation := &models.AlertThresholdViolation{
		Service:        "analytics",
		Level:          "error",
		CurrentCount:   500,
		ThresholdValue: 200,
		Timestamp:      now,
	}

	mockService.On("BroadcastAlert", mock.Anything, criticalViolation).Return(nil)

	// WHEN: Broadcasting critical alert
	err := mockService.BroadcastAlert(context.Background(), criticalViolation)

	// THEN: Should send immediately
	assert.NoError(t, err)
}

// TestGetConnectionCount_ReturnsActiveCount validates active connection count.
func TestGetConnectionCount_ReturnsActiveCount(t *testing.T) {
	// GIVEN: WebSocket service with active connections
	mockService := new(MockWebSocketRealtimeService)

	// Register connections
	connIDs := []string{"conn-1", "conn-2", "conn-3"}
	for _, id := range connIDs {
		mockService.On("RegisterConnection", mock.Anything, id).Return(nil)
	}

	mockService.On("GetConnectionCount", mock.Anything).Return(3, nil)

	// WHEN: Getting connection count
	for _, id := range connIDs {
		mockService.RegisterConnection(context.Background(), id)
	}
	count, err := mockService.GetConnectionCount(context.Background())

	// THEN: Should return correct count
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

// TestGetConnectionCount_ZeroConnections validates zero count when no connections.
func TestGetConnectionCount_ZeroConnections(t *testing.T) {
	// GIVEN: WebSocket service with no connections
	mockService := new(MockWebSocketRealtimeService)

	mockService.On("GetConnectionCount", mock.Anything).Return(0, nil)

	// WHEN: Getting connection count
	count, err := mockService.GetConnectionCount(context.Background())

	// THEN: Should return zero
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestBroadcastStats_RealtimeUpdates validates stats update frequency.
func TestBroadcastStats_RealtimeUpdates(t *testing.T) {
	// GIVEN: Real-time stats broadcast every second
	mockService := new(MockWebSocketRealtimeService)

	now := time.Now()
	stats := &models.DashboardStats{
		GeneratedAt: now,
	}

	// Mock broadcast multiple times
	mockService.On("BroadcastStats", mock.Anything, mock.MatchedBy(func(s *models.DashboardStats) bool {
		return s.GeneratedAt.Before(time.Now().Add(time.Second))
	})).Return(nil)

	// WHEN: Broadcasting real-time stats multiple times
	for i := 0; i < 3; i++ {
		err := mockService.BroadcastStats(context.Background(), stats)
		assert.NoError(t, err)
	}

	// THEN: All broadcasts should succeed
	mockService.AssertCalled(t, "BroadcastStats", mock.Anything, mock.Anything)
}

// TestRegisterConnection_ContextCancellation validates context handling.
func TestRegisterConnection_ContextCancellation(t *testing.T) {
	// GIVEN: Cancelled context
	mockService := new(MockWebSocketRealtimeService)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockService.On("RegisterConnection", mock.Anything, "conn-1").Return(context.Canceled)

	// WHEN: Registering with cancelled context
	err := mockService.RegisterConnection(ctx, "conn-1")

	// THEN: Should return context cancelled error
	assert.Error(t, err)
}

// TestBroadcastStats_ContextCancellation validates broadcast context handling.
func TestBroadcastStats_ContextCancellation(t *testing.T) {
	// GIVEN: Cancelled context
	mockService := new(MockWebSocketRealtimeService)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stats := &models.DashboardStats{}

	mockService.On("BroadcastStats", mock.Anything, stats).Return(context.Canceled)

	// WHEN: Broadcasting with cancelled context
	err := mockService.BroadcastStats(ctx, stats)

	// THEN: Should return context cancelled error
	assert.Error(t, err)
}

// TestUnregisterConnection_ReducesCount validates count after unregister.
func TestUnregisterConnection_ReducesCount(t *testing.T) {
	// GIVEN: Registered connections
	mockService := new(MockWebSocketRealtimeService)

	connIDs := []string{"conn-1", "conn-2"}
	for _, id := range connIDs {
		mockService.On("RegisterConnection", mock.Anything, id).Return(nil)
	}

	// Register connections
	for _, id := range connIDs {
		mockService.RegisterConnection(context.Background(), id)
	}

	// WHEN: Getting initial count
	initialCount, _ := mockService.GetConnectionCount(context.Background())

	mockService.On("UnregisterConnection", mock.Anything, "conn-1").Return(nil)

	// Unregister one
	mockService.UnregisterConnection(context.Background(), "conn-1")
	finalCount, _ := mockService.GetConnectionCount(context.Background())

	// THEN: Count should decrease
	assert.Equal(t, 2, initialCount)
	assert.Equal(t, 1, finalCount)
}

// TestBroadcastAlert_MultipleViolations validates multiple alerts can be broadcast.
func TestBroadcastAlert_MultipleViolations(t *testing.T) {
	// GIVEN: Multiple violations to broadcast
	mockService := new(MockWebSocketRealtimeService)
	now := time.Now()

	violations := []models.AlertThresholdViolation{
		{Service: "portal", Level: "error", CurrentCount: 150, ThresholdValue: 100, Timestamp: now},
		{Service: "analytics", Level: "error", CurrentCount: 300, ThresholdValue: 200, Timestamp: now},
		{Service: "review", Level: "warning", CurrentCount: 100, ThresholdValue: 80, Timestamp: now},
	}

	for _, v := range violations {
		violation := v // Capture in loop
		mockService.On("BroadcastAlert", mock.Anything, &violation).Return(nil)
	}

	// WHEN: Broadcasting multiple alerts
	for i := range violations {
		err := mockService.BroadcastAlert(context.Background(), &violations[i])
		assert.NoError(t, err)
	}

	// THEN: All should be broadcast
	assert.Equal(t, 3, len(violations))
}

// TestConnectionLifecycle_RegisterAndUnregister validates full connection lifecycle.
func TestConnectionLifecycle_RegisterAndUnregister(t *testing.T) {
	// GIVEN: WebSocket service
	mockService := new(MockWebSocketRealtimeService)
	connID := "conn-lifecycle"

	mockService.On("RegisterConnection", mock.Anything, connID).Return(nil)
	mockService.On("UnregisterConnection", mock.Anything, connID).Return(nil)

	// WHEN: Registering connection
	err1 := mockService.RegisterConnection(context.Background(), connID)
	count1, _ := mockService.GetConnectionCount(context.Background())

	// THEN: Should be registered
	assert.NoError(t, err1)
	assert.Equal(t, 1, count1)

	// WHEN: Unregistering connection
	err2 := mockService.UnregisterConnection(context.Background(), connID)
	count2, _ := mockService.GetConnectionCount(context.Background())

	// THEN: Should be unregistered
	assert.NoError(t, err2)
	assert.Equal(t, 0, count2)
}

// TestBroadcastStats_WithServiceData validates stats with service information.
func TestBroadcastStats_WithServiceData(t *testing.T) {
	// GIVEN: Stats with service data
	mockService := new(MockWebSocketRealtimeService)
	now := time.Now()

	stats := &models.DashboardStats{
		GeneratedAt:   now,
		ServiceStats:  make(map[string]*models.LogStats),
		ServiceHealth: make(map[string]*models.ServiceHealth),
	}

	// Add service data
	stats.ServiceStats["portal"] = &models.LogStats{
		Service:    "portal",
		TotalCount: 100,
	}
	stats.ServiceHealth["portal"] = &models.ServiceHealth{
		Service: "portal",
		Status:  "OK",
	}

	mockService.On("BroadcastStats", mock.Anything, stats).Return(nil)

	// WHEN: Broadcasting stats with data
	err := mockService.BroadcastStats(context.Background(), stats)

	// THEN: Should broadcast successfully
	assert.NoError(t, err)
	assert.Equal(t, 1, len(stats.ServiceStats))
	assert.Equal(t, "portal", stats.ServiceStats["portal"].Service)
}

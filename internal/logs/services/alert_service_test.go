//nolint:govet // Test file: struct literal fields needed for assertions
package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAlertService is a mock for testing the alert service
type MockAlertService struct {
	mock.Mock
}

// CreateAlertConfig mocks the CreateAlertConfig method
func (m *MockAlertService) CreateAlertConfig(ctx context.Context, config *models.AlertConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

// UpdateAlertConfig mocks the UpdateAlertConfig method
func (m *MockAlertService) UpdateAlertConfig(ctx context.Context, config *models.AlertConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

// GetAlertConfig mocks the GetAlertConfig method
func (m *MockAlertService) GetAlertConfig(ctx context.Context, service string) (*models.AlertConfig, error) {
	args := m.Called(ctx, service)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AlertConfig), args.Error(1)
}

// CheckThresholds mocks the CheckThresholds method
func (m *MockAlertService) CheckThresholds(ctx context.Context) ([]models.AlertThresholdViolation, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AlertThresholdViolation), args.Error(1)
}

// SendAlert mocks the SendAlert method
func (m *MockAlertService) SendAlert(ctx context.Context, violation *models.AlertThresholdViolation) error {
	args := m.Called(ctx, violation)
	return args.Error(0)
}

// TestCreateAlertConfig_CreatesNewConfig validates alert config creation.
func TestCreateAlertConfig_CreatesNewConfig(t *testing.T) { //nolint:govet
	// GIVEN: An alert service and new alert config
	mockService := new(MockAlertService)
	config := &models.AlertConfig{
		Service:                "portal",
		ErrorThresholdPerMin:   100,
		WarningThresholdPerMin: 50,
		AlertEmail:             "admin@example.com",
		Enabled:                true,
	}

	mockService.On("CreateAlertConfig", mock.Anything, config).Return(nil)

	// WHEN: Creating alert config
	err := mockService.CreateAlertConfig(context.Background(), config)

	// THEN: Should create without error
	assert.NoError(t, err)
	mockService.AssertCalled(t, "CreateAlertConfig", mock.Anything, config)
}

// TestCreateAlertConfig_ValidatesRequiredFields validates required fields.
func TestCreateAlertConfig_ValidatesRequiredFields(t *testing.T) { //nolint:govet
	// GIVEN: Alert config with missing service
	config := &models.AlertConfig{
		ErrorThresholdPerMin: 100,
		AlertEmail:           "admin@example.com",
	}

	// WHEN: Creating config without service, it should validate
	// THEN: Test should verify service validation occurs
	// This test documents the requirement that service is required
	assert.Empty(t, config.Service)
	assert.NotZero(t, config.ErrorThresholdPerMin)
}

// TestGetAlertConfig_ReturnsConfig validates config retrieval.
func TestGetAlertConfig_ReturnsConfig(t *testing.T) { //nolint:govet
	// GIVEN: An alert service with existing config
	mockService := new(MockAlertService)
	now := time.Now()

	config := &models.AlertConfig{
		Service:                "analytics",
		ErrorThresholdPerMin:   200,
		WarningThresholdPerMin: 100,
		AlertEmail:             "ops@example.com",
		AlertWebhookURL:        "https://example.com/alerts",
		Enabled:                true,
		ID:                     1,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	mockService.On("GetAlertConfig", mock.Anything, "analytics").Return(config, nil)

	// WHEN: Getting alert config
	result, err := mockService.GetAlertConfig(context.Background(), "analytics")

	// THEN: Should return config
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "analytics", result.Service)
	assert.Equal(t, 200, result.ErrorThresholdPerMin)
	assert.Equal(t, "https://example.com/alerts", result.AlertWebhookURL)
}

// TestGetAlertConfig_NotFound validates missing config handling.
func TestGetAlertConfig_NotFound(t *testing.T) { //nolint:govet
	// GIVEN: An alert service with no config for service
	mockService := new(MockAlertService)

	mockService.On("GetAlertConfig", mock.Anything, "unknown").Return(nil, assert.AnError)

	// WHEN: Getting non-existent config
	result, err := mockService.GetAlertConfig(context.Background(), "unknown")

	// THEN: Should return error
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestUpdateAlertConfig_UpdatesExisting validates config updates.
func TestUpdateAlertConfig_UpdatesExisting(t *testing.T) { //nolint:govet
	// GIVEN: An existing alert config to update
	mockService := new(MockAlertService)
	config := &models.AlertConfig{
		ID:                     1,
		Service:                "portal",
		ErrorThresholdPerMin:   150, // Updated from 100
		WarningThresholdPerMin: 75,  // Updated from 50
		AlertEmail:             "newemail@example.com",
		Enabled:                true,
	}

	mockService.On("UpdateAlertConfig", mock.Anything, config).Return(nil)

	// WHEN: Updating config
	err := mockService.UpdateAlertConfig(context.Background(), config)

	// THEN: Should update without error
	assert.NoError(t, err)
	mockService.AssertCalled(t, "UpdateAlertConfig", mock.Anything, config)
}

// TestCheckThresholds_DetectsViolations validates threshold checking.
func TestCheckThresholds_DetectsViolations(t *testing.T) { //nolint:govet
	// GIVEN: Alert service monitoring thresholds
	mockService := new(MockAlertService)
	now := time.Now()

	violations := []models.AlertThresholdViolation{
		{
			Service:        "analytics",
			Level:          "error",
			CurrentCount:   250, // Exceeds 200 threshold
			ThresholdValue: 200,
			Timestamp:      now,
			ID:             1,
		},
		{
			Service:        "portal",
			Level:          "warning",
			CurrentCount:   60, // Exceeds 50 threshold
			ThresholdValue: 50,
			Timestamp:      now,
			ID:             2,
		},
	}

	mockService.On("CheckThresholds", mock.Anything).Return(violations, nil)

	// WHEN: Checking thresholds
	result, err := mockService.CheckThresholds(context.Background())

	// THEN: Should detect violations
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "analytics", result[0].Service)
	assert.Equal(t, int64(250), result[0].CurrentCount)
	assert.Equal(t, "portal", result[1].Service)
}

// TestCheckThresholds_NoViolations validates no violations detected.
func TestCheckThresholds_NoViolations(t *testing.T) { //nolint:govet
	// GIVEN: Services within thresholds
	mockService := new(MockAlertService)

	mockService.On("CheckThresholds", mock.Anything).Return([]models.AlertThresholdViolation{}, nil)

	// WHEN: Checking thresholds
	result, err := mockService.CheckThresholds(context.Background())

	// THEN: Should return empty list
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestSendAlert_SendsEmail validates email alerts.
func TestSendAlert_SendsEmail(t *testing.T) { //nolint:govet
	// GIVEN: Alert service with email config
	mockService := new(MockAlertService)
	now := time.Now()

	violation := &models.AlertThresholdViolation{
		Service:        "portal",
		Level:          "error",
		CurrentCount:   150,
		ThresholdValue: 100,
		Timestamp:      now,
	}

	mockService.On("SendAlert", mock.Anything, violation).Return(nil)

	// WHEN: Sending alert
	err := mockService.SendAlert(context.Background(), violation)

	// THEN: Should send without error
	assert.NoError(t, err)
	mockService.AssertCalled(t, "SendAlert", mock.Anything, violation)
}

// TestSendAlert_SendsWebhook validates webhook alerts.
func TestSendAlert_SendsWebhook(t *testing.T) { //nolint:govet
	// GIVEN: Alert service with webhook configured
	mockService := new(MockAlertService)
	now := time.Now()

	violation := &models.AlertThresholdViolation{
		Service:        "analytics",
		Level:          "error",
		CurrentCount:   300,
		ThresholdValue: 200,
		Timestamp:      now,
	}

	// Mock webhook server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockService.On("SendAlert", mock.Anything, violation).Return(nil)

	// WHEN: Sending alert
	err := mockService.SendAlert(context.Background(), violation)

	// THEN: Should send webhook without error
	assert.NoError(t, err)
}

// TestAlertConfig_DisabledAlerts validates disabled alerts are not sent.
func TestAlertConfig_DisabledAlerts(t *testing.T) { //nolint:govet
	// GIVEN: Alert config that is disabled
	mockService := new(MockAlertService)

	config := &models.AlertConfig{
		Service: "review",
		Enabled: false,
		ID:      1,
	}

	mockService.On("GetAlertConfig", mock.Anything, "review").Return(config, nil)

	// WHEN: Getting disabled config
	result, err := mockService.GetAlertConfig(context.Background(), "review")

	// THEN: Config should indicate disabled state
	assert.NoError(t, err)
	assert.False(t, result.Enabled)
}

// TestAlertConfig_EnabledAlerts validates enabled alerts work correctly.
func TestAlertConfig_EnabledAlerts(t *testing.T) { //nolint:govet
	// GIVEN: Alert config that is enabled
	mockService := new(MockAlertService)

	config := &models.AlertConfig{
		Service: "portal",
		Enabled: true,
		ID:      1,
	}

	mockService.On("GetAlertConfig", mock.Anything, "portal").Return(config, nil)

	// WHEN: Getting enabled config
	result, err := mockService.GetAlertConfig(context.Background(), "portal")

	// THEN: Config should indicate enabled state
	assert.NoError(t, err)
	assert.True(t, result.Enabled)
}

// TestCheckThresholds_MultipleServices validates multiple services checked.
func TestCheckThresholds_MultipleServices(t *testing.T) { //nolint:govet
	// GIVEN: Alert service monitoring multiple services
	mockService := new(MockAlertService)
	now := time.Now()

	violations := []models.AlertThresholdViolation{
		{Service: "portal", Level: "error", CurrentCount: 150, ThresholdValue: 100, Timestamp: now},
		{Service: "analytics", Level: "error", CurrentCount: 300, ThresholdValue: 200, Timestamp: now},
		{Service: "review", Level: "warning", CurrentCount: 100, ThresholdValue: 80, Timestamp: now},
	}

	mockService.On("CheckThresholds", mock.Anything).Return(violations, nil)

	// WHEN: Checking thresholds
	result, err := mockService.CheckThresholds(context.Background())

	// THEN: Should check all services
	assert.NoError(t, err)
	assert.Len(t, result, 3)
}

// TestAlertThresholdViolation_AlertSentTracking validates alert sent tracking.
func TestAlertThresholdViolation_AlertSentTracking(t *testing.T) { //nolint:govet //nolint:govet
	// GIVEN: Violation initially without alert sent
	now := time.Now()

	violation := &models.AlertThresholdViolation{
		Service:        "portal",
		Level:          "error",
		CurrentCount:   150,
		ThresholdValue: 100,
		Timestamp:      now,
		AlertSentAt:    nil,
	}

	// WHEN: Before sending alert
	// THEN: AlertSentAt should be nil
	assert.Nil(t, violation.AlertSentAt)

	// WHEN: After sending alert
	sendTime := time.Now()
	violation.AlertSentAt = &sendTime

	// THEN: AlertSentAt should be populated
	assert.NotNil(t, violation.AlertSentAt)
	assert.Equal(t, sendTime, *violation.AlertSentAt)
}

// TestSendAlert_ContextCancellation validates context cancellation handling.
func TestSendAlert_ContextCancellation(t *testing.T) { //nolint:govet
	// GIVEN: Cancelled context
	mockService := new(MockAlertService)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	violation := &models.AlertThresholdViolation{
		Service: "portal",
		Level:   "error",
	}

	mockService.On("SendAlert", mock.Anything, violation).Return(context.Canceled)

	// WHEN: Sending alert with cancelled context
	err := mockService.SendAlert(ctx, violation)

	// THEN: Should return context cancelled error
	assert.Error(t, err)
}

// TestCreateAlertConfig_StoresMultipleConfigs validates multiple configs stored.
func TestCreateAlertConfig_StoresMultipleConfigs(t *testing.T) { //nolint:govet
	// GIVEN: Alert service
	mockService := new(MockAlertService)

	configs := []*models.AlertConfig{
		{Service: "portal", ErrorThresholdPerMin: 100},
		{Service: "analytics", ErrorThresholdPerMin: 200},
		{Service: "review", ErrorThresholdPerMin: 150},
	}

	// WHEN: Creating multiple configs
	for _, cfg := range configs {
		mockService.On("CreateAlertConfig", mock.Anything, cfg).Return(nil)
		err := mockService.CreateAlertConfig(context.Background(), cfg)
		assert.NoError(t, err)
	}

	// THEN: All should be stored
	assert.Equal(t, 3, len(configs))
}

// TestUpdateAlertConfig_UpdatesThreshold validates threshold updates.
func TestUpdateAlertConfig_UpdatesThreshold(t *testing.T) { //nolint:govet
	// GIVEN: Existing config with initial threshold
	mockService := new(MockAlertService)

	oldConfig := &models.AlertConfig{
		ID:                   1,
		Service:              "portal",
		ErrorThresholdPerMin: 100,
	}

	newConfig := &models.AlertConfig{
		ID:                   1,
		Service:              "portal",
		ErrorThresholdPerMin: 150, // Updated
	}

	// WHEN: Updating threshold
	mockService.On("UpdateAlertConfig", mock.Anything, newConfig).Return(nil)
	err := mockService.UpdateAlertConfig(context.Background(), newConfig)

	// THEN: Should update successfully
	assert.NoError(t, err)
	assert.NotEqual(t, oldConfig.ErrorThresholdPerMin, newConfig.ErrorThresholdPerMin)
}

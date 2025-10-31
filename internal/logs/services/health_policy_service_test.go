package logs_services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPolicies_AllServicesConfigured(t *testing.T) {
	expectedServices := []string{"portal", "review", "logs", "analytics"}

	for _, svc := range expectedServices {
		_, ok := DefaultPolicies[svc]
		assert.True(t, ok, "Default policy missing for service: %s", svc)
	}
}

func TestDefaultPolicies_ValidConfiguration(t *testing.T) {
	tests := []struct {
		service         string
		expectedRepair  string
		expectedMax     int
		expectedEnabled bool
	}{
		{
			service:         "portal",
			expectedRepair:  "restart",
			expectedMax:     500,
			expectedEnabled: true,
		},
		{
			service:         "review",
			expectedRepair:  "restart",
			expectedMax:     1000,
			expectedEnabled: true,
		},
		{
			service:         "logs",
			expectedRepair:  "none",
			expectedMax:     500,
			expectedEnabled: false,
		},
		{
			service:         "analytics",
			expectedRepair:  "restart",
			expectedMax:     2000,
			expectedEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.service, func(t *testing.T) {
			policy, ok := DefaultPolicies[tt.service]
			assert.True(t, ok, "Policy not found for service: %s", tt.service)

			assert.Equal(t, tt.expectedMax, policy.MaxResponseTimeMs)
			assert.Equal(t, tt.expectedRepair, policy.RepairStrategy)
			assert.Equal(t, tt.expectedEnabled, policy.AutoRepairEnabled)
		})
	}
}

func TestGetPolicy_DefaultPolicy(t *testing.T) {
	service := NewHealthPolicyService(nil)
	policy, err := service.GetPolicy(context.Background(), "portal")

	assert.NoError(t, err)
	assert.NotNil(t, policy)
	assert.Equal(t, DefaultPolicies["portal"].ServiceName, policy.ServiceName)
}

func TestGetPolicy_UnknownService_Error(t *testing.T) {
	service := NewHealthPolicyService(nil)
	_, err := service.GetPolicy(context.Background(), "nonexistent")

	assert.Error(t, err)
}

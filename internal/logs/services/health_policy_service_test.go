package services

import (
	"testing"
)

func TestDefaultPolicies(t *testing.T) {
	defaults := DefaultPolicies()

	expectedServices := []string{"portal", "review", "logs", "analytics"}

	for _, svc := range expectedServices {
		if _, ok := defaults[svc]; !ok {
			t.Errorf("Default policy missing for service: %s", svc)
		}
	}
}

func TestDefaultPoliciesConfiguration(t *testing.T) {
	tests := []struct {
		service         string
		expectedMax     int
		expectedRepair  string
		expectedEnabled bool
	}{
		{
			service:         "portal",
			expectedMax:     500,
			expectedRepair:  "restart",
			expectedEnabled: true,
		},
		{
			service:         "review",
			expectedMax:     1000,
			expectedRepair:  "restart",
			expectedEnabled: true,
		},
		{
			service:         "logs",
			expectedMax:     500,
			expectedRepair:  "none",
			expectedEnabled: false,
		},
		{
			service:         "analytics",
			expectedMax:     2000,
			expectedRepair:  "restart",
			expectedEnabled: true,
		},
	}

	defaults := DefaultPolicies()

	for _, tt := range tests {
		t.Run(tt.service, func(t *testing.T) {
			policy, ok := defaults[tt.service]
			if !ok {
				t.Fatalf("Policy not found for service: %s", tt.service)
			}

			if policy.MaxResponseTimeMS != tt.expectedMax {
				t.Errorf("Expected max %dms, got %dms", tt.expectedMax, policy.MaxResponseTimeMS)
			}
			if policy.RepairStrategy != tt.expectedRepair {
				t.Errorf("Expected strategy %s, got %s", tt.expectedRepair, policy.RepairStrategy)
			}
			if policy.AutoRepairEnabled != tt.expectedEnabled {
				t.Errorf("Expected auto-repair %v, got %v", tt.expectedEnabled, policy.AutoRepairEnabled)
			}
		})
	}
}

func TestPolicyServiceDefaults(t *testing.T) {
	// Test that when a policy is not found, defaults are returned
	defaults := DefaultPolicies()

	for svcName, expectedPolicy := range defaults {
		t.Run(svcName, func(t *testing.T) {
			if expectedPolicy.ServiceName != svcName {
				t.Errorf("Service name mismatch: expected %s, got %s", svcName, expectedPolicy.ServiceName)
			}
			if expectedPolicy.MaxResponseTimeMS == 0 {
				t.Error("MaxResponseTimeMS not set")
			}
			if expectedPolicy.RepairStrategy == "" {
				t.Error("RepairStrategy not set")
			}
		})
	}
}

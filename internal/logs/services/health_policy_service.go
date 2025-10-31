package logs_services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

// Repair strategy constants
const (
	RepairStrategyRestart = "restart"
	RepairStrategyRebuild = "rebuild"
	RepairStrategyNone    = "none"
)

// HealthPolicy represents a health policy for a service
type HealthPolicy struct {
	UpdatedAt         time.Time `json:"updated_at"`
	ServiceName       string    `json:"service_name"`
	RepairStrategy    string    `json:"repair_strategy"` // restart, rebuild, none
	PolicyJSON        string    `json:"policy_json,omitempty"`
	ID                int       `json:"id"`
	MaxResponseTimeMs int       `json:"max_response_time_ms"`
	AutoRepairEnabled bool      `json:"auto_repair_enabled"`
	AlertOnWarn       bool      `json:"alert_on_warn"`
	AlertOnFail       bool      `json:"alert_on_fail"`
}

// HealthPolicyService manages health policies for services
type HealthPolicyService struct {
	db *sql.DB
}

// NewHealthPolicyService creates a new health policy service
func NewHealthPolicyService(db *sql.DB) *HealthPolicyService {
	return &HealthPolicyService{db: db}
}

// DefaultPolicies defines default policies for each service
var DefaultPolicies = map[string]HealthPolicy{
	"portal": {
		ServiceName:       "portal",
		MaxResponseTimeMs: 500,
		AutoRepairEnabled: true,
		RepairStrategy:    RepairStrategyRestart,
		AlertOnWarn:       false,
		AlertOnFail:       true,
	},
	"review": {
		ServiceName:       "review",
		MaxResponseTimeMs: 1000,
		AutoRepairEnabled: true,
		RepairStrategy:    RepairStrategyRestart,
		AlertOnWarn:       false,
		AlertOnFail:       true,
	},
	"logs": {
		ServiceName:       "logs",
		MaxResponseTimeMs: 500,
		AutoRepairEnabled: false,
		RepairStrategy:    RepairStrategyNone,
		AlertOnWarn:       false,
		AlertOnFail:       true,
	},
	"analytics": {
		ServiceName:       "analytics",
		MaxResponseTimeMs: 2000,
		AutoRepairEnabled: true,
		RepairStrategy:    RepairStrategyRestart,
		AlertOnWarn:       false,
		AlertOnFail:       true,
	},
}

// GetPolicy retrieves a policy for a service
func (s *HealthPolicyService) GetPolicy(ctx context.Context, serviceName string) (*HealthPolicy, error) {
	// If db is nil, return default policy
	if s.db == nil {
		if defaultPolicy, ok := DefaultPolicies[serviceName]; ok {
			return &defaultPolicy, nil
		}
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	var policy HealthPolicy
	err := s.db.QueryRowContext(ctx,
		`SELECT id, service_name, max_response_time_ms, auto_repair_enabled, repair_strategy, alert_on_warn, alert_on_fail, policy_json, updated_at
		 FROM logs.health_policies
		 WHERE service_name = $1`,
		serviceName,
	).Scan(&policy.ID, &policy.ServiceName, &policy.MaxResponseTimeMs, &policy.AutoRepairEnabled, &policy.RepairStrategy, &policy.AlertOnWarn, &policy.AlertOnFail, &policy.PolicyJSON, &policy.UpdatedAt)

	if err == sql.ErrNoRows {
		// Return default policy if not found
		if defaultPolicy, ok := DefaultPolicies[serviceName]; ok {
			return &defaultPolicy, nil
		}
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query policy: %w", err)
	}

	return &policy, nil
}

// GetAllPolicies retrieves all policies
func (s *HealthPolicyService) GetAllPolicies(ctx context.Context) ([]HealthPolicy, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, service_name, max_response_time_ms, auto_repair_enabled, repair_strategy, alert_on_warn, alert_on_fail, policy_json, updated_at
		 FROM logs.health_policies
		 ORDER BY service_name`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("warning: failed to close health policies rows: %v", err)
		}
	}()

	var policies []HealthPolicy
	for rows.Next() {
		var p HealthPolicy
		err := rows.Scan(&p.ID, &p.ServiceName, &p.MaxResponseTimeMs, &p.AutoRepairEnabled, &p.RepairStrategy, &p.AlertOnWarn, &p.AlertOnFail, &p.PolicyJSON, &p.UpdatedAt)
		if err != nil {
			continue
		}
		policies = append(policies, p)
	}

	return policies, rows.Err()
}

// UpdatePolicy updates a policy for a service
func (s *HealthPolicyService) UpdatePolicy(ctx context.Context, policy *HealthPolicy) error {
	// Check if policy exists
	var id int
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM logs.health_policies WHERE service_name = $1`,
		policy.ServiceName,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		// Insert new policy
		return s.createPolicy(ctx, policy)
	}

	if err != nil {
		return fmt.Errorf("failed to check existing policy: %w", err)
	}

	// Update existing policy
	_, err = s.db.ExecContext(ctx,
		`UPDATE logs.health_policies
		 SET max_response_time_ms = $1,
		     auto_repair_enabled = $2,
		     repair_strategy = $3,
		     alert_on_warn = $4,
		     alert_on_fail = $5,
		     policy_json = $6,
		     updated_at = NOW()
		 WHERE service_name = $7`,
		policy.MaxResponseTimeMs,
		policy.AutoRepairEnabled,
		policy.RepairStrategy,
		policy.AlertOnWarn,
		policy.AlertOnFail,
		policy.PolicyJSON,
		policy.ServiceName,
	)

	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	return nil
}

// createPolicy creates a new policy for a service
func (s *HealthPolicyService) createPolicy(ctx context.Context, policy *HealthPolicy) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO logs.health_policies
		 (service_name, max_response_time_ms, auto_repair_enabled, repair_strategy, alert_on_warn, alert_on_fail, policy_json, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`,
		policy.ServiceName,
		policy.MaxResponseTimeMs,
		policy.AutoRepairEnabled,
		policy.RepairStrategy,
		policy.AlertOnWarn,
		policy.AlertOnFail,
		policy.PolicyJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	return nil
}

// InitializeDefaultPolicies creates default policies for all known services
func (s *HealthPolicyService) InitializeDefaultPolicies(ctx context.Context) error {
	for _, policy := range DefaultPolicies {
		p := policy // Create copy
		err := s.UpdatePolicy(ctx, &p)
		if err != nil {
			fmt.Printf("failed to initialize policy for %s: %v\n", policy.ServiceName, err)
		}
	}
	return nil
}

// GetRepairStrategy returns the appropriate repair strategy for a service based on issue type
func (s *HealthPolicyService) GetRepairStrategy(ctx context.Context, serviceName, issueType string) (string, error) {
	policy, err := s.GetPolicy(ctx, serviceName)
	if err != nil {
		return RepairStrategyNone, err
	}

	if !policy.AutoRepairEnabled {
		return RepairStrategyNone, nil
	}

	// Override strategy for specific issue types
	switch issueType {
	case "timeout":
		return RepairStrategyRestart, nil
	case "crash":
		return RepairStrategyRebuild, nil
	case "security":
		return RepairStrategyRebuild, nil
	case "dependency":
		return RepairStrategyNone, nil // Can't repair dependency issues
	default:
		return policy.RepairStrategy, nil
	}
}

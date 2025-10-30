package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// HealthPolicy defines health check behavior for a service
type HealthPolicy struct {
	ID                  int       `json:"id"`
	ServiceName         string    `json:"service_name"`
	MaxResponseTimeMS   int       `json:"max_response_time_ms"`
	AutoRepairEnabled   bool      `json:"auto_repair_enabled"`
	RepairStrategy      string    `json:"repair_strategy"` // 'restart', 'rebuild', 'none'
	AlertOnWarn         bool      `json:"alert_on_warn"`
	AlertOnFail         bool      `json:"alert_on_fail"`
	PolicyJSON          map[string]interface{} `json:"policy_json,omitempty"`
	UpdatedAt           time.Time `json:"updated_at"`
	CreatedAt           time.Time `json:"created_at"`
}

// HealthPolicyService manages health check policies
type HealthPolicyService struct {
	db *sql.DB
}

// NewHealthPolicyService creates a new policy service
func NewHealthPolicyService(db *sql.DB) *HealthPolicyService {
	return &HealthPolicyService{db: db}
}

// DefaultPolicies returns sensible defaults for all services
func DefaultPolicies() map[string]HealthPolicy {
	return map[string]HealthPolicy{
		"portal": {
			ServiceName:       "portal",
			MaxResponseTimeMS: 500,
			AutoRepairEnabled: true,
			RepairStrategy:    "restart",
			AlertOnWarn:       false,
			AlertOnFail:       true,
		},
		"review": {
			ServiceName:       "review",
			MaxResponseTimeMS: 1000,
			AutoRepairEnabled: true,
			RepairStrategy:    "restart",
			AlertOnWarn:       false,
			AlertOnFail:       true,
		},
		"logs": {
			ServiceName:       "logs",
			MaxResponseTimeMS: 500,
			AutoRepairEnabled: false,
			RepairStrategy:    "none",
			AlertOnWarn:       false,
			AlertOnFail:       true,
		},
		"analytics": {
			ServiceName:       "analytics",
			MaxResponseTimeMS: 2000,
			AutoRepairEnabled: true,
			RepairStrategy:    "restart",
			AlertOnWarn:       false,
			AlertOnFail:       true,
		},
	}
}

// GetPolicy retrieves a policy for a service
func (s *HealthPolicyService) GetPolicy(ctx context.Context, serviceName string) (*HealthPolicy, error) {
	query := `
		SELECT id, service_name, max_response_time_ms, auto_repair_enabled, repair_strategy, 
		       alert_on_warn, alert_on_fail, policy_json, updated_at, created_at
		FROM logs.health_policies
		WHERE service_name = $1
	`

	var policy HealthPolicy
	var policyJSON sql.NullString

	err := s.db.QueryRowContext(ctx, query, serviceName).Scan(
		&policy.ID,
		&policy.ServiceName,
		&policy.MaxResponseTimeMS,
		&policy.AutoRepairEnabled,
		&policy.RepairStrategy,
		&policy.AlertOnWarn,
		&policy.AlertOnFail,
		&policyJSON,
		&policy.UpdatedAt,
		&policy.CreatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default policy if not found
		defaults := DefaultPolicies()
		if defaultPolicy, ok := defaults[serviceName]; ok {
			return &defaultPolicy, nil
		}
		return nil, fmt.Errorf("policy not found for service %s and no default available", serviceName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query policy: %w", err)
	}

	// Unmarshal policy JSON if present
	if policyJSON.Valid {
		var policyMap map[string]interface{}
		if err := json.Unmarshal([]byte(policyJSON.String), &policyMap); err != nil {
			// Log warning but don't fail
			fmt.Printf("Warning: failed to unmarshal policy JSON: %v\n", err)
		} else {
			policy.PolicyJSON = policyMap
		}
	}

	return &policy, nil
}

// GetAllPolicies retrieves all policies
func (s *HealthPolicyService) GetAllPolicies(ctx context.Context) ([]HealthPolicy, error) {
	query := `
		SELECT id, service_name, max_response_time_ms, auto_repair_enabled, repair_strategy,
		       alert_on_warn, alert_on_fail, policy_json, updated_at, created_at
		FROM logs.health_policies
		ORDER BY service_name ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	var policies []HealthPolicy

	for rows.Next() {
		var policy HealthPolicy
		var policyJSON sql.NullString

		err := rows.Scan(
			&policy.ID,
			&policy.ServiceName,
			&policy.MaxResponseTimeMS,
			&policy.AutoRepairEnabled,
			&policy.RepairStrategy,
			&policy.AlertOnWarn,
			&policy.AlertOnFail,
			&policyJSON,
			&policy.UpdatedAt,
			&policy.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		if policyJSON.Valid {
			var policyMap map[string]interface{}
			if err := json.Unmarshal([]byte(policyJSON.String), &policyMap); err == nil {
				policy.PolicyJSON = policyMap
			}
		}

		policies = append(policies, policy)
	}

	return policies, rows.Err()
}

// UpdatePolicy updates an existing policy
func (s *HealthPolicyService) UpdatePolicy(ctx context.Context, policy *HealthPolicy) error {
	var policyJSON sql.NullString

	if policy.PolicyJSON != nil && len(policy.PolicyJSON) > 0 {
		jsonBytes, err := json.Marshal(policy.PolicyJSON)
		if err != nil {
			return fmt.Errorf("failed to marshal policy JSON: %w", err)
		}
		policyJSON = sql.NullString{String: string(jsonBytes), Valid: true}
	}

	query := `
		UPDATE logs.health_policies
		SET max_response_time_ms = $1, auto_repair_enabled = $2, repair_strategy = $3,
		    alert_on_warn = $4, alert_on_fail = $5, policy_json = $6, updated_at = NOW()
		WHERE service_name = $7
	`

	result, err := s.db.ExecContext(ctx,
		query,
		policy.MaxResponseTimeMS,
		policy.AutoRepairEnabled,
		policy.RepairStrategy,
		policy.AlertOnWarn,
		policy.AlertOnFail,
		policyJSON,
		policy.ServiceName,
	)

	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Policy doesn't exist, create it
		return s.createPolicy(ctx, policy)
	}

	return nil
}

// createPolicy creates a new policy
func (s *HealthPolicyService) createPolicy(ctx context.Context, policy *HealthPolicy) error {
	var policyJSON sql.NullString

	if policy.PolicyJSON != nil && len(policy.PolicyJSON) > 0 {
		jsonBytes, err := json.Marshal(policy.PolicyJSON)
		if err != nil {
			return fmt.Errorf("failed to marshal policy JSON: %w", err)
		}
		policyJSON = sql.NullString{String: string(jsonBytes), Valid: true}
	}

	query := `
		INSERT INTO logs.health_policies
		(service_name, max_response_time_ms, auto_repair_enabled, repair_strategy, alert_on_warn, alert_on_fail, policy_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.ExecContext(ctx,
		query,
		policy.ServiceName,
		policy.MaxResponseTimeMS,
		policy.AutoRepairEnabled,
		policy.RepairStrategy,
		policy.AlertOnWarn,
		policy.AlertOnFail,
		policyJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	return nil
}

// InitializeDefaultPolicies creates default policies if they don't exist
func (s *HealthPolicyService) InitializeDefaultPolicies(ctx context.Context) error {
	defaults := DefaultPolicies()

	for _, policy := range defaults {
		// Check if policy exists
		_, err := s.GetPolicy(ctx, policy.ServiceName)
		if err != nil {
			// Policy doesn't exist, create it
			if err := s.createPolicy(ctx, &policy); err != nil {
				fmt.Printf("Warning: failed to create default policy for %s: %v\n", policy.ServiceName, err)
			}
		}
	}

	return nil
}

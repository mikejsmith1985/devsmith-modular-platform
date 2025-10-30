package services

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/healthcheck"
)

// RepairAction represents an auto-repair action taken
type RepairAction struct {
	ID              int       `json:"id"`
	Timestamp       time.Time `json:"timestamp"`
	ServiceName     string    `json:"service_name"`
	IssueType       string    `json:"issue_type"` // 'timeout', 'crash', 'dependency', 'security'
	RepairAction    string    `json:"repair_action"` // 'restart', 'rebuild', 'rollback'
	Status          string    `json:"status"` // 'pending', 'success', 'failed'
	Error           string    `json:"error,omitempty"`
	DurationMS      int64     `json:"duration_ms"`
}

// AutoRepairService determines and executes repairs for unhealthy services
type AutoRepairService struct {
	db            *sql.DB
	policyService *HealthPolicyService
}

// NewAutoRepairService creates a new auto-repair service
func NewAutoRepairService(db *sql.DB, policyService *HealthPolicyService) *AutoRepairService {
	return &AutoRepairService{
		db:            db,
		policyService: policyService,
	}
}

// AnalyzeAndRepair analyzes health report and performs repairs
func (s *AutoRepairService) AnalyzeAndRepair(ctx context.Context, report *healthcheck.HealthReport) ([]RepairAction, error) {
	var repairs []RepairAction

	// Find failed or degraded services
	for _, check := range report.Checks {
		if check.Status == healthcheck.StatusFail || check.Status == healthcheck.StatusWarn {
			service := s.extractServiceName(check.Name)
			if service == "" {
				continue
			}

			// Get policy for this service
			policy, err := s.policyService.GetPolicy(ctx, service)
			if err != nil || !policy.AutoRepairEnabled {
				continue
			}

			// Determine issue type and repair strategy
			issueType := s.classifyIssue(check)
			repairStrategy := s.determineRepairStrategy(issueType, policy)

			if repairStrategy == "none" {
				continue
			}

			// Execute repair
			action, err := s.executeRepair(ctx, service, issueType, repairStrategy)
			if err != nil {
				action.Error = err.Error()
				action.Status = "failed"
			}

			repairs = append(repairs, action)
		}
	}

	return repairs, nil
}

// extractServiceName extracts service name from check name
// e.g., "http_portal" -> "portal", "gateway_routing" -> "gateway"
func (s *AutoRepairService) extractServiceName(checkName string) string {
	parts := strings.Split(checkName, "_")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

// classifyIssue determines the type of issue based on check result
func (s *AutoRepairService) classifyIssue(check healthcheck.CheckResult) string {
	// Normalize to lowercase for case-insensitive matching
	msg := strings.ToLower(check.Message)

	if strings.Contains(msg, "timeout") || strings.Contains(msg, "refused") {
		return "timeout"
	}
	if strings.Contains(msg, "crash") || strings.Contains(msg, "stopped") {
		return "crash"
	}
	// Match both 'dependency' and 'dependent' phrasing (e.g. "Dependent service 'logs' is not responding")
	if strings.Contains(msg, "dependency") || strings.Contains(msg, "dependencies") || strings.Contains(msg, "dependent") {
		return "dependency"
	}
	if strings.Contains(msg, "critical") || strings.Contains(msg, "vulnerability") {
		return "security"
	}
	return "unknown"
}

// determineRepairStrategy uses policy and issue type to pick repair action
// This is the intelligent decision logic mentioned by the architect
func (s *AutoRepairService) determineRepairStrategy(issueType string, policy *HealthPolicy) string {
	switch issueType {
	case "timeout":
		// Timeout usually means service is hung - restart first
		return "restart"
	case "crash":
		// Container crashed - if it keeps crashing, rebuild with fresh image
		return "rebuild"
	case "dependency":
		// Dependency issue - no point restarting this service
		// Should restart the dependencies instead
		return "none"
	case "security":
		// Security vulnerability - rebuild with updated base image
		return "rebuild"
	default:
		// Unknown issue - use policy default
		return policy.RepairStrategy
	}
}

// executeRepair executes the repair action
func (s *AutoRepairService) executeRepair(ctx context.Context, service string, issueType string, strategy string) (RepairAction, error) {
	action := RepairAction{
		Timestamp:   time.Now(),
		ServiceName: service,
		IssueType:   issueType,
		RepairAction: strategy,
		Status:      "pending",
	}

	start := time.Now()
	var err error

	switch strategy {
	case "restart":
		err = s.restartService(ctx, service)
	case "rebuild":
		err = s.rebuildService(ctx, service)
	case "rollback":
		err = s.rollbackService(ctx, service)
	default:
		return action, fmt.Errorf("unknown repair strategy: %s", strategy)
	}

	action.DurationMS = time.Since(start).Milliseconds()

	if err == nil {
		action.Status = "success"
		// Log the successful repair
		s.logRepairAction(ctx, action)
	} else {
		action.Status = "failed"
		action.Error = err.Error()
	}

	return action, nil
}

// restartService restarts a Docker service
func (s *AutoRepairService) restartService(ctx context.Context, service string) error {
	cmd := exec.CommandContext(ctx, "docker-compose", "restart", service)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart service: %w", err)
	}
	return nil
}

// rebuildService rebuilds and restarts a Docker service
func (s *AutoRepairService) rebuildService(ctx context.Context, service string) error {
	cmd := exec.CommandContext(ctx, "docker-compose", "up", "-d", "--build", service)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to rebuild service: %w", err)
	}
	return nil
}

// rollbackService rolls back to the previous version
// This is typically done by reverting the image tag
func (s *AutoRepairService) rollbackService(ctx context.Context, service string) error {
	// In a real system, this would query a version history or use Docker image tags
	// For now, just restart the service
	cmd := exec.CommandContext(ctx, "docker-compose", "restart", service)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to rollback service: %w", err)
	}
	return nil
}

// logRepairAction stores the repair action in the database
func (s *AutoRepairService) logRepairAction(ctx context.Context, action RepairAction) {
	query := `
		INSERT INTO logs.auto_repairs
		(service_name, issue_type, repair_action, status, duration_ms)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.db.ExecContext(ctx,
		query,
		action.ServiceName,
		action.IssueType,
		action.RepairAction,
		action.Status,
		action.DurationMS,
	)

	if err != nil {
		fmt.Printf("Warning: failed to log repair action: %v\n", err)
	}
}

// GetRepairHistory retrieves recent repair actions
func (s *AutoRepairService) GetRepairHistory(ctx context.Context, limit int) ([]RepairAction, error) {
	query := `
		SELECT id, timestamp, service_name, issue_type, repair_action, status, error, duration_ms
		FROM logs.auto_repairs
		ORDER BY timestamp DESC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query repair history: %w", err)
	}
	defer rows.Close()

	var repairs []RepairAction

	for rows.Next() {
		var action RepairAction
		var errStr sql.NullString

		err := rows.Scan(
			&action.ID,
			&action.Timestamp,
			&action.ServiceName,
			&action.IssueType,
			&action.RepairAction,
			&action.Status,
			&errStr,
			&action.DurationMS,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan repair action: %w", err)
		}

		if errStr.Valid {
			action.Error = errStr.String
		}

		repairs = append(repairs, action)
	}

	return repairs, rows.Err()
}

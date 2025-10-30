package logs_services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"time"
)

// RepairAction represents a repair action taken
type RepairAction struct {
	ID            int       `json:"id"`
	HealthCheckID int       `json:"health_check_id,omitempty"`
	ServiceName   string    `json:"service_name"`
	IssueType     string    `json:"issue_type"`
	Action        string    `json:"action"`
	Status        string    `json:"status"`
	Error         string    `json:"error,omitempty"`
	DurationMs    int       `json:"duration_ms"`
	Timestamp     time.Time `json:"timestamp"`
}

// AutoRepairService handles intelligent service repair
type AutoRepairService struct {
	db            *sql.DB
	policyService *HealthPolicyService
}

// NewAutoRepairService creates a new auto-repair service
func NewAutoRepairService(db *sql.DB, policyService *HealthPolicyService) *AutoRepairService {
	return &AutoRepairService{db: db, policyService: policyService}
}

// AnalyzeAndRepair analyzes health check failures and repairs services
func (s *AutoRepairService) AnalyzeAndRepair(ctx context.Context, healthCheckID int, issues map[string]string) ([]RepairAction, error) {
	var actions []RepairAction

	for serviceName, issueType := range issues {
		strategy, err := s.policyService.GetRepairStrategy(ctx, serviceName, issueType)
		if err != nil || strategy == "none" {
			continue
		}

		action := RepairAction{
			HealthCheckID: healthCheckID,
			ServiceName:   serviceName,
			IssueType:     issueType,
			Action:        strategy,
			Status:        "pending",
			Timestamp:     time.Now(),
		}

		startTime := time.Now()

		// Execute repair
		switch strategy {
		case "restart":
			err = s.restartService(ctx, serviceName)
		case "rebuild":
			err = s.rebuildService(ctx, serviceName)
		default:
			err = fmt.Errorf("unknown repair strategy: %s", strategy)
		}

		action.DurationMs = int(time.Since(startTime).Milliseconds())

		if err != nil {
			action.Status = "failed"
			action.Error = err.Error()
		} else {
			action.Status = "success"
		}

		// Log repair action
		if err := s.logRepairAction(ctx, &action); err != nil {
			// Log but don't fail - repair already completed or attempted
			fmt.Printf("failed to log repair action: %v\n", err)
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// restartService restarts a service container
func (s *AutoRepairService) restartService(ctx context.Context, serviceName string) error {
	cmd := exec.CommandContext(ctx, "docker-compose", "restart", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart %s: %w", serviceName, err)
	}

	// Wait for service to be healthy
	return s.waitForServiceHealth(ctx, serviceName)
}

// rebuildService rebuilds and restarts a service
func (s *AutoRepairService) rebuildService(ctx context.Context, serviceName string) error {
	cmd := exec.CommandContext(ctx, "docker-compose", "up", "-d", "--build", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to rebuild %s: %w", serviceName, err)
	}

	// Wait for service to be healthy
	return s.waitForServiceHealth(ctx, serviceName)
}

// waitForServiceHealth waits for a service to become healthy
func (s *AutoRepairService) waitForServiceHealth(ctx context.Context, serviceName string) error {
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		// Check service health via health endpoint
		// This would be implemented with actual HTTP checks
		select {
		case <-time.After(3 * time.Second):
			// Continue checking
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// logRepairAction logs a repair action to the database
func (s *AutoRepairService) logRepairAction(ctx context.Context, action *RepairAction) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO logs.auto_repairs
		 (health_check_id, service_name, issue_type, repair_action, status, error, duration_ms, timestamp)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		action.HealthCheckID,
		action.ServiceName,
		action.IssueType,
		action.Action,
		action.Status,
		action.Error,
		action.DurationMs,
		action.Timestamp,
	)

	return err
}

// GetRepairHistory retrieves recent repair actions
func (s *AutoRepairService) GetRepairHistory(ctx context.Context, limit int) ([]RepairAction, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, health_check_id, service_name, issue_type, repair_action, status, error, duration_ms, timestamp
		 FROM logs.auto_repairs
		 ORDER BY timestamp DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query repair history: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("warning: failed to close repair history rows: %v", err)
		}
	}()

	var actions []RepairAction
	for rows.Next() {
		var action RepairAction
		err := rows.Scan(&action.ID, &action.HealthCheckID, &action.ServiceName, &action.IssueType, &action.Action, &action.Status, &action.Error, &action.DurationMs, &action.Timestamp)
		if err != nil {
			continue
		}
		actions = append(actions, action)
	}

	return actions, rows.Err()
}

// ManualRepair triggers a manual repair for a service
func (s *AutoRepairService) ManualRepair(ctx context.Context, serviceName string, strategy string) error {
	startTime := time.Now()
	var err error

	switch strategy {
	case "restart":
		err = s.restartService(ctx, serviceName)
	case "rebuild":
		err = s.rebuildService(ctx, serviceName)
	default:
		err = fmt.Errorf("unknown strategy: %s", strategy)
	}

	duration := int(time.Since(startTime).Milliseconds())

	// Log the action
	action := RepairAction{
		ServiceName: serviceName,
		Action:      strategy,
		DurationMs:  duration,
		Timestamp:   time.Now(),
	}

	if err != nil {
		action.Status = "failed"
		action.Error = err.Error()
	} else {
		action.Status = "success"
	}

	if err := s.logRepairAction(ctx, &action); err != nil {
		// Log but don't fail - repair already completed
		fmt.Printf("failed to log manual repair action: %v\n", err)
	}
	return err
}

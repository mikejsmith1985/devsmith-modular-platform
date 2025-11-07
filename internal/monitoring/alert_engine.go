// Package monitoring provides health monitoring and alerting functionality.
package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// AlertEngine monitors metrics and triggers alerts when thresholds are exceeded.
type AlertEngine struct {
	db                 *sql.DB
	thresholds         AlertThresholds
	evaluationInterval time.Duration
	ticker             *time.Ticker
	stopChan           chan struct{}
	logger             *log.Logger
}

// NewAlertEngine creates a new alert monitoring engine.
func NewAlertEngine(db *sql.DB, thresholds AlertThresholds, evaluationInterval time.Duration, logger *log.Logger) *AlertEngine {
	if logger == nil {
		logger = log.Default()
	}

	if evaluationInterval == 0 {
		evaluationInterval = 1 * time.Minute
	}

	return &AlertEngine{
		db:                 db,
		thresholds:         thresholds,
		evaluationInterval: evaluationInterval,
		stopChan:           make(chan struct{}),
		logger:             logger,
	}
}

// Start begins monitoring metrics and generating alerts.
func (e *AlertEngine) Start() {
	e.logger.Println("Alert engine started")
	e.ticker = time.NewTicker(e.evaluationInterval)

	go func() {
		// Run immediately on start
		e.evaluateMetrics()

		// Then run on interval
		for {
			select {
			case <-e.ticker.C:
				e.evaluateMetrics()
			case <-e.stopChan:
				e.logger.Println("Alert engine stopped")
				return
			}
		}
	}()
}

// Stop gracefully shuts down the alert engine.
func (e *AlertEngine) Stop() {
	if e.ticker != nil {
		e.ticker.Stop()
	}
	close(e.stopChan)
}

// evaluateMetrics checks current metrics against thresholds and creates alerts.
func (e *AlertEngine) evaluateMetrics() {
	ctx := context.Background()

	// Check error rate
	e.checkErrorRate(ctx)

	// Check response times
	e.checkResponseTimes(ctx)

	// Check service health (from health_checks table)
	e.checkServiceHealth(ctx)
}

// checkErrorRate evaluates error rate and creates alert if threshold exceeded.
func (e *AlertEngine) checkErrorRate(ctx context.Context) {
	window := e.evaluationInterval

	// Calculate error rate for the last evaluation window
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status_code >= 400) as error_count,
			COUNT(*) as total_count
		FROM monitoring.api_metrics
		WHERE timestamp >= NOW() - $1::interval
	`

	var errorCount, totalCount int64
	err := e.db.QueryRowContext(ctx, query, window).Scan(&errorCount, &totalCount)
	if err != nil {
		e.logger.Printf("Failed to query error rate: %v", err)
		return
	}

	if totalCount == 0 {
		return // No data to evaluate
	}

	// Calculate errors per minute
	errorRate := float64(errorCount) / window.Minutes()

	if errorRate > e.thresholds.APIErrorRate {
		e.createAlert(ctx, Alert{
			AlertType:   "error_rate_high",
			Severity:    "warning",
			ServiceName: "all_services",
			Message:     fmt.Sprintf("Error rate %.2f errors/min exceeds threshold %.2f errors/min", errorRate, e.thresholds.APIErrorRate),
			MetricValue: errorRate,
			Threshold:   e.thresholds.APIErrorRate,
		})
	} else {
		// Clear alert if error rate is back to normal
		e.clearAlert(ctx, "error_rate_high", "all_services")
	}
}

// checkResponseTimes evaluates P95 response time and creates alert if threshold exceeded.
func (e *AlertEngine) checkResponseTimes(ctx context.Context) {
	window := e.evaluationInterval

	// Get recent successful response times
	query := `
		SELECT response_time_ms
		FROM monitoring.api_metrics
		WHERE timestamp >= NOW() - $1::interval
		  AND status_code < 400
		ORDER BY response_time_ms ASC
	`

	rows, err := e.db.QueryContext(ctx, query, window)
	if err != nil {
		e.logger.Printf("Failed to query response times: %v", err)
		return
	}
	defer rows.Close()

	var responseTimes []int
	for rows.Next() {
		var rt int
		if err := rows.Scan(&rt); err != nil {
			e.logger.Printf("Failed to scan response time: %v", err)
			continue
		}
		responseTimes = append(responseTimes, rt)
	}

	if len(responseTimes) == 0 {
		return // No data to evaluate
	}

	// Calculate P95 (95th percentile)
	p95Index := int(float64(len(responseTimes)) * 0.95)
	if p95Index >= len(responseTimes) {
		p95Index = len(responseTimes) - 1
	}
	p95 := responseTimes[p95Index]

	if int64(p95) > e.thresholds.ResponseTimeP95 {
		e.createAlert(ctx, Alert{
			AlertType:   "response_time_high",
			Severity:    "warning",
			ServiceName: "all_services",
			Message:     fmt.Sprintf("P95 response time %dms exceeds threshold %dms", p95, e.thresholds.ResponseTimeP95),
			MetricValue: float64(p95),
			Threshold:   float64(e.thresholds.ResponseTimeP95),
		})
	} else {
		// Clear alert if response time is back to normal
		e.clearAlert(ctx, "response_time_high", "all_services")
	}
}

// checkServiceHealth evaluates service health check results.
func (e *AlertEngine) checkServiceHealth(ctx context.Context) {
	// Query recent health check results for each service
	query := `
		WITH recent_checks AS (
			SELECT 
				COALESCE(service, '') as service,
				COALESCE(overall_status, 'unknown') as status,
				timestamp as checked_at,
				ROW_NUMBER() OVER (PARTITION BY service ORDER BY timestamp DESC) as rn
			FROM logs.health_checks
			WHERE timestamp >= NOW() - INTERVAL '5 minutes'
		)
		SELECT service, status
		FROM recent_checks
		WHERE rn <= $1 AND service != ''
		ORDER BY service, rn
	`

	rows, err := e.db.QueryContext(ctx, query, e.thresholds.ServiceDown)
	if err != nil {
		e.logger.Printf("Failed to query service health: %v", err)
		return
	}
	defer rows.Close()

	// Group results by service
	serviceChecks := make(map[string][]string)
	for rows.Next() {
		var service, status string
		if err := rows.Scan(&service, &status); err != nil {
			e.logger.Printf("Failed to scan health check: %v", err)
			continue
		}
		
		// Skip empty service names
		if service == "" {
			continue
		}
		
		serviceChecks[service] = append(serviceChecks[service], status)
	}

	// Check each service for consecutive failures
	for service, statuses := range serviceChecks {
		if len(statuses) < e.thresholds.ServiceDown {
			continue // Not enough data
		}

		// Check if all recent checks failed
		allFailed := true
		for _, status := range statuses[:e.thresholds.ServiceDown] {
			if status == "healthy" {
				allFailed = false
				break
			}
		}

		if allFailed {
			severity := "critical"
			if statuses[0] == "degraded" {
				severity = "warning"
			}

			e.createAlert(ctx, Alert{
				AlertType:   "service_health_check_failed",
				Severity:    severity,
				ServiceName: service,
				Message:     fmt.Sprintf("Service %s failed %d consecutive health checks", service, e.thresholds.ServiceDown),
				MetricValue: float64(len(statuses)),
				Threshold:   float64(e.thresholds.ServiceDown),
			})
		} else {
			// Clear alert if service is healthy
			e.clearAlert(ctx, "service_health_check_failed", service)
		}
	}
}

// Alert represents a monitoring alert.
type Alert struct {
	AlertType   string
	Severity    string // critical, warning, info
	ServiceName string
	Message     string
	MetricValue float64
	Threshold   float64
}

// createAlert inserts or updates an alert in the database.
func (e *AlertEngine) createAlert(ctx context.Context, alert Alert) {
	// Check if alert already exists and is active
	var existingID int
	var existingCount int
	checkQuery := `
		SELECT id, occurrence_count
		FROM monitoring.alerts
		WHERE alert_type = $1 
		  AND service_name = $2
		  AND resolved IS NULL
	`
	err := e.db.QueryRowContext(ctx, checkQuery, alert.AlertType, alert.ServiceName).Scan(&existingID, &existingCount)

	if err == sql.ErrNoRows {
		// Create new alert
		insertQuery := `
			INSERT INTO monitoring.alerts (
				alert_type, severity, service_name, message, 
				metric_value, threshold, triggered, occurrence_count
			) VALUES ($1, $2, $3, $4, $5, $6, NOW(), 1)
		`
		_, err = e.db.ExecContext(ctx, insertQuery,
			alert.AlertType, alert.Severity, alert.ServiceName, alert.Message,
			alert.MetricValue, alert.Threshold)

		if err != nil {
			e.logger.Printf("Failed to create alert: %v", err)
			return
		}

		e.logger.Printf("ALERT CREATED: [%s] %s - %s", alert.Severity, alert.AlertType, alert.Message)
	} else if err == nil {
		// Update existing alert with new occurrence
		updateQuery := `
			UPDATE monitoring.alerts
			SET occurrence_count = occurrence_count + 1,
			    last_occurred = NOW(),
			    metric_value = $1,
			    message = $2
			WHERE id = $3
		`
		_, err = e.db.ExecContext(ctx, updateQuery, alert.MetricValue, alert.Message, existingID)

		if err != nil {
			e.logger.Printf("Failed to update alert: %v", err)
			return
		}

		e.logger.Printf("ALERT UPDATED: [%s] %s (occurrence %d)", alert.Severity, alert.AlertType, existingCount+1)
	} else {
		e.logger.Printf("Failed to check for existing alert: %v", err)
	}
}

// clearAlert resolves an active alert if it exists.
func (e *AlertEngine) clearAlert(ctx context.Context, alertType, serviceName string) {
	query := `
		UPDATE monitoring.alerts
		SET resolved = NOW()
		WHERE alert_type = $1 
		  AND service_name = $2
		  AND resolved IS NULL
		RETURNING id
	`

	var alertID int
	err := e.db.QueryRowContext(ctx, query, alertType, serviceName).Scan(&alertID)

	if err == nil {
		e.logger.Printf("ALERT RESOLVED: %s for %s (ID: %d)", alertType, serviceName, alertID)
	} else if err != sql.ErrNoRows {
		e.logger.Printf("Failed to clear alert: %v", err)
	}
}

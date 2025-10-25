# Issue #011: [COPILOT] Analytics Service - Foundation & Core Analysis

**Labels:** `copilot`, `analytics`, `data-analysis`, `visualization`
**Assignee:** Mike (with Copilot assistance)
**Created:** 2025-10-19
**Issue:** #11
**Estimated Complexity:** High
**Target Service:** analytics
**Estimated Time:** 120-150 minutes
**Depends On:** Issue #009 (Logs Service Foundation)

---

# 🚨 CRITICAL: FIRST STEP - CREATE FEATURE BRANCH 🚨

**⚠️ DO NOT PROCEED UNTIL YOU COMPLETE THIS STEP ⚠️**

## STEP 0: Verify and Create Feature Branch

**BEFORE doing ANYTHING else (reading specs, planning, writing code, or writing tests):**

### 1. Check Current Branch

Run this command FIRST:
```bash
git branch --show-current
```

**Expected output:** `development`

**If you see anything else (like a feature branch), STOP!**
- You may be continuing work on an existing branch
- Or you're on the wrong branch
- Double-check with the user before proceeding

### 2. Verify Branch is Clean

```bash
git status
```

**Expected output:** `nothing to commit, working tree clean`

**If you see uncommitted changes, STOP!**
- Commit or stash changes before creating feature branch
- Ask the user how to proceed

### 3. Update Development Branch

```bash
git pull origin development
```

### 4. Create Feature Branch

```bash
git checkout -b feature/011-analytics-service-foundation
```

**Branch naming convention:** `feature/<issue-number>-<short-description>`

### 5. Verify You're on the New Branch

```bash
git branch --show-current
```

**Expected output:** `feature/011-analytics-service-foundation`

---

**✅ ONLY AFTER COMPLETING STEP 0, proceed with reading the rest of this spec.**

---

## Task Description

Build the Analytics Service foundation to analyze log data from the Logs service. This service provides frequency analysis, trend detection, anomaly detection, and exportable reports - giving users insights into their platform usage and error patterns.

**Why This Task for Copilot:**
- Clear bounded context (statistical analysis of logs)
- Well-defined data pipeline (read from logs.entries, aggregate to analytics.aggregations)
- Standard Go patterns (Gin, pgx, scheduled jobs)
- Builds on existing Logs Service foundation

**Core Responsibility:**
Analytics service reads log data and transforms it into actionable insights through aggregation, trend detection, and anomaly analysis.

---

## Overview

### Feature Description

Implement the Analytics Service with core analysis capabilities: frequency analysis (most common errors), trend detection (patterns over time), anomaly detection (unusual spikes), and performance metrics. Provide REST API for querying analysis results and exporting reports.

### User Story

As a developer monitoring my platform, I want to see trends and anomalies in my logs so that I can identify recurring issues, performance degradation, and unusual activity patterns before they become critical problems.

### Success Criteria

- [ ] Analytics Service starts and connects to PostgreSQL
- [ ] Database schema `analytics.*` created with migrations
- [ ] Service can read from `logs.entries` table (cross-schema query)
- [ ] Frequency analysis endpoint returns top errors/warnings
- [ ] Trend detection endpoint shows metrics over time buckets (hourly)
- [ ] Anomaly detection identifies unusual spikes in error rates
- [ ] Export endpoint generates CSV/JSON reports
- [ ] Scheduled aggregation job runs every hour
- [ ] All tests pass with 70%+ coverage
- [ ] Health check endpoint works

---

## ⚠️ CRITICAL: Test-Driven Development (TDD) Required

**YOU MUST WRITE TESTS FIRST, THEN IMPLEMENTATION.**

### TDD Workflow - Analytics Service

```go
// internal/analytics/services/frequency_analyzer_test.go
func TestFrequencyAnalyzer_AnalyzeErrors_ReturnsTopErrors(t *testing.T) {
	mockRepo := new(MockLogReader)
	analyzer := NewFrequencyAnalyzer(mockRepo)

	result, err := analyzer.AnalyzeErrors(ctx, TimeWindow{Days: 7})

	assert.NoError(t, err)
	assert.Len(t, result.TopErrors, 10)
	assert.Greater(t, result.TopErrors[0].Count, result.TopErrors[1].Count)
}

// internal/analytics/services/trend_detector_test.go
func TestTrendDetector_DetectTrends_IdentifiesIncreasingPattern(t *testing.T) {
	mockRepo := new(MockAggregationRepo)
	detector := NewTrendDetector(mockRepo)

	trend, _ := detector.DetectTrends(ctx, "error_rate", TimeWindow{Days: 30})

	assert.Equal(t, "increasing", trend.Direction)
	assert.Greater(t, trend.Confidence, 0.8)
}

// internal/analytics/services/anomaly_detector_test.go
func TestAnomalyDetector_DetectAnomalies_FlagsSpikes(t *testing.T) {
	mockRepo := new(MockAggregationRepo)
	detector := NewAnomalyDetector(mockRepo)

	anomalies, _ := detector.DetectAnomalies(ctx, TimeWindow{Hours: 24})

	assert.NotEmpty(t, anomalies)
	assert.Equal(t, "spike", anomalies[0].Type)
	assert.Greater(t, anomalies[0].Severity, 2.0) // 2 std devs
}
```

**Commit tests (RED):**
```bash
git add internal/analytics/services/*_test.go
git commit -m "test(analytics): add frequency/trend/anomaly detector tests (RED phase)"
```

**Implement and commit (GREEN):**
```bash
git add internal/analytics/services/*.go
git commit -m "feat(analytics): implement analysis services (GREEN phase)"
```

**Reference:** DevsmithTDD.md lines 15-36

---

## Context for Cognitive Load Management

### Bounded Context

**Service:** Analytics
**Domain:** Statistical Analysis and Insights
**Related Entities:**
- `Aggregation` (analytics context) - Pre-computed metric aggregations (hourly buckets)
- `Trend` (analytics context) - Time-series pattern data
- `Anomaly` (analytics context) - Detected unusual patterns
- `LogEntry` (logs context, READ-ONLY) - Source data from logs.entries

**Context Boundaries:**
- ✅ **Within scope:** Statistical analysis, aggregation, trend detection, anomaly detection, report generation
- ❌ **Out of scope:** Log ingestion (Logs service), user authentication (Portal), AI explanations (Review service can integrate later)

**Why This Separation:**
Analytics service ONLY reads logs and generates insights. It does NOT write logs (Logs service), handle auth (Portal), or provide AI analysis.

**Critical: READ-ONLY access to logs schema:**
```sql
-- Analytics service can SELECT from logs.entries
-- But NEVER INSERT, UPDATE, or DELETE
GRANT SELECT ON logs.entries TO analytics_user;
```

---

### Layering

**Primary Layer:** All three layers required (Controller → Orchestration → Data)

#### Data Layer Files

```
internal/analytics/db/
├── migrations/
│   ├── 20251019_001_create_aggregations_table.sql
│   ├── 20251019_002_create_trends_table.sql
│   └── 20251019_003_create_anomalies_table.sql
├── aggregation_repository.go         # CRUD for analytics.aggregations
├── aggregation_repository_test.go    # Repository tests
├── log_reader.go                     # READ-ONLY access to logs.entries
└── log_reader_test.go                # Log reading tests
```

#### Orchestration Layer Files

```
internal/analytics/services/
├── aggregator_service.go             # Hourly aggregation job
├── aggregator_service_test.go        # Aggregation tests
├── trend_service.go                  # Trend detection logic
├── trend_service_test.go             # Trend tests
├── anomaly_service.go                # Anomaly detection logic
├── anomaly_service_test.go           # Anomaly tests
├── export_service.go                 # CSV/JSON export
└── export_service_test.go            # Export tests
```

#### Controller Layer Files

```
cmd/analytics/handlers/
├── analytics_handler.go              # Main API endpoints
├── analytics_handler_test.go         # Handler tests
└── export_handler.go                 # Export endpoints
```

**Cross-Layer Rules:**
- ✅ `analytics_handler.go` calls `trend_service.go`
- ✅ `trend_service.go` calls `aggregation_repository.go`
- ✅ `aggregator_service.go` calls both `log_reader.go` (read logs) and `aggregation_repository.go` (write aggregations)
- ❌ Handlers MUST NOT query database directly
- ❌ Services MUST NOT import handler packages
- ❌ No circular dependencies

---

## Implementation Specification

### Phase 1: Data Models

#### File: `internal/analytics/models/analytics.go`

```go
package models

import "time"

// MetricType represents the type of aggregated metric
type MetricType string

const (
	ErrorFrequency   MetricType = "error_frequency"
	WarnFrequency    MetricType = "warn_frequency"
	ResponseTime     MetricType = "response_time"
	RequestCount     MetricType = "request_count"
	ServiceActivity  MetricType = "service_activity"
)

// Aggregation represents a pre-computed metric for a time bucket
type Aggregation struct {
	ID         int64      `json:"id" db:"id"`
	MetricType MetricType `json:"metric_type" db:"metric_type"`
	Service    string     `json:"service" db:"service"`         // "portal", "review", "logs", etc.
	TimeBucket time.Time  `json:"time_bucket" db:"time_bucket"` // Hourly bucket (2025-10-19 14:00:00)
	Value      float64    `json:"value" db:"value"`             // Count, avg, sum, etc.
	Metadata   string     `json:"metadata" db:"metadata"`       // JSONB: {"top_messages": [...], "avg": 123}
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// Trend represents a detected pattern over time
type Trend struct {
	ID         int64      `json:"id" db:"id"`
	MetricType MetricType `json:"metric_type" db:"metric_type"`
	Service    string     `json:"service" db:"service"`
	StartTime  time.Time  `json:"start_time" db:"start_time"`
	EndTime    time.Time  `json:"end_time" db:"end_time"`
	Direction  string     `json:"direction" db:"direction"` // "increasing", "decreasing", "stable"
	Magnitude  float64    `json:"magnitude" db:"magnitude"`  // Rate of change
	Metadata   string     `json:"metadata" db:"metadata"`    // JSONB: additional context
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// Anomaly represents an unusual spike or dip
type Anomaly struct {
	ID           int64      `json:"id" db:"id"`
	MetricType   MetricType `json:"metric_type" db:"metric_type"`
	Service      string     `json:"service" db:"service"`
	DetectedAt   time.Time  `json:"detected_at" db:"detected_at"`
	Severity     string     `json:"severity" db:"severity"`       // "low", "medium", "high"
	ExpectedVal  float64    `json:"expected_val" db:"expected_val"`
	ActualVal    float64    `json:"actual_val" db:"actual_val"`
	Deviation    float64    `json:"deviation" db:"deviation"`     // Standard deviations from mean
	Description  string     `json:"description" db:"description"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// LogEntry represents a log from logs.entries (READ-ONLY model)
type LogEntry struct {
	ID        int64     `json:"id" db:"id"`
	UserID    *int64    `json:"user_id" db:"user_id"`
	Service   string    `json:"service" db:"service"`
	Level     string    `json:"level" db:"level"`
	Message   string    `json:"message" db:"message"`
	Metadata  string    `json:"metadata" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TrendResponse is the API response for trend analysis
type TrendResponse struct {
	MetricType MetricType              `json:"metric_type"`
	Service    string                  `json:"service"`
	TimeRange  TimeRange               `json:"time_range"`
	DataPoints []AggregationDataPoint  `json:"data_points"`
	Trend      *TrendSummary           `json:"trend,omitempty"`
}

// AggregationDataPoint represents a single point in time-series data
type AggregationDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TrendSummary provides high-level trend insights
type TrendSummary struct {
	Direction string  `json:"direction"` // "increasing", "decreasing", "stable"
	Magnitude float64 `json:"magnitude"` // Percent change
	Summary   string  `json:"summary"`   // Human-readable description
}

// TimeRange represents a time window for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// TopIssuesResponse returns most frequent errors/warnings
type TopIssuesResponse struct {
	TimeRange TimeRange   `json:"time_range"`
	Issues    []IssueItem `json:"issues"`
}

// IssueItem represents a frequent issue
type IssueItem struct {
	Service   string `json:"service"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Count     int    `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// AnomalyResponse returns detected anomalies
type AnomalyResponse struct {
	TimeRange TimeRange `json:"time_range"`
	Anomalies []Anomaly `json:"anomalies"`
}
```

---

### Phase 2: Database Layer

#### File: `internal/analytics/db/migrations/20251019_001_create_aggregations_table.sql`

```sql
-- Create analytics schema (isolated from logs and other schemas)
CREATE SCHEMA IF NOT EXISTS analytics;

-- Aggregations table: Pre-computed hourly metrics
CREATE TABLE IF NOT EXISTS analytics.aggregations (
    id BIGSERIAL PRIMARY KEY,
    metric_type VARCHAR(50) NOT NULL,
    service VARCHAR(50) NOT NULL,
    time_bucket TIMESTAMP NOT NULL,
    value NUMERIC NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Unique constraint: one aggregation per metric/service/bucket
    CONSTRAINT uq_aggregation UNIQUE (metric_type, service, time_bucket)
);

-- Indexes for common queries
CREATE INDEX idx_aggregations_metric_service ON analytics.aggregations(metric_type, service, time_bucket DESC);
CREATE INDEX idx_aggregations_time_bucket ON analytics.aggregations(time_bucket DESC);
CREATE INDEX idx_aggregations_service ON analytics.aggregations(service, time_bucket DESC);
```

#### File: `internal/analytics/db/migrations/20251019_002_create_trends_table.sql`

```sql
-- Trends table: Detected patterns over time
CREATE TABLE IF NOT EXISTS analytics.trends (
    id BIGSERIAL PRIMARY KEY,
    metric_type VARCHAR(50) NOT NULL,
    service VARCHAR(50) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    direction VARCHAR(20) NOT NULL,  -- 'increasing', 'decreasing', 'stable'
    magnitude NUMERIC NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_direction CHECK (direction IN ('increasing', 'decreasing', 'stable'))
);

CREATE INDEX idx_trends_service ON analytics.trends(service, created_at DESC);
CREATE INDEX idx_trends_time_range ON analytics.trends(start_time, end_time);
```

#### File: `internal/analytics/db/migrations/20251019_003_create_anomalies_table.sql`

```sql
-- Anomalies table: Detected unusual spikes/dips
CREATE TABLE IF NOT EXISTS analytics.anomalies (
    id BIGSERIAL PRIMARY KEY,
    metric_type VARCHAR(50) NOT NULL,
    service VARCHAR(50) NOT NULL,
    detected_at TIMESTAMP NOT NULL,
    severity VARCHAR(20) NOT NULL,  -- 'low', 'medium', 'high'
    expected_val NUMERIC NOT NULL,
    actual_val NUMERIC NOT NULL,
    deviation NUMERIC NOT NULL,  -- Standard deviations from mean
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_severity CHECK (severity IN ('low', 'medium', 'high'))
);

CREATE INDEX idx_anomalies_service ON analytics.anomalies(service, detected_at DESC);
CREATE INDEX idx_anomalies_severity ON analytics.anomalies(severity, detected_at DESC);
CREATE INDEX idx_anomalies_detected ON analytics.anomalies(detected_at DESC);
```

#### File: `internal/analytics/db/aggregation_repository.go`

```go
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

type AggregationRepository struct {
	db *pgxpool.Pool
}

func NewAggregationRepository(db *pgxpool.Pool) *AggregationRepository {
	return &AggregationRepository{db: db}
}

// Upsert creates or updates an aggregation for a time bucket
func (r *AggregationRepository) Upsert(ctx context.Context, agg *models.Aggregation) error {
	query := `
		INSERT INTO analytics.aggregations (metric_type, service, time_bucket, value, metadata)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (metric_type, service, time_bucket)
		DO UPDATE SET value = EXCLUDED.value, metadata = EXCLUDED.metadata, created_at = NOW()
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query,
		agg.MetricType,
		agg.Service,
		agg.TimeBucket,
		agg.Value,
		agg.Metadata,
	).Scan(&agg.ID, &agg.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert aggregation: %w", err)
	}
	return nil
}

// FindByRange retrieves aggregations within a time range
func (r *AggregationRepository) FindByRange(ctx context.Context, metricType models.MetricType, service string, start, end time.Time) ([]*models.Aggregation, error) {
	query := `
		SELECT id, metric_type, service, time_bucket, value, metadata, created_at
		FROM analytics.aggregations
		WHERE metric_type = $1 AND service = $2 AND time_bucket >= $3 AND time_bucket < $4
		ORDER BY time_bucket ASC
	`
	rows, err := r.db.Query(ctx, query, metricType, service, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to find aggregations: %w", err)
	}
	defer rows.Close()

	var aggregations []*models.Aggregation
	for rows.Next() {
		agg := &models.Aggregation{}
		err := rows.Scan(
			&agg.ID,
			&agg.MetricType,
			&agg.Service,
			&agg.TimeBucket,
			&agg.Value,
			&agg.Metadata,
			&agg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan aggregation: %w", err)
		}
		aggregations = append(aggregations, agg)
	}
	return aggregations, nil
}

// FindAllServices returns list of all services that have aggregations
func (r *AggregationRepository) FindAllServices(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT service FROM analytics.aggregations ORDER BY service`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find services: %w", err)
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, service)
	}
	return services, nil
}
```

#### File: `internal/analytics/db/log_reader.go`

```go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

// LogReader provides READ-ONLY access to logs.entries
type LogReader struct {
	db *pgxpool.Pool
}

func NewLogReader(db *pgxpool.Pool) *LogReader {
	return &LogReader{db: db}
}

// CountByServiceAndLevel counts log entries by service and level within a time range
func (r *LogReader) CountByServiceAndLevel(ctx context.Context, service, level string, start, end time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM logs.entries
		WHERE service = $1 AND level = $2 AND created_at >= $3 AND created_at < $4
	`
	var count int
	err := r.db.QueryRow(ctx, query, service, level, start, end).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count logs: %w", err)
	}
	return count, nil
}

// FindTopMessages finds most frequent log messages within a time range
func (r *LogReader) FindTopMessages(ctx context.Context, service, level string, start, end time.Time, limit int) ([]models.IssueItem, error) {
	query := `
		SELECT
			service,
			level,
			message,
			COUNT(*) as count,
			MIN(created_at) as first_seen,
			MAX(created_at) as last_seen
		FROM logs.entries
		WHERE service = $1 AND level = $2 AND created_at >= $3 AND created_at < $4
		GROUP BY service, level, message
		ORDER BY count DESC
		LIMIT $5
	`
	rows, err := r.db.Query(ctx, query, service, level, start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find top messages: %w", err)
	}
	defer rows.Close()

	var issues []models.IssueItem
	for rows.Next() {
		var item models.IssueItem
		err := rows.Scan(&item.Service, &item.Level, &item.Message, &item.Count, &item.FirstSeen, &item.LastSeen)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue item: %w", err)
		}
		issues = append(issues, item)
	}
	return issues, nil
}

// FindAllServices returns list of all services that have logged
func (r *LogReader) FindAllServices(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT service FROM logs.entries ORDER BY service`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find services: %w", err)
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, service)
	}
	return services, nil
}
```

---

### Phase 3: Aggregation Service (Hourly Job)

#### File: `internal/analytics/services/aggregator_service.go`

```go
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

type AggregatorService struct {
	logReader  LogReaderInterface
	aggRepo    AggregationRepositoryInterface
	ticker     *time.Ticker
	stopChan   chan bool
}

type LogReaderInterface interface {
	CountByServiceAndLevel(ctx context.Context, service, level string, start, end time.Time) (int, error)
	FindAllServices(ctx context.Context) ([]string, error)
}

type AggregationRepositoryInterface interface {
	Upsert(ctx context.Context, agg *models.Aggregation) error
}

func NewAggregatorService(logReader LogReaderInterface, aggRepo AggregationRepositoryInterface) *AggregatorService {
	return &AggregatorService{
		logReader: logReader,
		aggRepo:   aggRepo,
		stopChan:  make(chan bool),
	}
}

// Start begins the hourly aggregation job
func (s *AggregatorService) Start(ctx context.Context) {
	// Run immediately on start
	s.runAggregation(ctx)

	// Then run every hour
	s.ticker = time.NewTicker(1 * time.Hour)
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.runAggregation(ctx)
			case <-s.stopChan:
				s.ticker.Stop()
				return
			}
		}
	}()
}

// Stop halts the aggregation job
func (s *AggregatorService) Stop() {
	s.stopChan <- true
}

// runAggregation performs hourly aggregation
func (s *AggregatorService) runAggregation(ctx context.Context) {
	log.Println("Starting hourly aggregation...")

	// Get the previous hour's time bucket
	now := time.Now().UTC()
	timeBucket := now.Truncate(time.Hour).Add(-1 * time.Hour)
	start := timeBucket
	end := timeBucket.Add(1 * time.Hour)

	// Get all services
	services, err := s.logReader.FindAllServices(ctx)
	if err != nil {
		log.Printf("Failed to find services: %v", err)
		return
	}

	// Aggregate for each service and log level
	levels := []string{"error", "warn", "info", "debug"}
	for _, service := range services {
		for _, level := range levels {
			count, err := s.logReader.CountByServiceAndLevel(ctx, service, level, start, end)
			if err != nil {
				log.Printf("Failed to count logs for %s/%s: %v", service, level, err)
				continue
			}

			// Determine metric type based on level
			var metricType models.MetricType
			switch level {
			case "error":
				metricType = models.ErrorFrequency
			case "warn":
				metricType = models.WarnFrequency
			default:
				metricType = models.ServiceActivity
			}

			// Upsert aggregation
			agg := &models.Aggregation{
				MetricType: metricType,
				Service:    service,
				TimeBucket: timeBucket,
				Value:      float64(count),
				Metadata:   fmt.Sprintf(`{"level": "%s"}`, level),
			}
			if err := s.aggRepo.Upsert(ctx, agg); err != nil {
				log.Printf("Failed to upsert aggregation: %v", err)
			}
		}
	}

	log.Printf("Completed aggregation for time bucket: %s", timeBucket.Format(time.RFC3339))
}
```

---

### Phase 4: Trend Analysis Service

#### File: `internal/analytics/services/trend_service.go`

```go
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

type TrendService struct {
	aggRepo AggregationRepositoryInterface
}

type AggregationRepoReadInterface interface {
	FindByRange(ctx context.Context, metricType models.MetricType, service string, start, end time.Time) ([]*models.Aggregation, error)
}

func NewTrendService(aggRepo AggregationRepositoryInterface) *TrendService {
	return &TrendService{aggRepo: aggRepo}
}

// GetTrends analyzes trends for a metric over a time range
func (s *TrendService) GetTrends(ctx context.Context, metricType models.MetricType, service string, start, end time.Time) (*models.TrendResponse, error) {
	// Fetch aggregations
	aggregations, err := s.aggRepo.FindByRange(ctx, metricType, service, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch aggregations: %w", err)
	}

	if len(aggregations) == 0 {
		return &models.TrendResponse{
			MetricType: metricType,
			Service:    service,
			TimeRange:  models.TimeRange{Start: start, End: end},
			DataPoints: []models.AggregationDataPoint{},
			Trend:      nil,
		}, nil
	}

	// Convert to data points
	dataPoints := make([]models.AggregationDataPoint, len(aggregations))
	for i, agg := range aggregations {
		dataPoints[i] = models.AggregationDataPoint{
			Timestamp: agg.TimeBucket,
			Value:     agg.Value,
		}
	}

	// Calculate trend direction
	trendSummary := s.calculateTrend(dataPoints)

	return &models.TrendResponse{
		MetricType: metricType,
		Service:    service,
		TimeRange:  models.TimeRange{Start: start, End: end},
		DataPoints: dataPoints,
		Trend:      trendSummary,
	}, nil
}

// calculateTrend determines direction and magnitude of change
func (s *TrendService) calculateTrend(dataPoints []models.AggregationDataPoint) *models.TrendSummary {
	if len(dataPoints) < 2 {
		return &models.TrendSummary{
			Direction: "stable",
			Magnitude: 0,
			Summary:   "Insufficient data for trend analysis",
		}
	}

	// Simple linear trend: compare first and last values
	first := dataPoints[0].Value
	last := dataPoints[len(dataPoints)-1].Value

	percentChange := 0.0
	if first > 0 {
		percentChange = ((last - first) / first) * 100
	}

	direction := "stable"
	summary := "No significant change"

	if percentChange > 10 {
		direction = "increasing"
		summary = fmt.Sprintf("Increased by %.1f%%", percentChange)
	} else if percentChange < -10 {
		direction = "decreasing"
		summary = fmt.Sprintf("Decreased by %.1f%%", -percentChange)
	}

	return &models.TrendSummary{
		Direction: direction,
		Magnitude: percentChange,
		Summary:   summary,
	}
}
```

---

### Phase 5: Anomaly Detection Service

#### File: `internal/analytics/services/anomaly_service.go`

```go
package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

type AnomalyService struct {
	aggRepo AggregationRepoReadInterface
}

func NewAnomalyService(aggRepo AggregationRepoReadInterface) *AnomalyService {
	return &AnomalyService{aggRepo: aggRepo}
}

// DetectAnomalies finds unusual spikes/dips in a metric
func (s *AnomalyService) DetectAnomalies(ctx context.Context, metricType models.MetricType, service string, start, end time.Time) (*models.AnomalyResponse, error) {
	// Fetch aggregations
	aggregations, err := s.aggRepo.FindByRange(ctx, metricType, service, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch aggregations: %w", err)
	}

	if len(aggregations) < 3 {
		// Need at least 3 data points for anomaly detection
		return &models.AnomalyResponse{
			TimeRange: models.TimeRange{Start: start, End: end},
			Anomalies: []models.Anomaly{},
		}, nil
	}

	// Calculate mean and standard deviation
	mean, stdDev := s.calculateStats(aggregations)

	// Detect anomalies (values > 2 std deviations from mean)
	var anomalies []models.Anomaly
	for _, agg := range aggregations {
		deviation := (agg.Value - mean) / stdDev
		if math.Abs(deviation) > 2.0 {
			severity := "low"
			if math.Abs(deviation) > 3.0 {
				severity = "high"
			} else if math.Abs(deviation) > 2.5 {
				severity = "medium"
			}

			description := fmt.Sprintf("Value %.1f is %.1f standard deviations from mean (%.1f)", agg.Value, deviation, mean)

			anomaly := models.Anomaly{
				MetricType:  metricType,
				Service:     service,
				DetectedAt:  agg.TimeBucket,
				Severity:    severity,
				ExpectedVal: mean,
				ActualVal:   agg.Value,
				Deviation:   deviation,
				Description: description,
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return &models.AnomalyResponse{
		TimeRange: models.TimeRange{Start: start, End: end},
		Anomalies: anomalies,
	}, nil
}

// calculateStats computes mean and standard deviation
func (s *AnomalyService) calculateStats(aggregations []*models.Aggregation) (mean, stdDev float64) {
	// Calculate mean
	sum := 0.0
	for _, agg := range aggregations {
		sum += agg.Value
	}
	mean = sum / float64(len(aggregations))

	// Calculate standard deviation
	variance := 0.0
	for _, agg := range aggregations {
		variance += math.Pow(agg.Value-mean, 2)
	}
	variance /= float64(len(aggregations))
	stdDev = math.Sqrt(variance)

	return mean, stdDev
}
```

---

### Phase 6: Top Issues Service

#### File: `internal/analytics/services/top_issues_service.go`

```go
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

type TopIssuesService struct {
	logReader LogReaderInterface
}

type LogReaderTopMessagesInterface interface {
	FindTopMessages(ctx context.Context, service, level string, start, end time.Time, limit int) ([]models.IssueItem, error)
	FindAllServices(ctx context.Context) ([]string, error)
}

func NewTopIssuesService(logReader LogReaderTopMessagesInterface) *TopIssuesService {
	return &TopIssuesService{logReader: logReader}
}

// GetTopIssues returns most frequent errors and warnings
func (s *TopIssuesService) GetTopIssues(ctx context.Context, start, end time.Time, limit int) (*models.TopIssuesResponse, error) {
	// Get all services
	services, err := s.logReader.FindAllServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find services: %w", err)
	}

	var allIssues []models.IssueItem

	// Fetch top errors and warnings for each service
	for _, service := range services {
		errors, err := s.logReader.FindTopMessages(ctx, service, "error", start, end, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to find top errors for %s: %w", service, err)
		}
		allIssues = append(allIssues, errors...)

		warnings, err := s.logReader.FindTopMessages(ctx, service, "warn", start, end, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to find top warnings for %s: %w", service, err)
		}
		allIssues = append(allIssues, warnings...)
	}

	// Sort by count (descending) and limit
	sortedIssues := s.sortAndLimit(allIssues, limit)

	return &models.TopIssuesResponse{
		TimeRange: models.TimeRange{Start: start, End: end},
		Issues:    sortedIssues,
	}, nil
}

// sortAndLimit sorts issues by count and returns top N
func (s *TopIssuesService) sortAndLimit(issues []models.IssueItem, limit int) []models.IssueItem {
	// Simple bubble sort (for small datasets; use better algorithm for production)
	for i := 0; i < len(issues)-1; i++ {
		for j := 0; j < len(issues)-i-1; j++ {
			if issues[j].Count < issues[j+1].Count {
				issues[j], issues[j+1] = issues[j+1], issues[j]
			}
		}
	}

	if len(issues) > limit {
		return issues[:limit]
	}
	return issues
}
```

---

### Phase 7: Export Service

#### File: `internal/analytics/services/export_service.go`

```go
package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

type ExportService struct{}

func NewExportService() *ExportService {
	return &ExportService{}
}

// ExportTrendsAsCSV writes trend data to CSV format
func (s *ExportService) ExportTrendsAsCSV(w io.Writer, trends *models.TrendResponse) error {
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	header := []string{"Timestamp", "Value"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data points
	for _, dp := range trends.DataPoints {
		row := []string{
			dp.Timestamp.Format(time.RFC3339),
			fmt.Sprintf("%.2f", dp.Value),
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// ExportTrendsAsJSON writes trend data to JSON format
func (s *ExportService) ExportTrendsAsJSON(w io.Writer, trends *models.TrendResponse) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(trends); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}
	return nil
}

// ExportTopIssuesAsCSV writes top issues to CSV format
func (s *ExportService) ExportTopIssuesAsCSV(w io.Writer, topIssues *models.TopIssuesResponse) error {
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	header := []string{"Service", "Level", "Message", "Count", "First Seen", "Last Seen"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write issues
	for _, issue := range topIssues.Issues {
		row := []string{
			issue.Service,
			issue.Level,
			issue.Message,
			fmt.Sprintf("%d", issue.Count),
			issue.FirstSeen.Format(time.RFC3339),
			issue.LastSeen.Format(time.RFC3339),
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
```

---

### Phase 8: HTTP Handlers (Controller Layer)

#### File: `cmd/analytics/handlers/analytics_handler.go`

```go
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
)

type AnalyticsHandler struct {
	trendService     *services.TrendService
	anomalyService   *services.AnomalyService
	topIssuesService *services.TopIssuesService
	exportService    *services.ExportService
}

func NewAnalyticsHandler(
	trendService *services.TrendService,
	anomalyService *services.AnomalyService,
	topIssuesService *services.TopIssuesService,
	exportService *services.ExportService,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		trendService:     trendService,
		anomalyService:   anomalyService,
		topIssuesService: topIssuesService,
		exportService:    exportService,
	}
}

// GetTrends handles GET /api/analytics/trends
func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
	// Parse query parameters
	metricTypeStr := c.DefaultQuery("metric_type", "error_frequency")
	service := c.DefaultQuery("service", "portal")
	hours := c.DefaultQuery("hours", "24")

	hoursInt, err := strconv.Atoi(hours)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hours parameter"})
		return
	}

	// Calculate time range
	end := time.Now().UTC()
	start := end.Add(-time.Duration(hoursInt) * time.Hour)

	// Get trends
	metricType := models.MetricType(metricTypeStr)
	trends, err := h.trendService.GetTrends(c.Request.Context(), metricType, service, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trends)
}

// GetAnomalies handles GET /api/analytics/anomalies
func (h *AnalyticsHandler) GetAnomalies(c *gin.Context) {
	// Parse query parameters
	metricTypeStr := c.DefaultQuery("metric_type", "error_frequency")
	service := c.DefaultQuery("service", "portal")
	hours := c.DefaultQuery("hours", "24")

	hoursInt, err := strconv.Atoi(hours)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hours parameter"})
		return
	}

	// Calculate time range
	end := time.Now().UTC()
	start := end.Add(-time.Duration(hoursInt) * time.Hour)

	// Detect anomalies
	metricType := models.MetricType(metricTypeStr)
	anomalies, err := h.anomalyService.DetectAnomalies(c.Request.Context(), metricType, service, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, anomalies)
}

// GetTopIssues handles GET /api/analytics/top-issues
func (h *AnalyticsHandler) GetTopIssues(c *gin.Context) {
	// Parse query parameters
	hours := c.DefaultQuery("hours", "24")
	limit := c.DefaultQuery("limit", "10")

	hoursInt, err := strconv.Atoi(hours)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hours parameter"})
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	// Calculate time range
	end := time.Now().UTC()
	start := end.Add(-time.Duration(hoursInt) * time.Hour)

	// Get top issues
	topIssues, err := h.topIssuesService.GetTopIssues(c.Request.Context(), start, end, limitInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, topIssues)
}

// ExportTrends handles GET /api/analytics/export/trends
func (h *AnalyticsHandler) ExportTrends(c *gin.Context) {
	// Parse query parameters
	metricTypeStr := c.DefaultQuery("metric_type", "error_frequency")
	service := c.DefaultQuery("service", "portal")
	hours := c.DefaultQuery("hours", "24")
	format := c.DefaultQuery("format", "json") // "json" or "csv"

	hoursInt, err := strconv.Atoi(hours)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hours parameter"})
		return
	}

	// Calculate time range
	end := time.Now().UTC()
	start := end.Add(-time.Duration(hoursInt) * time.Hour)

	// Get trends
	metricType := models.MetricType(metricTypeStr)
	trends, err := h.trendService.GetTrends(c.Request.Context(), metricType, service, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Export in requested format
	if format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=trends.csv")
		if err := h.exportService.ExportTrendsAsCSV(c.Writer, trends); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	} else {
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=trends.json")
		if err := h.exportService.ExportTrendsAsJSON(c.Writer, trends); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}
```

---

### Phase 9: Main Service Entry Point

#### File: `cmd/analytics/main.go`

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mikejsmith1985/devsmith-modular-platform/cmd/analytics/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/services"
)

func main() {
	// Get database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	dbPool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Initialize repositories
	aggRepo := db.NewAggregationRepository(dbPool)
	logReader := db.NewLogReader(dbPool)

	// Initialize services
	aggregatorService := services.NewAggregatorService(logReader, aggRepo)
	trendService := services.NewTrendService(aggRepo)
	anomalyService := services.NewAnomalyService(aggRepo)
	topIssuesService := services.NewTopIssuesService(logReader)
	exportService := services.NewExportService()

	// Start hourly aggregation job
	ctx := context.Background()
	go aggregatorService.Start(ctx)
	defer aggregatorService.Stop()

	// Initialize handlers
	analyticsHandler := handlers.NewAnalyticsHandler(trendService, anomalyService, topIssuesService, exportService)

	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "analytics",
			"status":  "healthy",
		})
	})

	// API routes
	api := router.Group("/api/analytics")
	{
		api.GET("/trends", analyticsHandler.GetTrends)
		api.GET("/anomalies", analyticsHandler.GetAnomalies)
		api.GET("/top-issues", analyticsHandler.GetTopIssues)
		api.GET("/export/trends", analyticsHandler.ExportTrends)
	}

	// Get port from environment or default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	fmt.Printf("Analytics service starting on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

---

## Implementation Checklist

### Phase 1: Models ✅
- [ ] Create `internal/analytics/models/analytics.go` with all models
- [ ] Commit: `git add internal/analytics/models/ && git commit -m "feat(analytics): add data models for aggregations, trends, and anomalies"`

### Phase 2: Database Layer ✅
- [ ] Create migration files in `internal/analytics/db/migrations/`
- [ ] Create `aggregation_repository.go`
- [ ] Create `log_reader.go` (READ-ONLY access to logs.entries)
- [ ] Commit: `git add internal/analytics/db/ && git commit -m "feat(analytics): implement database layer with cross-schema read access"`

### Phase 3: Aggregation Service ✅
- [ ] Create `internal/analytics/services/aggregator_service.go`
- [ ] Test manually: Verify hourly job runs
- [ ] Commit: `git add internal/analytics/services/aggregator* && git commit -m "feat(analytics): implement hourly aggregation job"`

### Phase 4: Trend Service ✅
- [ ] Create `internal/analytics/services/trend_service.go`
- [ ] Run: `go test ./internal/analytics/services/...`
- [ ] Commit: `git add internal/analytics/services/trend* && git commit -m "feat(analytics): implement trend analysis service"`

### Phase 5: Anomaly Service ✅
- [ ] Create `internal/analytics/services/anomaly_service.go`
- [ ] Run: `go test ./internal/analytics/services/...`
- [ ] Commit: `git add internal/analytics/services/anomaly* && git commit -m "feat(analytics): implement anomaly detection service"`

### Phase 6: Top Issues Service ✅
- [ ] Create `internal/analytics/services/top_issues_service.go`
- [ ] Run: `go test ./internal/analytics/services/...`
- [ ] Commit: `git add internal/analytics/services/top_issues* && git commit -m "feat(analytics): implement top issues service"`

### Phase 7: Export Service ✅
- [ ] Create `internal/analytics/services/export_service.go`
- [ ] Run: `go test ./internal/analytics/services/...`
- [ ] Commit: `git add internal/analytics/services/export* && git commit -m "feat(analytics): implement CSV/JSON export service"`

### Phase 8: HTTP Handlers ✅
- [ ] Create `cmd/analytics/handlers/analytics_handler.go`
- [ ] Update `cmd/analytics/main.go` with routes
- [ ] Run: `go test ./cmd/analytics/handlers/...`
- [ ] Commit: `git add cmd/analytics/ && git commit -m "feat(analytics): add HTTP handlers and API routes"`

### Phase 9: Integration Testing ✅
- [ ] Start services: `make dev`
- [ ] Test health endpoint: `curl http://localhost:3000/analytics/health`
- [ ] Test trends API: `curl http://localhost:3000/analytics/api/analytics/trends?service=portal&hours=24`
- [ ] Test top issues API: `curl http://localhost:3000/analytics/api/analytics/top-issues?hours=24`
- [ ] Verify aggregation job runs (check logs)
- [ ] Verify database records created in `analytics.aggregations`

### Phase 10: Final PR ✅
- [ ] Review all commits: `git log development..HEAD --oneline`
- [ ] Run full test suite: `make test`
- [ ] Push: `git push`
- [ ] Create PR on GitHub (Title: `[Issue #011] Analytics Service - Foundation`)
- [ ] Verify CI passes
- [ ] Tag @Claude for review

---

## Environment Variables

Add to `.env.example`:

```bash
# Analytics Service
ANALYTICS_PORT=8003
```

---

## Testing Strategy

### Unit Tests (70%+ coverage required)

**Test Coverage Targets:**
- Models: 80%+
- Repositories: 75%+
- Services: 80%+
- Handlers: 70%+

**Key Test Cases:**
1. ✅ Aggregation job counts logs correctly by service/level
2. ✅ Trend service calculates direction and magnitude
3. ✅ Anomaly detection identifies spikes > 2 std deviations
4. ✅ Top issues service returns most frequent errors
5. ✅ Export service generates valid CSV and JSON
6. ✅ Handlers validate query parameters
7. ✅ Cross-schema read access works (logs.entries → analytics)

---

## Success Metrics

This issue is complete when:

1. ✅ Analytics service starts without errors
2. ✅ Health check endpoint returns 200 OK
3. ✅ Hourly aggregation job runs successfully
4. ✅ Database schema `analytics.*` created
5. ✅ Can query trends for any service/metric
6. ✅ Anomaly detection identifies spikes correctly
7. ✅ Top issues API returns frequent errors
8. ✅ Export endpoints generate CSV and JSON files
9. ✅ All unit tests pass with 70%+ coverage
10. ✅ Integration tests pass
11. ✅ CI/CD pipeline passes

---

## Cognitive Load Optimization Notes

### For Intrinsic Complexity (Simplify)
- Statistical analysis is complex → Simplified anomaly detection (2 std deviations)
- Cross-schema queries → Abstracted in `LogReader` interface
- Clear service boundaries: Analytics reads logs but doesn't write them

### For Extraneous Load (Reduce)
- No magic strings: Use `MetricType` enum constants
- Explicit time buckets: Always hourly (no variable granularity for MVP)
- Consistent naming: `GetTrends`, `DetectAnomalies`, `GetTopIssues`

### For Germane Load (Maximize)
- Follows 3-layer architecture (Controller → Service → Data)
- Respects bounded contexts (Analytics ≠ Logs)
- Scheduled job pattern clear (ticker-based hourly aggregation)
- Repository pattern enables testing without real database

---

## References

- `ARCHITECTURE.md` - Analytics Service specification (lines 1147-1165)
- `Requirements.md` - Analytics requirements (lines 438-482)
- Issue #009 - Logs Service Foundation (dependency)
- Go pgx library: https://github.com/jackc/pgx
- Chart.js (for future frontend): https://www.chartjs.org/

---

**Next Steps (For Copilot):**
1. Create feature branch: `git checkout -b feature/011-analytics-service-foundation`
2. Follow implementation checklist phase by phase
3. **Commit after each phase** (10 commits expected)
4. Test after each phase: `go test ./...`
5. Push regularly: `git push` after every 2-3 commits
6. Create PR when complete
7. Tag Claude for architecture review

**Estimated Time:** 120-150 minutes
**Test Coverage Target:** 70%+ (aim for 75%+)
**Success Metric:** Analytics service provides trends, anomalies, and top issues from log data
**Depends On:** Issue #009 (Logs Service Foundation) - Requires logs.entries table populated

package metrics

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// TimeRange represents a time period for analysis
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// MetricsSummary contains aggregated metrics
type MetricsSummary struct {
	TestPassRate        float64              `json:"test_pass_rate"`
	DeploymentFrequency float64              `json:"deployment_frequency"` // Per day
	CertificateRate     float64              `json:"certificate_rate"`     // Per day
	RuleViolations      int                  `json:"rule_violations"`
	AvgServiceHealth    float64              `json:"avg_service_health"`
	AvgResponseTime     float64              `json:"avg_response_time"`
	Trends              map[string]TrendData `json:"trends"`
	TopViolations       []ViolationSummary   `json:"top_violations"`
}

// TrendData represents trend information for a metric
type TrendData struct {
	Direction string  `json:"direction"` // "up", "down", "stable"
	Change    float64 `json:"change"`    // Percentage change
	Recent    float64 `json:"recent"`    // Most recent value
	Previous  float64 `json:"previous"`  // Previous value
}

// ViolationSummary represents aggregated violation data
type ViolationSummary struct {
	Rule      string    `json:"rule"`
	Count     int       `json:"count"`
	Severity  string    `json:"severity"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// Analyzer analyzes collected metrics
type Analyzer struct {
	metricsDir string
}

// NewAnalyzer creates a new metrics analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		metricsDir: filepath.Join("test-results", "metrics"),
	}
}

// Analyze generates a summary for the given time range
func (a *Analyzer) Analyze(timeRange TimeRange) (*MetricsSummary, error) {
	metrics, err := a.loadMetrics(timeRange)
	if err != nil {
		return nil, err
	}

	summary := &MetricsSummary{
		Trends: make(map[string]TrendData),
	}

	// Analyze test runs
	testMetrics := filterByType(metrics, MetricTestRun)
	if len(testMetrics) > 0 {
		summary.TestPassRate = calculateAverage(testMetrics)
		summary.Trends["test_pass_rate"] = calculateTrend(testMetrics)
	}

	// Analyze deployments
	deployMetrics := filterByType(metrics, MetricDeployment)
	days := timeRange.End.Sub(timeRange.Start).Hours() / 24
	if days > 0 {
		summary.DeploymentFrequency = float64(len(deployMetrics)) / days
	}

	// Analyze certificates
	certMetrics := filterByType(metrics, MetricCertificate)
	if days > 0 {
		summary.CertificateRate = float64(len(certMetrics)) / days
	}

	// Analyze rule violations
	violationMetrics := filterByType(metrics, MetricRuleViolation)
	summary.RuleViolations = len(violationMetrics)
	summary.TopViolations = aggregateViolations(violationMetrics)

	// Analyze service health
	healthMetrics := filterByType(metrics, MetricServiceHealth)
	if len(healthMetrics) > 0 {
		summary.AvgServiceHealth = calculateAverage(healthMetrics) * 100

		// Calculate average response time
		var totalResponseTime float64
		var responseTimeCount int
		for _, m := range healthMetrics {
			if rt, ok := m.Metadata["response_time_ms"].(float64); ok {
				totalResponseTime += rt
				responseTimeCount++
			}
		}
		if responseTimeCount > 0 {
			summary.AvgResponseTime = totalResponseTime / float64(responseTimeCount)
		}
	}

	return summary, nil
}

// loadMetrics reads all metrics from the time range
func (a *Analyzer) loadMetrics(timeRange TimeRange) ([]Metric, error) {
	var allMetrics []Metric

	// Iterate through each day in the range
	for d := timeRange.Start; d.Before(timeRange.End) || d.Equal(timeRange.End); d = d.AddDate(0, 0, 1) {
		filename := filepath.Join(a.metricsDir, d.Format("2006-01-02")+".jsonl")

		file, err := os.Open(filename)
		if err != nil {
			if os.IsNotExist(err) {
				continue // No metrics for this day
			}
			return nil, err
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var metric Metric
			if err := json.Unmarshal(scanner.Bytes(), &metric); err != nil {
				file.Close()
				return nil, err
			}

			// Filter by time range
			if !metric.Timestamp.Before(timeRange.Start) && !metric.Timestamp.After(timeRange.End) {
				allMetrics = append(allMetrics, metric)
			}
		}
		file.Close()

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return allMetrics, nil
}

// filterByType returns metrics of a specific type
func filterByType(metrics []Metric, metricType MetricType) []Metric {
	var filtered []Metric
	for _, m := range metrics {
		if m.Type == metricType {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// calculateAverage calculates the average value
func calculateAverage(metrics []Metric) float64 {
	if len(metrics) == 0 {
		return 0
	}

	var sum float64
	for _, m := range metrics {
		sum += m.Value
	}
	return sum / float64(len(metrics))
}

// calculateTrend determines the trend direction and change
func calculateTrend(metrics []Metric) TrendData {
	if len(metrics) < 2 {
		return TrendData{Direction: "stable", Change: 0}
	}

	// Sort by timestamp
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Timestamp.Before(metrics[j].Timestamp)
	})

	// Split into two halves
	mid := len(metrics) / 2
	firstHalf := metrics[:mid]
	secondHalf := metrics[mid:]

	avgFirst := calculateAverage(firstHalf)
	avgSecond := calculateAverage(secondHalf)

	change := 0.0
	if avgFirst > 0 {
		change = ((avgSecond - avgFirst) / avgFirst) * 100
	}

	direction := "stable"
	if change > 5 {
		direction = "up"
	} else if change < -5 {
		direction = "down"
	}

	return TrendData{
		Direction: direction,
		Change:    change,
		Recent:    avgSecond,
		Previous:  avgFirst,
	}
}

// aggregateViolations groups violations by rule
func aggregateViolations(metrics []Metric) []ViolationSummary {
	violationMap := make(map[string]*ViolationSummary)

	for _, m := range metrics {
		rule, _ := m.Metadata["rule"].(string)
		severity, _ := m.Metadata["severity"].(string)

		if existing, ok := violationMap[rule]; ok {
			existing.Count++
			if m.Timestamp.After(existing.LastSeen) {
				existing.LastSeen = m.Timestamp
			}
			if m.Timestamp.Before(existing.FirstSeen) {
				existing.FirstSeen = m.Timestamp
			}
		} else {
			violationMap[rule] = &ViolationSummary{
				Rule:      rule,
				Count:     1,
				Severity:  severity,
				FirstSeen: m.Timestamp,
				LastSeen:  m.Timestamp,
			}
		}
	}

	// Convert to slice and sort by count
	var violations []ViolationSummary
	for _, v := range violationMap {
		violations = append(violations, *v)
	}

	sort.Slice(violations, func(i, j int) bool {
		return violations[i].Count > violations[j].Count
	})

	// Return top 10
	if len(violations) > 10 {
		violations = violations[:10]
	}

	return violations
}

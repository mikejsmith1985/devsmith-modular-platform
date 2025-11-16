package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// MetricType represents different types of metrics we track
type MetricType string

const (
	MetricTestRun       MetricType = "test_run"
	MetricDeployment    MetricType = "deployment"
	MetricCertificate   MetricType = "certificate"
	MetricRuleViolation MetricType = "rule_violation"
	MetricServiceHealth MetricType = "service_health"
)

// Metric represents a single measurement
type Metric struct {
	Type      MetricType             `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Success   bool                   `json:"success"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Collector handles metric collection and storage
type Collector struct {
	metricsDir string
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	metricsDir := filepath.Join("test-results", "metrics")
	os.MkdirAll(metricsDir, 0755)

	return &Collector{
		metricsDir: metricsDir,
	}
}

// Record saves a metric to disk (JSONL format)
func (c *Collector) Record(metric Metric) error {
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	// One file per day: test-results/metrics/2025-11-14.jsonl
	filename := filepath.Join(
		c.metricsDir,
		metric.Timestamp.Format("2006-01-02")+".jsonl",
	)

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	_, err = file.Write(append(data, '\n'))
	return err
}

// RecordTestRun records a test execution
func (c *Collector) RecordTestRun(passed, failed int, duration time.Duration) error {
	total := passed + failed
	passRate := 0.0
	if total > 0 {
		passRate = float64(passed) / float64(total) * 100
	}

	return c.Record(Metric{
		Type:    MetricTestRun,
		Value:   passRate,
		Success: failed == 0,
		Metadata: map[string]interface{}{
			"passed":       passed,
			"failed":       failed,
			"total":        total,
			"duration_sec": duration.Seconds(),
		},
	})
}

// RecordDeployment records a deployment event
func (c *Collector) RecordDeployment(service string, success bool, duration time.Duration) error {
	value := 0.0
	if success {
		value = 1.0
	}

	return c.Record(Metric{
		Type:    MetricDeployment,
		Value:   value,
		Success: success,
		Metadata: map[string]interface{}{
			"service":      service,
			"duration_sec": duration.Seconds(),
		},
	})
}

// RecordCertificateGeneration records certificate creation
func (c *Collector) RecordCertificateGeneration(success bool) error {
	value := 0.0
	if success {
		value = 1.0
	}

	return c.Record(Metric{
		Type:    MetricCertificate,
		Value:   value,
		Success: success,
	})
}

// RecordRuleViolation records when a rule is violated
func (c *Collector) RecordRuleViolation(rule string, severity string) error {
	return c.Record(Metric{
		Type:    MetricRuleViolation,
		Value:   1.0,
		Success: false,
		Metadata: map[string]interface{}{
			"rule":     rule,
			"severity": severity,
		},
	})
}

// RecordServiceHealth records service health status
func (c *Collector) RecordServiceHealth(service string, healthy bool, responseTime time.Duration) error {
	value := 0.0
	if healthy {
		value = 1.0
	}

	return c.Record(Metric{
		Type:    MetricServiceHealth,
		Value:   value,
		Success: healthy,
		Metadata: map[string]interface{}{
			"service":           service,
			"response_time_ms": responseTime.Milliseconds(),
		},
	})
}

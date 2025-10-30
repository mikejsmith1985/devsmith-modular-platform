package analytics_models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricTypeConstants(t *testing.T) {
	assert.Equal(t, MetricType("error_frequency"), ErrorFrequency)
	assert.Equal(t, MetricType("service_activity"), ServiceActivity)
}

func TestAggregation_Structure(t *testing.T) {
	now := time.Now()
	agg := Aggregation{
		TimeBucket: now,
		MetricType: ErrorFrequency,
		Service:    "portal",
		Value:      42.5,
	}

	assert.Equal(t, now, agg.TimeBucket)
	assert.Equal(t, ErrorFrequency, agg.MetricType)
	assert.Equal(t, "portal", agg.Service)
	assert.Equal(t, 42.5, agg.Value)
}

func TestTrend_Structure(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Hour)
	trend := Trend{
		StartTime:  start,
		EndTime:    end,
		MetricType: ServiceActivity,
		Service:    "review",
		Direction:  "increasing",
		Confidence: 0.95,
	}

	assert.Equal(t, start, trend.StartTime)
	assert.Equal(t, end, trend.EndTime)
	assert.Equal(t, ServiceActivity, trend.MetricType)
	assert.Equal(t, "review", trend.Service)
	assert.Equal(t, "increasing", trend.Direction)
	assert.Equal(t, 0.95, trend.Confidence)
}

func TestAnomaly_Structure(t *testing.T) {
	now := time.Now()
	anomaly := Anomaly{
		TimeBucket: now,
		MetricType: ErrorFrequency,
		Service:    "logs",
		Severity:   "high",
		Value:      100.5,
		ZScore:     3.2,
	}

	assert.Equal(t, now, anomaly.TimeBucket)
	assert.Equal(t, ErrorFrequency, anomaly.MetricType)
	assert.Equal(t, "logs", anomaly.Service)
	assert.Equal(t, "high", anomaly.Severity)
	assert.Equal(t, 100.5, anomaly.Value)
	assert.Equal(t, 3.2, anomaly.ZScore)
}

func TestLogEntry_Structure(t *testing.T) {
	now := time.Now()
	entry := LogEntry{
		CreatedAt: now,
		Service:   "analytics",
		Level:     "ERROR",
		Message:   "Test message",
	}

	assert.Equal(t, now, entry.CreatedAt)
	assert.Equal(t, "analytics", entry.Service)
	assert.Equal(t, "ERROR", entry.Level)
	assert.Equal(t, "Test message", entry.Message)
}

func TestTrendResponse_Structure(t *testing.T) {
	summary := &TrendSummary{
		Direction:  "increasing",
		Summary:    "Trend is increasing",
		Confidence: 0.9,
	}

	response := TrendResponse{
		Trend:      summary,
		MetricType: ErrorFrequency,
		Service:    "portal",
	}

	assert.NotNil(t, response.Trend)
	assert.Equal(t, ErrorFrequency, response.MetricType)
	assert.Equal(t, "portal", response.Service)
}

func TestAggregationDataPoint_Structure(t *testing.T) {
	now := time.Now()
	point := AggregationDataPoint{
		Timestamp: now,
		Value:     50.0,
	}

	assert.Equal(t, now, point.Timestamp)
	assert.Equal(t, 50.0, point.Value)
}

func TestTrendSummary_Structure(t *testing.T) {
	summary := TrendSummary{
		Direction:  "decreasing",
		Summary:    "Error rates are decreasing",
		Confidence: 0.85,
	}

	assert.Equal(t, "decreasing", summary.Direction)
	assert.Equal(t, "Error rates are decreasing", summary.Summary)
	assert.Greater(t, summary.Confidence, 0.0)
}

func TestTimeRange_Structure(t *testing.T) {
	start := time.Now()
	end := start.Add(24 * time.Hour)

	tr := TimeRange{
		Start: start,
		End:   end,
	}

	assert.True(t, tr.End.After(tr.Start))
}

func TestIssueItem_Structure(t *testing.T) {
	now := time.Now()
	item := IssueItem{
		LastSeen: now,
		Service:  "review",
		Level:    "WARN",
		Message:  "Low memory",
		Count:    5,
		Value:    25.0,
	}

	assert.Equal(t, now, item.LastSeen)
	assert.Equal(t, "review", item.Service)
	assert.Equal(t, "WARN", item.Level)
	assert.Equal(t, "Low memory", item.Message)
	assert.Equal(t, 5, item.Count)
	assert.Equal(t, 25.0, item.Value)
}

func TestTopIssuesResponse_Structure(t *testing.T) {
	start := time.Now()
	response := TopIssuesResponse{
		TimeRange: TimeRange{Start: start, End: start.Add(time.Hour)},
		Issues: []IssueItem{
			{Service: "portal", Count: 10},
			{Service: "review", Count: 5},
		},
	}

	assert.Len(t, response.Issues, 2)
	assert.Equal(t, 10, response.Issues[0].Count)
}

func TestAnomalyResponse_Structure(t *testing.T) {
	now := time.Now()
	response := AnomalyResponse{
		TimeRange: TimeRange{Start: now, End: now.Add(time.Hour)},
		Anomalies: []Anomaly{
			{Service: "logs", Severity: "critical"},
		},
	}

	assert.Len(t, response.Anomalies, 1)
}

func TestTrendAnalysis_Structure(t *testing.T) {
	start := time.Now()
	analysis := TrendAnalysis{
		Start:            start,
		End:              start.Add(time.Hour),
		MetricType:       ServiceActivity,
		Service:          "analytics",
		Direction:        "increasing",
		Change:           15.0,
		PercentageChange: 12.5,
	}

	assert.Equal(t, start, analysis.Start)
	assert.True(t, analysis.End.After(start))
	assert.Equal(t, ServiceActivity, analysis.MetricType)
	assert.Equal(t, "analytics", analysis.Service)
	assert.Equal(t, "increasing", analysis.Direction)
	assert.Equal(t, 15.0, analysis.Change)
	assert.Equal(t, 12.5, analysis.PercentageChange)
}

func TestTimeVariables(t *testing.T) {
	assert.Equal(t, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), MinTime)
	assert.Equal(t, time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC), MaxTime)
	assert.True(t, MaxTime.After(MinTime))
}

func TestZeroValues(t *testing.T) {
	var agg Aggregation
	assert.Equal(t, time.Time{}, agg.CreatedAt)
	assert.Equal(t, MetricType(""), agg.MetricType)
	assert.Equal(t, 0.0, agg.Value)

	var trend Trend
	assert.Equal(t, time.Time{}, trend.StartTime)
	assert.Equal(t, 0.0, trend.Confidence)
}

func TestServiceAggregation_Structure(t *testing.T) {
	agg := ServiceAggregation{
		ServiceName:      "portal",
		AggregationValue: 42.5,
	}

	assert.Equal(t, "portal", agg.ServiceName)
	assert.Equal(t, 42.5, agg.AggregationValue)
}

func TestServiceAggregation_MultipleInstances(t *testing.T) {
	agg1 := ServiceAggregation{ServiceName: "portal", AggregationValue: 100}
	agg2 := ServiceAggregation{ServiceName: "review", AggregationValue: 200}

	assert.NotEqual(t, agg1.ServiceName, agg2.ServiceName)
	assert.NotEqual(t, agg1.AggregationValue, agg2.AggregationValue)
}

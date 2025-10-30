package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMetricsChecker_measureEndpoint(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // Simulate some processing time
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := &MetricsChecker{
		CheckName: "test_metrics",
	}

	endpoint := MetricEndpoint{
		Name: "test_service",
		URL:  server.URL,
	}

	metric := checker.measureEndpoint(endpoint)

	if metric.Endpoint != "test_service" {
		t.Errorf("Expected endpoint name test_service, got %s", metric.Endpoint)
	}

	if metric.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", metric.StatusCode)
	}

	if metric.Status != "ok" {
		t.Errorf("Expected status ok, got %s", metric.Status)
	}

	if metric.ResponseTime < 50 {
		t.Errorf("Expected response time >= 50ms, got %dms", metric.ResponseTime)
	}
}

func TestMetricsChecker_measureEndpoint_Timeout(t *testing.T) {
	// Create a slow server that will timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // Longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := &MetricsChecker{
		CheckName: "test_metrics",
	}

	endpoint := MetricEndpoint{
		Name: "slow_service",
		URL:  server.URL,
	}

	metric := checker.measureEndpoint(endpoint)

	if metric.Status != "timeout" {
		t.Errorf("Expected status timeout, got %s", metric.Status)
	}
}

func TestMetricsChecker_Check(t *testing.T) {
	// Create test servers with different response times
	fastServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer fastServer.Close()

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1100 * time.Millisecond) // Over slow threshold
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	checker := &MetricsChecker{
		CheckName: "test_metrics",
		Endpoints: []MetricEndpoint{
			{Name: "fast", URL: fastServer.URL},
			{Name: "slow", URL: slowServer.URL},
		},
	}

	result := checker.Check()

	if result.Status != StatusWarn {
		t.Errorf("Expected status warn (due to slow endpoint), got %s", result.Status)
	}

	metrics, ok := result.Details["metrics"].([]PerformanceMetric)
	if !ok {
		t.Fatal("Expected metrics in details")
	}

	if len(metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(metrics))
	}
}

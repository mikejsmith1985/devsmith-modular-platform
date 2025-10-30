package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDependencyChecker_isServiceHealthy(t *testing.T) {
	// Create a healthy service
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer healthyServer.Close()

	// Create an unhealthy service
	unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer unhealthyServer.Close()

	checker := &DependencyChecker{
		CheckName: "test_deps",
	}

	// Test healthy service
	if !checker.isServiceHealthy(healthyServer.URL) {
		t.Error("Expected healthy service to return true")
	}

	// Test unhealthy service
	if checker.isServiceHealthy(unhealthyServer.URL) {
		t.Error("Expected unhealthy service to return false")
	}
}

func TestDependencyChecker_Check(t *testing.T) {
	// Create test servers
	portalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer portalServer.Close()

	reviewServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer reviewServer.Close()

	logsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable) // Unhealthy
	}))
	defer logsServer.Close()

	checker := &DependencyChecker{
		CheckName: "test_deps",
		Dependencies: map[string][]string{
			"portal": {},
			"review": {"portal", "logs"}, // review depends on portal and logs
			"logs":   {"portal"},
		},
		HealthChecks: map[string]string{
			"portal": portalServer.URL,
			"review": reviewServer.URL,
			"logs":   logsServer.URL,
		},
	}

	result := checker.Check()

	// logs is unhealthy, so review should be degraded
	if result.Status == StatusPass {
		t.Error("Expected status to not be pass (logs service is unhealthy)")
	}

	depStatuses, ok := result.Details["dependency_status"].([]ServiceDependency)
	if !ok {
		t.Fatal("Expected dependency_status in details")
	}

	// Find review service status
	var reviewStatus string
	for _, dep := range depStatuses {
		if dep.Service == "review" {
			reviewStatus = dep.Status
			if dep.TotalDeps != 2 {
				t.Errorf("Expected review to have 2 dependencies, got %d", dep.TotalDeps)
			}
		}
	}

	if reviewStatus != "degraded" {
		t.Errorf("Expected review status to be degraded, got %s", reviewStatus)
	}
}

func TestDependencyChecker_Check_AllHealthy(t *testing.T) {
	// All services healthy
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := &DependencyChecker{
		CheckName: "test_deps",
		Dependencies: map[string][]string{
			"portal": {},
			"review": {"portal"},
		},
		HealthChecks: map[string]string{
			"portal": server.URL,
			"review": server.URL,
		},
	}

	result := checker.Check()

	if result.Status != StatusPass {
		t.Errorf("Expected status pass when all healthy, got %s", result.Status)
	}
}

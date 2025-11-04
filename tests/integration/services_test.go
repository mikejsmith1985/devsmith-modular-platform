//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllServicesHealthy(t *testing.T) {
	services := []struct {
		name string
		url  string
		port string
	}{
		{"Portal", "http://localhost:8080/health", "8080"},
		{"Review", "http://localhost:8081/health", "8081"},
		{"Logs", "http://localhost:8082/health", "8082"},
		{"Analytics", "http://localhost:8083/health", "8083"},
	}

	for _, service := range services {
		t.Run(service.name, func(t *testing.T) {
			resp, err := http.Get(service.url)
			require.NoError(t, err, "%s service should be reachable at %s", service.name, service.url)
			assert.Equal(t, http.StatusOK, resp.StatusCode, "%s health check failed", service.name)
			defer resp.Body.Close()
		})
	}
}

func TestPortalHealthCheck(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/health")
	require.NoError(t, err, "Portal service should be reachable")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
	resp.Body.Close()
}

func TestReviewHealthCheck(t *testing.T) {
	resp, err := http.Get("http://localhost:8081/health")
	require.NoError(t, err, "Review service should be reachable")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestLogsHealthCheck(t *testing.T) {
	resp, err := http.Get("http://localhost:8082/health")
	require.NoError(t, err, "Logs service should be reachable")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestAnalyticsHealthCheck(t *testing.T) {
	resp, err := http.Get("http://localhost:8083/health")
	require.NoError(t, err, "Analytics service should be reachable")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

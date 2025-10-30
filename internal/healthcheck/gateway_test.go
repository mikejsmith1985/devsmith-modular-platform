package healthcheck

import (
	"os"
	"testing"
)

func TestGatewayChecker_parseNginxConfig(t *testing.T) {
	// Create a temporary nginx config for testing
	tmpfile, err := os.CreateTemp("", "nginx-test-*.conf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	config := `
server {
    listen 80;
    
    location / {
        proxy_pass http://portal/;
    }
    
    location /api/review {
        proxy_pass http://review/api/review;
    }
    
    location /api/logs {
        proxy_pass http://logs/api/logs;
    }
}
`
	if _, err := tmpfile.Write([]byte(config)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	checker := &GatewayChecker{
		CheckName:  "test_gateway",
		ConfigPath: tmpfile.Name(),
		GatewayURL: "http://localhost:3000",
	}

	routes, err := checker.parseNginxConfig()
	if err != nil {
		t.Fatalf("parseNginxConfig failed: %v", err)
	}

	expectedRoutes := 3
	if len(routes) != expectedRoutes {
		t.Errorf("Expected %d routes, got %d", expectedRoutes, len(routes))
	}

	// Check first route
	if len(routes) > 0 {
		if routes[0].Path != "/" {
			t.Errorf("Expected path /, got %s", routes[0].Path)
		}
		if routes[0].TargetService != "portal" {
			t.Errorf("Expected target portal, got %s", routes[0].TargetService)
		}
	}
}

func TestGatewayChecker_parseNginxConfig_EmptyFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "nginx-empty-*.conf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	checker := &GatewayChecker{
		CheckName:  "test_gateway",
		ConfigPath: tmpfile.Name(),
		GatewayURL: "http://localhost:3000",
	}

	routes, err := checker.parseNginxConfig()
	if err != nil {
		t.Fatalf("parseNginxConfig failed: %v", err)
	}

	if len(routes) != 0 {
		t.Errorf("Expected 0 routes from empty file, got %d", len(routes))
	}
}

func TestGatewayChecker_parseNginxConfig_FileNotFound(t *testing.T) {
	checker := &GatewayChecker{
		CheckName:  "test_gateway",
		ConfigPath: "/nonexistent/nginx.conf",
		GatewayURL: "http://localhost:3000",
	}

	_, err := checker.parseNginxConfig()
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

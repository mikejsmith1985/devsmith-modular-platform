package healthcheck

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// GatewayChecker validates nginx gateway routing configuration
type GatewayChecker struct {
	CheckName  string
	ConfigPath string
	GatewayURL string
}

// Name returns the checker name
func (c *GatewayChecker) Name() string {
	return c.CheckName
}

// RouteMapping represents a discovered nginx route
type RouteMapping struct {
	Path          string
	ProxyPass     string
	TargetService string
}

// Check validates gateway routing configuration
func (c *GatewayChecker) Check() CheckResult {
	start := time.Now()
	result := CheckResult{
		Name:      c.CheckName,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Parse nginx configuration
	routes, err := c.parseNginxConfig()
	if err != nil {
		result.Status = StatusFail
		result.Message = "Failed to parse nginx configuration"
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	result.Details["routes_discovered"] = len(routes)

	// Validate each route
	validRoutes := 0
	invalidRoutes := []string{}
	routeDetails := []map[string]string{}

	for _, route := range routes {
		// Test if the route responds
		testURL := c.GatewayURL + route.Path
		if c.testRoute(testURL) {
			validRoutes++
			routeDetails = append(routeDetails, map[string]string{
				"path":    route.Path,
				"target":  route.TargetService,
				"status":  "ok",
			})
		} else {
			invalidRoutes = append(invalidRoutes, route.Path)
			routeDetails = append(routeDetails, map[string]string{
				"path":    route.Path,
				"target":  route.TargetService,
				"status":  "failed",
			})
		}
	}

	result.Details["valid_routes"] = validRoutes
	result.Details["invalid_routes"] = len(invalidRoutes)
	result.Details["route_details"] = routeDetails

	if len(invalidRoutes) > 0 {
		result.Status = StatusWarn
		result.Message = fmt.Sprintf("%d/%d routes responding", validRoutes, len(routes))
		result.Error = fmt.Sprintf("Failed routes: %s", strings.Join(invalidRoutes, ", "))
	} else if len(routes) == 0 {
		result.Status = StatusWarn
		result.Message = "No routes discovered in nginx config"
	} else {
		result.Status = StatusPass
		result.Message = fmt.Sprintf("All %d gateway routes responding", validRoutes)
	}

	result.Duration = time.Since(start)
	return result
}

// parseNginxConfig extracts route mappings from nginx.conf
func (c *GatewayChecker) parseNginxConfig() ([]RouteMapping, error) {
	file, err := os.Open(c.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open nginx config: %w", err)
	}
	defer file.Close()

	var routes []RouteMapping
	var currentLocation string
	
	// Regex patterns
	locationPattern := regexp.MustCompile(`location\s+([\S]+)\s+\{`)
	proxyPassPattern := regexp.MustCompile(`proxy_pass\s+http://([^/]+)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Match location block
		if matches := locationPattern.FindStringSubmatch(line); len(matches) > 1 {
			currentLocation = matches[1]
		}

		// Match proxy_pass directive
		if matches := proxyPassPattern.FindStringSubmatch(line); len(matches) > 1 && currentLocation != "" {
			targetService := matches[1]
			routes = append(routes, RouteMapping{
				Path:          currentLocation,
				ProxyPass:     line,
				TargetService: targetService,
			})
			currentLocation = "" // Reset for next location block
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading nginx config: %w", err)
	}

	return routes, nil
}

// testRoute performs a simple HTTP test of a gateway route
func (c *GatewayChecker) testRoute(url string) bool {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Consider 2xx, 3xx, 401, 403 as "working" (route exists and responds)
	// 404 means route doesn't exist
	// 5xx means service error (but route is configured)
	return resp.StatusCode != 404
}


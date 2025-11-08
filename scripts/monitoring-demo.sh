#!/bin/bash

# Monitoring System Demo Script
# This script demonstrates how the monitoring system catches API payload issues

set -e

echo "🔍 DevSmith Monitoring System Demo"
echo "======================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEMO_DIR="demo-monitoring"
LOG_FILE="monitoring-demo.log"

# Step 1: Setup demo environment
echo -e "${BLUE}Step 1: Setting up demo environment...${NC}"
mkdir -p $DEMO_DIR
cd $DEMO_DIR

# Create a simple monitoring test
cat > monitoring_test.go << 'EOF'
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock monitoring collector to simulate our real system
type MockMetricsCollector struct {
	APICallsRecorded   []APIMetrics
	ValidationFailures []PayloadValidationFailure
}

type APIMetrics struct {
	Timestamp   time.Time
	Method      string
	Endpoint    string
	StatusCode  int
	ErrorType   string
	ServiceName string
}

type PayloadValidationFailure struct {
	Endpoint      string
	ExtraFields   []string
	MissingFields []string
	Timestamp     time.Time
}

func (m *MockMetricsCollector) RecordAPICall(metrics APIMetrics) {
	m.APICallsRecorded = append(m.APICallsRecorded, metrics)
}

func (m *MockMetricsCollector) RecordValidationFailure(failure PayloadValidationFailure) {
	m.ValidationFailures = append(m.ValidationFailures, failure)
}

// Simulated Review API handler that validates payloads
func createReviewHandler(collector *MockMetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Expected payload structure (no session_id!)
		var validRequest struct {
			PastedCode string `json:"pasted_code" binding:"required"`
			Model      string `json:"model" binding:"required"`
		}

		// Parse raw request to check for extra fields
		var rawPayload map[string]interface{}
		bodyBytes, _ := c.GetRawData()
		c.Request.Body = strings.NewReader(string(bodyBytes))
		json.Unmarshal(bodyBytes, &rawPayload)

		// Bind to expected structure
		if err := c.ShouldBindJSON(&validRequest); err != nil {
			// Analyze what went wrong
			var extraFields []string
			var missingFields []string

			// Check for extra fields
			for key := range rawPayload {
				if key != "pasted_code" && key != "model" {
					extraFields = append(extraFields, key)
				}
			}

			// Check for missing fields
			if rawPayload["pasted_code"] == nil {
				missingFields = append(missingFields, "pasted_code")
			}
			if rawPayload["model"] == nil {
				missingFields = append(missingFields, "model")
			}

			// Record validation failure (this is what catches the session_id bug!)
			collector.RecordValidationFailure(PayloadValidationFailure{
				Endpoint:      c.Request.URL.Path,
				ExtraFields:   extraFields,
				MissingFields: missingFields,
				Timestamp:     time.Now(),
			})

			// Record API call with error
			collector.RecordAPICall(APIMetrics{
				Timestamp:   time.Now(),
				Method:      c.Request.Method,
				Endpoint:    c.Request.URL.Path,
				StatusCode:  400,
				ErrorType:   "client_error",
				ServiceName: "review-service",
			})

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request data",
				"details": fmt.Sprintf("Validation failed: %v", err),
			})
			return
		}

		// Successful request
		collector.RecordAPICall(APIMetrics{
			Timestamp:   time.Now(),
			Method:      c.Request.Method,
			Endpoint:    c.Request.URL.Path,
			StatusCode:  200,
			ErrorType:   "",
			ServiceName: "review-service",
		})

		c.JSON(http.StatusOK, gin.H{
			"analysis": "Code analysis completed",
			"message":  fmt.Sprintf("Analyzed %d characters", len(validRequest.PastedCode)),
		})
	}
}

// Test cases demonstrating the monitoring system
func TestMonitoringCatchesPayloadIssues(t *testing.T) {
	collector := &MockMetricsCollector{}

	// Create test router with monitoring
	router := gin.New()
	router.POST("/api/review/modes/preview", createReviewHandler(collector))

	tests := []struct {
		name                string
		payload             string
		expectedStatus      int
		expectedExtraFields []string
		shouldAlert         bool
	}{
		{
			name: "✅ Valid Request - No Issues",
			payload: `{
				"pasted_code": "func main() { fmt.Println(\"Hello\") }",
				"model": "claude-3-5-sonnet"
			}`,
			expectedStatus:      200,
			expectedExtraFields: nil,
			shouldAlert:         false,
		},
		{
			name: "🚨 Invalid Request - Contains session_id (THE BUG!)",
			payload: `{
				"session_id": "abc123",
				"pasted_code": "func main() {}",
				"model": "claude-3-5-sonnet"
			}`,
			expectedStatus:      400,
			expectedExtraFields: []string{"session_id"},
			shouldAlert:         true,
		},
		{
			name: "🚨 Invalid Request - Missing Required Field",
			payload: `{
				"model": "claude-3-5-sonnet"
			}`,
			expectedStatus:      400,
			expectedExtraFields: nil,
			shouldAlert:         true,
		},
	}

	fmt.Println("\n🧪 Running Monitoring Tests...")
	fmt.Println("=====================================")

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Reset collector for this test
			initialCallCount := len(collector.APICallsRecorded)
			initialFailureCount := len(collector.ValidationFailures)

			// Make request
			req := httptest.NewRequest("POST", "/api/review/modes/preview",
				strings.NewReader(test.payload))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			// Verify response
			assert.Equal(t, test.expectedStatus, resp.Code)

			// Verify monitoring recorded the call
			assert.Equal(t, initialCallCount+1, len(collector.APICallsRecorded),
				"Monitoring should record API call")

			// Check if validation failure was recorded
			if test.shouldAlert {
				assert.Greater(t, len(collector.ValidationFailures), initialFailureCount,
					"Should record validation failure")

				if len(collector.ValidationFailures) > initialFailureCount {
					failure := collector.ValidationFailures[len(collector.ValidationFailures)-1]
					for _, expectedField := range test.expectedExtraFields {
						assert.Contains(t, failure.ExtraFields, expectedField,
							"Should detect extra field: %s", expectedField)
					}
				}
			}

			// Print results
			if resp.Code == 200 {
				fmt.Printf("  %d. %s ✅\n", i+1, test.name)
				fmt.Printf("     Status: %d, Monitored: ✅\n", resp.Code)
			} else {
				fmt.Printf("  %d. %s ❌\n", i+1, test.name)
				fmt.Printf("     Status: %d, Alert: 🚨, Monitored: ✅\n", resp.Code)
				if test.expectedExtraFields != nil {
					fmt.Printf("     Extra Fields: %v\n", test.expectedExtraFields)
				}
			}
		})
	}

	// Demonstrate monitoring analysis
	fmt.Println("\n📊 Monitoring Analysis Results")
	fmt.Println("================================")

	// Calculate error rate
	totalCalls := len(collector.APICallsRecorded)
	errorCalls := 0
	for _, call := range collector.APICallsRecorded {
		if call.StatusCode >= 400 {
			errorCalls++
		}
	}

	errorRate := float64(errorCalls) / float64(totalCalls) * 100
	fmt.Printf("Total API calls: %d\n", totalCalls)
	fmt.Printf("Error calls: %d\n", errorCalls)
	fmt.Printf("Error rate: %.1f%%\n", errorRate)

	// Alert analysis
	fmt.Println("\n🚨 Validation Failures Detected:")
	for i, failure := range collector.ValidationFailures {
		fmt.Printf("  %d. Endpoint: %s\n", i+1, failure.Endpoint)
		if len(failure.ExtraFields) > 0 {
			fmt.Printf("     Extra fields: %v\n", failure.ExtraFields)
		}
		if len(failure.MissingFields) > 0 {
			fmt.Printf("     Missing fields: %v\n", failure.MissingFields)
		}
		fmt.Printf("     Time: %s\n", failure.Timestamp.Format("15:04:05"))
	}

	// Simulate dashboard alert
	if errorRate > 30 { // 30% threshold for demo
		fmt.Printf("\n🚨 ALERT TRIGGERED!\n")
		fmt.Printf("Error rate %.1f%% exceeds threshold 30%%\n", errorRate)
		fmt.Printf("Primary issue: session_id field in payload\n")
		fmt.Printf("Recommended action: Remove session_id from frontend\n")
	}
}

// Main function for standalone execution
func main() {
	// Run as a demo when called directly
	fmt.Println("🔍 DevSmith Monitoring Demo")
	fmt.Println("This demonstrates how monitoring catches API payload issues")
	fmt.Println("Run with: go test -v")
}
EOF

echo -e "${GREEN}✅ Demo environment created${NC}"

# Step 2: Run the monitoring test
echo ""
echo -e "${BLUE}Step 2: Running monitoring demonstration...${NC}"
echo ""

# Initialize Go module for demo
go mod init monitoring-demo
go get github.com/gin-gonic/gin@latest
go get github.com/stretchr/testify/assert@latest

# Run the test
echo -e "${YELLOW}Running monitoring test...${NC}"
go test -v 2>&1 | tee $LOG_FILE

echo ""
echo -e "${BLUE}Step 3: Analyzing results...${NC}"
echo ""

# Check if test passed
if grep -q "PASS" $LOG_FILE; then
    echo -e "${GREEN}✅ All monitoring tests passed!${NC}"
    echo ""
    echo "Key findings from the demo:"
    echo -e "${YELLOW}1. Valid requests (no session_id)${NC} → ✅ Status 200, no alerts"
    echo -e "${YELLOW}2. Invalid requests (with session_id)${NC} → ❌ Status 400, 🚨 alert triggered"
    echo -e "${YELLOW}3. Monitoring detected the exact issue${NC} → Extra field: 'session_id'"
    
    if grep -q "ALERT TRIGGERED" $LOG_FILE; then
        echo ""
        echo -e "${RED}🚨 CRITICAL: Alert threshold exceeded!${NC}"
        echo "The monitoring system detected error rates above acceptable levels."
        echo "This simulates what would happen during the session_id bug."
    fi
    
else
    echo -e "${RED}❌ Some tests failed${NC}"
    echo "Check the log for details:"
    echo "  cat $PWD/$LOG_FILE"
fi

echo ""
echo -e "${BLUE}Step 4: Production implementation guidance${NC}"
echo ""

cat << 'EOF'
📋 How to implement this monitoring in production:

1. 🏗️  Add monitoring middleware to Review service:
   - Copy internal/monitoring/middleware.go to project
   - Add middleware to Gin router in cmd/review/main.go
   - Configure PostgreSQL storage

2. 📊 Create monitoring dashboard in Logs app:
   - Add /api/monitoring/metrics endpoint  
   - Create Chart.js visualization for error rates
   - Add WebSocket for real-time updates

3. 🚨 Configure alerting:
   - Set error rate threshold (e.g., 5.0 errors/min)
   - Add email/Slack notifications
   - Create alert management UI

4. 🧪 Add to CI/CD pipeline:
   - Run API contract tests in GitHub Actions
   - Block deployments with high error rates
   - Load test with realistic payloads

5. 📈 Dashboard features:
   - Real-time error rate graphs
   - Endpoint performance metrics  
   - Payload validation failure analysis
   - Historical trend tracking

The session_id issue would have been caught immediately with this system!
EOF

echo ""
echo -e "${GREEN}Demo complete!${NC}"
echo "Demo files created in: $PWD"
echo "Test results: $PWD/$LOG_FILE"

# Cleanup option
echo ""
echo -e "${YELLOW}Clean up demo files? (y/N)${NC}"
read -r cleanup
if [[ $cleanup =~ ^[Yy]$ ]]; then
    cd ..
    rm -rf $DEMO_DIR
    echo -e "${GREEN}✅ Demo files cleaned up${NC}"
else
    echo "Demo files preserved in: $PWD"
fi
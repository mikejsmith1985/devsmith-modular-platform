# Monitoring Integration Example - Review Service

This example shows how to integrate the monitoring system with the Review service to catch payload validation issues like the session_id problem.

## 1. Integration Steps

### Step 1: Add Monitoring to Review Service

```go
// cmd/review/main.go

package main

import (
    "context"
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
    
    "github.com/mikejsmith1985/devsmith-modular-platform/internal/monitoring"
    // ... other imports
)

func main() {
    // Initialize database connection
    dbPool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer dbPool.Close()

    // Initialize monitoring
    metricsCollector := monitoring.NewPostgreSQLMetricsCollector(dbPool)
    
    // Initialize monitoring schema
    if err := metricsCollector.InitializeSchema(context.Background()); err != nil {
        log.Fatal("Failed to initialize monitoring schema:", err)
    }

    // Create Gin router
    router := gin.New()
    
    // Add monitoring middleware - THIS CATCHES ALL API CALLS
    router.Use(monitoring.MetricsMiddleware(metricsCollector, "review-service"))
    
    // Add other middleware (logging, CORS, etc.)
    
    // Register routes
    setupRoutes(router)
    
    // Start server
    router.Run(":8081")
}
```

### Step 2: Enhanced Payload Validation with Monitoring

```go
// internal/review/handlers/modes_handler.go

package handlers

import (
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "github.com/mikejsmith1985/devsmith-modular-platform/internal/monitoring"
)

type CodeRequest struct {
    PastedCode string `json:"pasted_code" binding:"required"`
    Model      string `json:"model" binding:"required"`
    // NOTE: No session_id field - that was the bug!
}

func (h *ModesHandler) HandlePreview(c *gin.Context) {
    var req CodeRequest
    
    // Custom validation that detects extra fields
    if err := c.ShouldBindJSON(&req); err != nil {
        // Check if request contains session_id field (the problematic field)
        var rawPayload map[string]interface{}
        if bindErr := c.ShouldBindJSON(&rawPayload); bindErr == nil {
            var extraFields []string
            var missingFields []string
            
            // Check for extra fields not in our struct
            for key := range rawPayload {
                if key != "pasted_code" && key != "model" {
                    extraFields = append(extraFields, key)
                }
            }
            
            // Check for missing required fields
            if rawPayload["pasted_code"] == nil {
                missingFields = append(missingFields, "pasted_code")
            }
            if rawPayload["model"] == nil {
                missingFields = append(missingFields, "model")
            }
            
            // Record detailed validation failure - MONITORING CATCHES THIS
            monitoring.RecordPayloadValidationFailure(c, []string{}, extraFields, missingFields)
        }
        
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request data",
            "details": err.Error(),
        })
        return
    }
    
    // Process the valid request
    result, err := h.service.ProcessPreview(c.Request.Context(), req.PastedCode, req.Model)
    if err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "error": "Service temporarily unavailable",
        })
        return
    }
    
    c.JSON(http.StatusOK, result)
}
```

### Step 3: Monitoring Dashboard Endpoint

```go
// cmd/logs/main.go - Add monitoring dashboard to Logs service

func setupMonitoringRoutes(router *gin.Engine, metricsCollector monitoring.MetricsCollector) {
    api := router.Group("/api/monitoring")
    
    // Real-time metrics endpoint
    api.GET("/metrics", func(c *gin.Context) {
        window := 15 * time.Minute // 15 minute window
        
        errorRate, _ := metricsCollector.GetErrorRate(c.Request.Context(), window)
        responseTimes, _ := metricsCollector.GetResponseTimes(c.Request.Context(), window)
        
        c.JSON(http.StatusOK, gin.H{
            "error_rate_per_minute": errorRate,
            "avg_response_time": calculateAverage(responseTimes),
            "p95_response_time": calculatePercentile(responseTimes, 0.95),
            "window_minutes": window.Minutes(),
        })
    })
    
    // Endpoint analysis
    api.GET("/endpoints", func(c *gin.Context) {
        window := 1 * time.Hour
        
        endpoints, _ := metricsCollector.GetEndpointMetrics(c.Request.Context(), window)
        
        c.JSON(http.StatusOK, gin.H{
            "endpoints": endpoints,
            "window_hours": window.Hours(),
        })
    })
}
```

## 2. How This Catches the session_id Issue

### Before Monitoring (What Happened)
```bash
# Frontend sends:
POST /api/review/modes/preview
{
  "session_id": "abc123",    # ← This field broke everything
  "pasted_code": "func main() {}",
  "model": "claude-3-5-sonnet"
}

# Backend response:
HTTP 400 - "Invalid request data"

# Result: 
# - No visibility into WHY it failed
# - No tracking of error frequency  
# - No alerting when error rate spikes
# - Manual testing required to discover
```

### With Monitoring (What Would Happen)
```bash
# Same request sent...

# Monitoring Middleware Records:
{
  "timestamp": "2025-01-15T10:30:00Z",
  "method": "POST", 
  "endpoint": "/api/review/modes/preview",
  "status_code": 400,
  "error_type": "client_error", 
  "error_message": "Bad request - likely payload validation failure",
  "service_name": "review-service"
}

# Validation Failure Detail Log:
{
  "endpoint": "/api/review/modes/preview",
  "extra_fields": ["session_id"],           # ← THE SMOKING GUN
  "missing_fields": [],
  "timestamp": "2025-01-15T10:30:00Z"
}

# Alert Triggered (if error rate > 5/minute):
🚨 ALERT: Review API error rate: 15.2/min (threshold: 5.0/min)
   Endpoint: /api/review/modes/preview  
   Error Type: payload validation failure
   Extra Fields: session_id
   Action: Investigate frontend payload structure
```

### Dashboard View

```
┌─────────────────────────────────────────────────────────────┐
│ API Error Rate (Last 15 Minutes)                          │
│                                                             │
│ 20 │     ████                                               │
│ 15 │   ██████                                               │
│ 10 │ ████████                                               │
│  5 │████████████ ← Alert Threshold                          │
│  0 │████████████████████████████████████████████████       │
│    └─────────────────────────────────────────────────────   │
│     10:15   10:20   10:25   10:30   10:35   10:40         │
│                                                             │
│ 🚨 SPIKE DETECTED at 10:30 - session_id field in payload   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Top Failing Endpoints                                      │
│                                                             │
│ /api/review/modes/preview    │ 89 errors │ session_id     │
│ /api/review/modes/skim       │ 12 errors │ session_id     │
│ /api/review/modes/detailed   │ 45 errors │ session_id     │
│                                                             │
│ 💡 Suggested Fix: Remove session_id from frontend payload  │
└─────────────────────────────────────────────────────────────┘
```

## 3. Benefits for Development Process

### Immediate Detection
- **Real-time alerts** when error rates spike
- **Automatic categorization** of payload validation failures
- **Field-level analysis** shows exactly what's wrong

### Historical Analysis
- **Track error trends** over time
- **Identify recurring issues** across deployments
- **Measure improvement** after fixes are deployed

### Proactive Prevention
- **Pre-commit hooks** that run API tests (using our test suite)
- **CI/CD integration** that fails builds on high error rates
- **Load testing integration** that simulates real payloads

### Better Debugging
- **Structured logs** with full context
- **Correlation IDs** to trace requests across services
- **Performance impact** analysis (response times)

## 4. Implementation Timeline

**Week 1: Foundation**
- ✅ Monitoring middleware and storage (COMPLETED)
- ⏳ Integrate with Review service
- ⏳ Basic dashboard in Logs app

**Week 2: Intelligence**  
- ⏳ Alert engine and thresholds
- ⏳ Payload validation failure analysis
- ⏳ Real-time dashboard features

**Week 3: Integration**
- ⏳ CI/CD pipeline integration  
- ⏳ Pre-commit hook API validation
- ⏳ Load testing with monitoring

**Week 4: Production**
- ⏳ Deploy to production with monitoring
- ⏳ Fine-tune alert thresholds
- ⏳ Train team on dashboard usage

## 5. Success Metrics

### Prevention Metrics
- **0 payload structure issues** reach production (like session_id)
- **< 5 minute detection time** for API issues  
- **90% alert accuracy** (alerts indicate real issues)

### Development Velocity
- **50% reduction** in manual testing time
- **75% reduction** in debugging time for API issues
- **24/7 monitoring** without human intervention

This monitoring system would have caught the session_id issue immediately and provided clear guidance on how to fix it, instead of requiring manual testing and investigation.
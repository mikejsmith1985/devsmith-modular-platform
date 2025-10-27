# Issue #024: Service Logging Configuration & Inter-Service Communication

**Priority:** High (Blocking)  
**Type:** Infrastructure/Configuration  
**Complexity:** Medium  
**Estimated Effort:** 1-2 days  

---

## Summary

Establish sustainable, scalable infrastructure for inter-service logging communication. Currently, services cannot reach the centralized Logs service due to missing environment variable configuration. This issue creates a production-ready configuration system that works across all deployment environments (local, Docker, staging, production).

**Why This Matters:**
- Issue #21 (validation logging) cannot reach logs service
- No visibility into what happens across services
- Configuration is environment-specific and fragile
- Need sustainable solution for all services to report logs
- Future issues will depend on this being robust

---

## Acceptance Criteria

### Core Features (All Required)
- [ ] **Configuration Management System**
  - Centralized service URL configuration
  - Environment-specific overrides
  - Sensible defaults for all environments
  - Configuration documented and validated

- [ ] **Docker Compose Configuration**
  - All services have LOGS_SERVICE_URL set correctly
  - Format: `http://<service-name>:<port>/api/logs`
  - Works with internal Docker networking
  - Health checks still pass

- [ ] **Local Development Setup**
  - Clear documentation for running locally
  - Works with `docker-compose up`
  - Works with local development server
  - Easy to debug connectivity issues

- [ ] **Environment Variables**
  - .env.example updated with LOGS_SERVICE_URL
  - All services documented
  - Clear naming convention (SERVICE_LOGS_URL pattern)
  - Validation on startup

- [ ] **Service Configuration**
  - Review service: LOGS_SERVICE_URL configured
  - Portal service: LOGS_SERVICE_URL configured
  - Analytics service: LOGS_SERVICE_URL configured (future logging)
  - Logs service: Self-reference handling

- [ ] **Startup Validation**
  - Services validate logging config on startup
  - Clear error messages if config invalid
  - Graceful fallback for missing config
  - Log startup configuration for debugging

- [ ] **Documentation**
  - ARCHITECTURE.md updated with logging infrastructure
  - Service-to-service communication documented
  - Deployment guide for different environments
  - Troubleshooting guide

### Testing
- [ ] Verify services can reach logs service via docker-compose
- [ ] Verify validation failures create logs
- [ ] Manual: Create invalid input, confirm logs appear
- [ ] Manual: Check logs dashboard for validation errors

---

## Implementation Plan

### Phase 1: Configuration Infrastructure

**Create:** `.env.example` additions
```bash
# Logging Service Configuration
# Format: http://<service-name>:<port>/api/logs
# For local development, use http://localhost:8082 or http://logs:8082 (docker)

LOGS_SERVICE_URL=http://logs:8082/api/logs              # Docker Compose
# LOGS_SERVICE_URL=http://localhost:8082/api/logs       # Local dev
# LOGS_SERVICE_URL=https://logs.example.com/api/logs    # Production
```

**Update:** `docker-compose.yml`
- Add LOGS_SERVICE_URL to all services
- Format: `http://logs:8082/api/logs` (uses Docker internal DNS)
- Services: review, portal, analytics

**Create:** Environment configuration package
- `internal/config/logging.go`
- Reads LOGS_SERVICE_URL from environment
- Validates URL format
- Provides sensible defaults

### Phase 2: Service Integration

**Update:** Review Service
- `cmd/review/main.go` - Load and validate logging config
- `cmd/review/handlers/validation_helper.go` - Use config
- `cmd/review/Dockerfile` - Document environment variables

**Update:** Portal Service (for future logging)
- Same pattern as review service

**Update:** Analytics Service (for future logging)
- Same pattern as review service

### Phase 3: Documentation & Validation

**Update:** `ARCHITECTURE.md`
- Add section on cross-service logging
- Explain service-to-service communication pattern
- Document environment variable naming convention

**Create:** `DEPLOYMENT.md`
- Local development setup
- Docker Compose setup
- Production deployment
- Troubleshooting connectivity

**Create:** `docs/LOGGING.md`
- How services log to centralized logs service
- Configuration options
- Examples for each service

---

## Technical Details

### Configuration Pattern (Reusable for Other Services)

**File:** `internal/config/logging.go`

```go
// LoadLogsConfig loads and validates logging service configuration
func LoadLogsConfig() (string, error) {
    url := os.Getenv("LOGS_SERVICE_URL")
    
    if url == "" {
        // Default based on environment
        if os.Getenv("ENVIRONMENT") == "docker" {
            url = "http://logs:8082/api/logs"
        } else {
            url = "http://localhost:8082/api/logs"
        }
    }
    
    // Validate URL format
    if err := validateLogsURL(url); err != nil {
        return "", fmt.Errorf("invalid LOGS_SERVICE_URL: %w", err)
    }
    
    return url, nil
}

// validateLogsURL ensures the URL is properly formatted
func validateLogsURL(url string) error {
    parsed, err := url.Parse(url)
    if err != nil {
        return fmt.Errorf("invalid URL format: %w", err)
    }
    
    if parsed.Scheme != "http" && parsed.Scheme != "https" {
        return fmt.Errorf("invalid scheme: must be http or https")
    }
    
    if parsed.Path != "/api/logs" {
        return fmt.Errorf("invalid path: must be /api/logs")
    }
    
    return nil
}
```

### Service Integration Pattern

**In each service's main.go:**

```go
// Load and validate logging configuration
logsURL, err := config.LoadLogsConfig()
if err != nil {
    log.Fatalf("Failed to load logging configuration: %v", err)
}
log.Printf("Logs service configured: %s", logsURL)

// Pass to handlers/services that need it
handler := handlers.NewReviewHandler(
    reviewService,
    scanService,
    logsURL,  // ← Pass config
)
```

**In handlers that log:**

```go
// Store logsURL as dependency
type ReviewHandler struct {
    reviewService ReviewService
    scanService   ScanService
    logsURL       string
}

// Use in logging function
func (h *ReviewHandler) logValidationFailure(errorType, message string, c *gin.Context) {
    logEntry := map[string]interface{}{
        "service": "review",
        "level":   "warning",
        "message": message,
        // ... metadata ...
    }
    
    jsonData, _ := json.Marshal(logEntry)
    
    // Use configured URL
    http.Post(
        h.logsURL,  // ← From configuration
        "application/json",
        bytes.NewReader(jsonData),
    )
}
```

---

## Docker Compose Changes

### Before
```yaml
review:
  build:
    context: .
    dockerfile: cmd/review/Dockerfile
  ports:
    - "8081:8081"
  environment:
    - PORT=8081
    - DATABASE_URL=postgres://devsmith:devsmith@postgres:5432/devsmith?sslmode=disable
    - REVIEW_DB_URL=postgres://devsmith:devsmith@postgres:5432/devsmith?sslmode=disable
```

### After
```yaml
review:
  build:
    context: .
    dockerfile: cmd/review/Dockerfile
  ports:
    - "8081:8081"
  environment:
    - PORT=8081
    - DATABASE_URL=postgres://devsmith:devsmith@postgres:5432/devsmith?sslmode=disable
    - REVIEW_DB_URL=postgres://devsmith:devsmith@postgres:5432/devsmith?sslmode=disable
    - LOGS_SERVICE_URL=http://logs:8082/api/logs
    - ENVIRONMENT=docker
  depends_on:
    logs:
      condition: service_healthy
```

---

## Environment Variables Convention

**Naming Pattern:** `<SERVICE>_LOGS_URL` or global `LOGS_SERVICE_URL`

**Recommendation:** Use global `LOGS_SERVICE_URL` for simplicity (single source of truth)

**Format Options:**
- Docker: `http://logs:8082/api/logs` (internal DNS)
- Local: `http://localhost:8082/api/logs` (localhost)
- Cloud: `https://logs.example.com/api/logs` (DNS name)

---

## Testing Strategy

### Unit Tests
```go
func TestLoadLogsConfig_EnvVar(t *testing.T) {
    os.Setenv("LOGS_SERVICE_URL", "http://test:8082/api/logs")
    url, err := LoadLogsConfig()
    assert.NoError(t, err)
    assert.Equal(t, "http://test:8082/api/logs", url)
}

func TestLoadLogsConfig_InvalidURL(t *testing.T) {
    os.Setenv("LOGS_SERVICE_URL", "ftp://invalid/path")
    _, err := LoadLogsConfig()
    assert.Error(t, err)
}

func TestLoadLogsConfig_DefaultDocker(t *testing.T) {
    os.Unsetenv("LOGS_SERVICE_URL")
    os.Setenv("ENVIRONMENT", "docker")
    url, err := LoadLogsConfig()
    assert.NoError(t, err)
    assert.Contains(t, url, "logs:8082")
}
```

### Integration Tests
```bash
# 1. Start docker compose
docker-compose up -d

# 2. Test connectivity from review service
curl http://localhost:8081/health  # Should be healthy (logs service reachable)

# 3. Create validation error and check logs
curl -X POST http://localhost:3000/api/review/sessions \
  -H "Content-Type: application/json" \
  -d '{"title":"","code_source":"invalid"}'

# 4. Verify logs appear in logs service
curl http://localhost:8082/api/logs?service=review&level=warning
```

---

## Scalability Considerations

This pattern is designed to be reusable for:

1. **Other Services** - Portal, Analytics, Build services can all follow same pattern
2. **Multiple Log Destinations** - Easy to add alerting, analytics backends
3. **Environment Management** - Supports local, Docker, staging, production
4. **Testing** - Mock logs URL in tests
5. **Feature Flags** - Can conditionally enable logging per service

---

## Success Criteria

When this issue is complete:

1. ✅ All services have LOGS_SERVICE_URL configured
2. ✅ Review service validation failures appear in logs
3. ✅ docker-compose up works without manual config
4. ✅ Documentation explains the pattern
5. ✅ Configuration is validated on startup
6. ✅ Clear error messages if logging unavailable
7. ✅ Sustainable for adding more services
8. ✅ Works in all environments (local, docker, prod)

---

## References

- **Issue #21:** Input Validation & Sanitization (depends on this)
- **Issue #023:** Logs Service Production Enhancements (depends on this)
- **ARCHITECTURE.md:** Service architecture and communication
- **docker-compose.yml:** Current deployment configuration

---

## Notes

- **Blocking:** This issue blocks Issue #023 dashboard implementation
- **Priority:** Complete BEFORE Issue #023 GREEN phase
- **Reusable:** Pattern can be used for other cross-service configs
- **Testing:** Manual validation required (create errors and verify logs appear)
- **Documentation:** Critical for onboarding new services

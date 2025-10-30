# Cross-Service Logging (DevSmith)

This document describes the logging configuration used by DevSmith services and how to configure per-service overrides and startup behavior.

Environment variables
- `LOGS_SERVICE_URL` — global URL for the Logs service (default: `http://localhost:8082/api/logs` locally, `http://logs:8082/api/logs` in Docker when `ENVIRONMENT=docker`).
- `SERVICE_LOGS_URL` — per-service override (pattern: `<SERVICE>_LOGS_URL`, e.g. `REVIEW_LOGS_URL`, `PORTAL_LOGS_URL`, `ANALYTICS_LOGS_URL`). If present it takes precedence over `LOGS_SERVICE_URL`.
- `LOGS_STRICT` — controls startup behavior when the logs URL is invalid. Defaults to `true`.
  - `true` (default): the service will validate the URL at startup and fail fast on invalid configuration.
  - `false`: the service will attempt to validate; if invalid, logging is disabled and the service continues startup.

Behavior
- Services call `internal/config.LoadLogsConfigFor(service)` to resolve the effective logs URL.
- To allow graceful startup when the Logs service may be temporarily unavailable (e.g. during maintenance), set `LOGS_STRICT=false`. The service will then continue without external logging.

Examples
- Docker Compose snippet (service `review`):

```yaml
environment:
  - PORT=8081
  - REVIEW_DB_URL=postgres://devsmith:devsmith@postgres:5432/devsmith?sslmode=disable
  - LOGS_SERVICE_URL=http://logs:8082/api/logs
  - ENVIRONMENT=docker
```

Troubleshooting
- If you see startup errors like "invalid logs url", verify `LOGS_SERVICE_URL` or set `LOGS_STRICT=false` temporarily.
- For per-service overrides, check e.g. `REVIEW_LOGS_URL` for the `review` service.

API
- `internal/config.LoadLogsConfigFor(service string) (string, error)` — resolves and validates URL.
- `internal/config.LoadLogsConfigWithFallbackFor(service string) (string, bool, error)` — resolves, validates, and returns `(url, enabled, err)` where `enabled=false` indicates logging disabled due to fallback.

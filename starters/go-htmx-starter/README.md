# Go + HTMX Starter Template

A minimal, opinionated starter template for building server-side rendered web applications with Go and HTMX. This template enforces the DevSmith platform contract and provides a solid foundation for building hypermedia-driven applications.

## Platform Contract

This starter enforces the following platform requirements:

### 1. Health Endpoint
- **Endpoint**: `GET /health`
- **Response**: JSON with service name and status
- **Purpose**: Docker health checks and monitoring
- **Example**:
  ```json
  {
    "service": "my-service",
    "status": "healthy",
    "timestamp": "2024-11-03T12:00:00Z"
  }
  ```

### 2. Structured Logging
- JSON-formatted logs for machine parsing
- Correlation IDs for distributed tracing
- Log levels: DEBUG, INFO, WARN, ERROR
- Automatic request/response logging with correlation

### 3. Correlation Headers
- **Header**: `X-Correlation-ID`
- Automatically extracted from incoming requests
- Generated if not present
- Propagated to all log entries
- Included in all outgoing HTTP responses

### 4. Container Packaging
- Multi-stage Docker build for minimal image size
- Non-root user for security
- Health check configured in Dockerfile
- Alpine-based for small footprint

### 5. Local Development
- docker-compose.yml for easy local setup
- Hot reload support (mount source as volume)
- Consistent port mapping

## Quick Start

### Prerequisites
- Go 1.22 or later
- Docker and Docker Compose (for containerized development)

### Local Development (Native)

```bash
# Install dependencies
go mod download

# Run the server
go run main.go logger.go

# Application runs at http://localhost:8080
```

### Local Development (Docker)

```bash
# Build and run with docker-compose
docker-compose up --build

# Application runs at http://localhost:8080

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

## Project Structure

```
.
├── main.go              # Application entry point, routes, handlers
├── logger.go            # Structured logging with correlation IDs
├── templates/           # HTML templates
│   ├── base.html       # Base layout with HTMX
│   ├── index.html      # Home page
│   └── fragment.html   # HTMX partial response
├── Dockerfile          # Multi-stage container build
├── docker-compose.yml  # Local development setup
├── app.yaml            # Application manifest
└── README.md           # This file
```

## Features

### Server-Side Rendering
- Go's `html/template` for safe, efficient rendering
- Layout inheritance with `base.html`
- Partial templates for HTMX responses

### HTMX Integration
- Installed via CDN in base template
- Example HTMX interaction (click to load fragment)
- Server returns HTML fragments, not JSON

### Correlation Tracking
- Middleware extracts or generates correlation ID
- All logs tagged with correlation ID
- Easy request tracing across services

### Health Monitoring
- `/health` endpoint for orchestration
- Docker HEALTHCHECK directive
- Returns JSON status

## Usage Examples

### Adding a New Route

```go
// In main.go, add to setupRoutes()
mux.HandleFunc("/my-route", func(w http.ResponseWriter, r *http.Request) {
    logger.Info(r.Context(), "my_route_accessed", map[string]interface{}{
        "user_agent": r.UserAgent(),
    })
    
    tmpl := template.Must(template.ParseFiles(
        "templates/base.html",
        "templates/my-page.html",
    ))
    tmpl.ExecuteTemplate(w, "base", nil)
})
```

### Adding an HTMX Endpoint

```go
// Return HTML fragment (no layout)
mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
    data := struct {
        Message string
        Count   int
    }{
        Message: "Hello from HTMX",
        Count:   42,
    }
    
    tmpl := template.Must(template.ParseFiles("templates/data-fragment.html"))
    tmpl.Execute(w, data)
})
```

### Logging Best Practices

```go
// Structured logging with context
logger.Info(ctx, "user_action", map[string]interface{}{
    "action": "login",
    "user_id": userID,
})

// Error logging
logger.Error(ctx, "database_error", map[string]interface{}{
    "error": err.Error(),
    "query": "SELECT ...",
})
```

## Testing

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# Build verification
go build -o bin/app main.go logger.go
```

## Deployment

### Building for Production

```bash
# Build Docker image
docker build -t my-app:latest .

# Run container
docker run -p 8080:8080 my-app:latest
```

### Environment Variables

- `PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Logging level: DEBUG, INFO, WARN, ERROR (default: INFO)
- `SERVICE_NAME` - Service identifier for logs (default: go-htmx-starter)

## CI/CD

The included `.github/workflows/ci.yml` workflow:
- Runs on push and pull requests
- Builds the Go module
- Verifies no compilation errors
- Can be extended with tests and linting

## Customization Guide

### Renaming the Service

1. Update `SERVICE_NAME` in docker-compose.yml
2. Update service name in main.go health endpoint
3. Update app.yaml metadata

### Adding a Database

1. Add database service to docker-compose.yml
2. Install database driver: `go get github.com/lib/pq`
3. Initialize connection in main.go
4. Update health check to verify database connectivity

### Adding Static Assets

1. Create `static/` directory
2. Add file server in main.go:
   ```go
   fs := http.FileServer(http.Dir("static"))
   mux.Handle("/static/", http.StripPrefix("/static/", fs))
   ```
3. Update Dockerfile to copy static files
4. Reference in templates: `<link rel="stylesheet" href="/static/style.css">`

## HTMX Patterns

### Swap Content on Click

```html
<button hx-get="/api/users" hx-target="#user-list" hx-swap="innerHTML">
    Load Users
</button>
<div id="user-list"></div>
```

### Form Submission

```html
<form hx-post="/api/users" hx-target="#result">
    <input type="text" name="username" />
    <button type="submit">Create User</button>
</form>
<div id="result"></div>
```

### Polling for Updates

```html
<div hx-get="/api/status" hx-trigger="every 2s" hx-swap="innerHTML">
    Checking status...
</div>
```

## Troubleshooting

### Container Won't Start
- Check logs: `docker-compose logs`
- Verify port 8080 is not in use: `lsof -i :8080`
- Check health endpoint: `curl http://localhost:8080/health`

### Templates Not Loading
- Ensure templates/ directory exists
- Check working directory in container
- Verify COPY directive in Dockerfile

### HTMX Not Working
- Check browser console for errors
- Verify HTMX CDN is accessible
- Check network tab for HTMX requests

## Resources

- [HTMX Documentation](https://htmx.org/docs/)
- [Go Templates](https://pkg.go.dev/html/template)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Structured Logging](https://www.structlog.org/en/stable/)

## License

Use this starter template freely for your projects. No attribution required.

## Support

For issues or questions about this starter:
1. Check the troubleshooting section
2. Review the platform documentation
3. Open an issue in the main repository

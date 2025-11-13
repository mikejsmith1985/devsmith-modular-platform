# Stage 1: Build Go binary
FROM golang:1.24-alpine AS builder

# Build arguments for version injection
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build with version information injected via ldflags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s \
    -X github.com/mikejsmith1985/devsmith-modular-platform/internal/version.Version=${VERSION} \
    -X github.com/mikejsmith1985/devsmith-modular-platform/internal/version.CommitHash=${GIT_COMMIT} \
    -X github.com/mikejsmith1985/devsmith-modular-platform/internal/version.BuildTime=${BUILD_TIME}" \
    -o /app/bin/portal ./cmd/portal

# Stage 3: Final runtime image
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wget curl
RUN addgroup -g 1000 appuser && \
	adduser -D -u 1000 -G appuser appuser
WORKDIR /home/appuser
COPY --from=builder /app/bin/portal ./portal
RUN chown -R appuser:appuser /home/appuser
USER appuser
COPY apps/portal/templates/ ./templates/
COPY --chown=appuser:appuser apps/portal/static/ ./static/
# Inject build timestamp for cache-busting (prevents stale frontend builds)
ARG BUILD_TIMESTAMP
RUN if [ -n "$BUILD_TIMESTAMP" ]; then \
      sed -i "s/BUILD_TIMESTAMP_PLACEHOLDER/${BUILD_TIMESTAMP}/" ./static/index.html; \
    else \
      sed -i "s/BUILD_TIMESTAMP_PLACEHOLDER/$(date +%s)/" ./static/index.html; \
    fi
EXPOSE 3001
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1
CMD ["./portal"]
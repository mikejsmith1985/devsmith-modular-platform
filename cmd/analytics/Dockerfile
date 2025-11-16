FROM golang:1.24-alpine AS builder

# Cache-busting argument - forces fresh build when source changes
# Pass with: --build-arg BUILD_TIMESTAMP=$(date +%s)
ARG BUILD_TIMESTAMP=0
RUN echo "Build timestamp: ${BUILD_TIMESTAMP}"

RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/analytics ./cmd/analytics

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wget curl
RUN addgroup -g 1000 appuser && \
	adduser -D -u 1000 -G appuser appuser
WORKDIR /home/appuser
COPY --from=builder /app/bin/analytics ./analytics
RUN chown -R appuser:appuser /home/appuser
USER appuser
EXPOSE 8083
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget --no-verbose --tries=1 --spider http://localhost:8083/api/analytics/health || exit 1
CMD ["./analytics"]
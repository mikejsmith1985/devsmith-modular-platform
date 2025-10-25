FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/review ./cmd/review

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wget curl netcat-openbsd
RUN addgroup -g 1000 appuser && \
	adduser -D -u 1000 -G appuser appuser
WORKDIR /home/appuser
COPY --from=builder /app/bin/review ./review
COPY scripts/wait-for-postgres.sh /wait-for-postgres.sh
COPY scripts/start-review.sh /start-review.sh
RUN chmod +x /wait-for-postgres.sh /start-review.sh && chown appuser:appuser /wait-for-postgres.sh /start-review.sh
RUN chown -R appuser:appuser /home/appuser
USER appuser
EXPOSE 8081
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1
CMD ["/start-review.sh"]
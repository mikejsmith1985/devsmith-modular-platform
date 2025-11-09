# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/portal ./cmd/portal

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
COPY apps/portal/static/ ./static/
COPY --from=frontend-builder /frontend/dist ./dist/
EXPOSE 3001
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1
CMD ["./portal"]
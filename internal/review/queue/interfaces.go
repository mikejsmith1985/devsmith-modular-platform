// Package queue provides request queuing for AI API calls.
package queue

import (
	"context"
	"errors"
	"time"
)

// Errors for queue operations.
var (
	ErrQueueFull = errors.New("queue is full")
)

// AIRequest represents a queued AI request.
type AIRequest struct {
	ID       string
	Mode     string
	Content  string
	UserID   int64
	MaxRetry int
}

// AIResponse represents an AI response.
type AIResponse struct {
	Result    interface{}
	RequestID string
	Duration  time.Duration
}

// RequestStatus represents current request status.
type RequestStatus struct {
	RequestID string
	State     string // queued, processing, complete, failed
}

// Queue defines the interface for a request queue.
type Queue interface {
	Enqueue(ctx context.Context, req *AIRequest) error
	Dequeue(ctx context.Context) (*AIRequest, error)
	MarkComplete(ctx context.Context, requestID string, resp *AIResponse) error
	GetStatus(ctx context.Context, requestID string) (*RequestStatus, error)
	Size() int
}

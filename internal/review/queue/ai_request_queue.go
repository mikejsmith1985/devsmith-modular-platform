// Package queue provides request queuing for AI API calls.
package queue

import (
	"context"
	"errors"
	"sync"
)

// fifoQueue implements a FIFO queue with thread-safe operations
//
//nolint:govet // internal test struct, field alignment not critical
type fifoQueue struct {
	mu        sync.RWMutex
	capacity  int
	requests  []*AIRequest
	responses map[string]*AIResponse
	statuses  map[string]*RequestStatus
}

// NewFIFOQueue creates a new FIFO queue with specified capacity
func NewFIFOQueue(capacity int) Queue {
	if capacity <= 0 {
		capacity = 1000 // Default capacity
	}
	return &fifoQueue{
		capacity:  capacity,
		requests:  make([]*AIRequest, 0, capacity),
		responses: make(map[string]*AIResponse),
		statuses:  make(map[string]*RequestStatus),
	}
}

// Enqueue adds a request to the queue
func (q *fifoQueue) Enqueue(ctx context.Context, req *AIRequest) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if req == nil {
		return errors.New("request cannot be nil")
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.requests) >= q.capacity {
		return ErrQueueFull
	}

	q.requests = append(q.requests, req)

	// Track status
	q.statuses[req.ID] = &RequestStatus{
		RequestID: req.ID,
		State:     "queued",
	}

	return nil
}

// Dequeue retrieves the next request from the queue (FIFO)
func (q *fifoQueue) Dequeue(ctx context.Context) (*AIRequest, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.requests) == 0 {
		return nil, nil
	}

	req := q.requests[0]
	q.requests = q.requests[1:]

	// Update status
	if status, exists := q.statuses[req.ID]; exists {
		status.State = "processing"
	}

	return req, nil
}

// MarkComplete marks a request as complete with response
func (q *fifoQueue) MarkComplete(ctx context.Context, requestID string, resp *AIResponse) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if requestID == "" || resp == nil {
		return errors.New("invalid parameters")
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.responses[requestID] = resp

	// Update status
	if status, exists := q.statuses[requestID]; exists {
		status.State = "complete"
	}

	return nil
}

// GetStatus retrieves the current status of a request
func (q *fifoQueue) GetStatus(ctx context.Context, requestID string) (*RequestStatus, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if requestID == "" {
		return nil, errors.New("invalid request ID")
	}

	q.mu.RLock()
	defer q.mu.RUnlock()

	status, exists := q.statuses[requestID]
	if !exists {
		return nil, errors.New("request not found")
	}

	return status, nil
}

// Size returns the current queue size
func (q *fifoQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.requests)
}

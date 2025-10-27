// Package queue provides request queuing for AI API calls.
package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestQueue_Enqueue_Success tests successful request enqueue
func TestQueue_Enqueue_Success(t *testing.T) {
	// GIVEN: Empty queue
	queue := NewFIFOQueue(100)
	ctx := context.Background()

	// WHEN: Enqueuing a request
	req := &AIRequest{
		ID:       "req-1",
		UserID:   123,
		Mode:     "preview",
		Content:  "code",
		MaxRetry: 3,
	}
	err := queue.Enqueue(ctx, req)

	// THEN: Request is enqueued successfully
	assert.NoError(t, err)
}

// TestQueue_Dequeue_FIFO tests FIFO ordering
func TestQueue_Dequeue_FIFO(t *testing.T) {
	// GIVEN: Queue with 3 requests
	queue := NewFIFOQueue(100)
	ctx := context.Background()

	reqs := []*AIRequest{
		{ID: "req-1", UserID: 1},
		{ID: "req-2", UserID: 2},
		{ID: "req-3", UserID: 3},
	}

	for _, req := range reqs {
		queue.Enqueue(ctx, req)
	}

	// WHEN: Dequeueing all requests
	results := []*AIRequest{}
	for i := 0; i < 3; i++ {
		req, _ := queue.Dequeue(ctx)
		results = append(results, req)
	}

	// THEN: Requests returned in FIFO order
	assert.Equal(t, "req-1", results[0].ID, "First request should be first")
	assert.Equal(t, "req-2", results[1].ID, "Second request should be second")
	assert.Equal(t, "req-3", results[2].ID, "Third request should be third")
}

// TestQueue_Dequeue_Empty tests dequeueing from empty queue
func TestQueue_Dequeue_Empty(t *testing.T) {
	// GIVEN: Empty queue
	queue := NewFIFOQueue(100)
	ctx := context.Background()

	// WHEN: Dequeueing from empty queue
	req, err := queue.Dequeue(ctx)

	// THEN: Returns nil and no error (non-blocking)
	assert.Nil(t, req)
	assert.NoError(t, err)
}

// TestQueue_Size tests queue size tracking
func TestQueue_Size(t *testing.T) {
	// GIVEN: Empty queue
	queue := NewFIFOQueue(100)
	ctx := context.Background()

	// WHEN: Enqueuing requests
	for i := 1; i <= 5; i++ {
		queue.Enqueue(ctx, &AIRequest{ID: "req-" + string(rune(48+i))})
	}

	// THEN: Size is correct
	size := queue.Size()
	assert.Equal(t, 5, size)

	// WHEN: Dequeueing one
	queue.Dequeue(ctx)

	// THEN: Size decreases
	assert.Equal(t, 4, queue.Size())
}

// TestQueue_MarkComplete tests completion tracking
func TestQueue_MarkComplete(t *testing.T) {
	// GIVEN: Queue with request
	queue := NewFIFOQueue(100)
	ctx := context.Background()
	queue.Enqueue(ctx, &AIRequest{ID: "req-1"})

	// WHEN: Marking request as complete
	resp := &AIResponse{
		RequestID: "req-1",
		Result:    "analysis",
		Duration:  100 * time.Millisecond,
	}
	err := queue.MarkComplete(ctx, "req-1", resp)

	// THEN: Completion recorded
	assert.NoError(t, err)
}

// TestQueue_GetStatus tests status retrieval
func TestQueue_GetStatus(t *testing.T) {
	// GIVEN: Queue with request
	queue := NewFIFOQueue(100)
	ctx := context.Background()
	queue.Enqueue(ctx, &AIRequest{
		ID:     "req-1",
		UserID: 123,
	})

	// WHEN: Getting status
	status, err := queue.GetStatus(ctx, "req-1")

	// THEN: Status retrieved
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "req-1", status.RequestID)
	assert.Equal(t, "queued", status.State)
}

// TestQueue_Capacity tests queue capacity limits
func TestQueue_Capacity(t *testing.T) {
	// GIVEN: Queue with capacity 3
	queue := NewFIFOQueue(3)
	ctx := context.Background()

	// WHEN: Enqueuing 3 requests (at capacity)
	for i := 1; i <= 3; i++ {
		err := queue.Enqueue(ctx, &AIRequest{ID: "req-" + string(rune(48+i))})
		assert.NoError(t, err)
	}

	// THEN: 4th request fails
	err := queue.Enqueue(ctx, &AIRequest{ID: "req-4"})
	assert.Error(t, err)
	assert.Equal(t, ErrQueueFull, err)
}

// TestQueue_ContextCancellation tests context handling
func TestQueue_ContextCancellation(t *testing.T) {
	// GIVEN: Queue
	queue := NewFIFOQueue(100)

	// WHEN: Context is cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// THEN: Operations respect cancellation
	req := &AIRequest{ID: "req-1"}
	err := queue.Enqueue(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestQueue_ConcurrentOperations tests thread safety
func TestQueue_ConcurrentOperations(t *testing.T) {
	// GIVEN: Queue
	queue := NewFIFOQueue(100)
	ctx := context.Background()

	// WHEN: Multiple goroutines enqueue and dequeue concurrently
	errorCount := 0
	for i := 1; i <= 10; i++ {
		go func(id int) {
			req := &AIRequest{ID: "req-" + string(rune(48+id))}
			if err := queue.Enqueue(ctx, req); err != nil {
				errorCount++
			}
		}(i)
	}

	// Allow time for all operations
	time.Sleep(100 * time.Millisecond)

	// THEN: All operations succeed (thread-safe)
	assert.Equal(t, 0, errorCount, "All enqueues should succeed")
	assert.Equal(t, 10, queue.Size())
}

// TestQueue_Dequeue_Blocking tests blocking behavior
func TestQueue_Dequeue_Blocking(t *testing.T) {
	// GIVEN: Queue with one request
	queue := NewFIFOQueue(100)
	ctx := context.Background()
	queue.Enqueue(ctx, &AIRequest{ID: "req-1"})

	// WHEN: Dequeueing first request
	req1, _ := queue.Dequeue(ctx)
	assert.NotNil(t, req1)

	// AND: Dequeueing from now-empty queue (non-blocking)
	req2, err := queue.Dequeue(ctx)

	// THEN: Returns nil without waiting
	assert.Nil(t, req2)
	assert.NoError(t, err)
}

// Errors and types
var (
	ErrQueueFull = errors.New("queue is full")
)

// AIRequest represents a queued AI request
type AIRequest struct {
	ID       string
	Mode     string
	Content  string
	UserID   int64
	MaxRetry int
}

// AIResponse represents an AI response
type AIResponse struct {
	Result    interface{}
	RequestID string
	Duration  time.Duration
}

// RequestStatus represents current request status
type RequestStatus struct {
	RequestID string
	State     string // queued, processing, complete, failed
}

// Queue defines the queue interface
type Queue interface {
	Enqueue(ctx context.Context, req *AIRequest) error
	Dequeue(ctx context.Context) (*AIRequest, error)
	MarkComplete(ctx context.Context, requestID string, resp *AIResponse) error
	GetStatus(ctx context.Context, requestID string) (*RequestStatus, error)
	Size() int
}

// NewFIFOQueue creates a new FIFO queue
func NewFIFOQueue(capacity int) Queue {
	return &fifoQueue{capacity: capacity}
}

// Minimal implementation stub
type fifoQueue struct {
	capacity int
}

func (q *fifoQueue) Enqueue(ctx context.Context, req *AIRequest) error {
	return nil
}

func (q *fifoQueue) Dequeue(ctx context.Context) (*AIRequest, error) {
	return nil, nil
}

func (q *fifoQueue) MarkComplete(ctx context.Context, requestID string, resp *AIResponse) error {
	return nil
}

func (q *fifoQueue) GetStatus(ctx context.Context, requestID string) (*RequestStatus, error) {
	return nil, nil
}

func (q *fifoQueue) Size() int {
	return 0
}

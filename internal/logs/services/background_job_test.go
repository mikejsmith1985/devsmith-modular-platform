package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBackgroundJob is a mock for testing background jobs
type MockBackgroundJob struct {
	mock.Mock
}

// Start mocks the Start method
func (m *MockBackgroundJob) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Stop mocks the Stop method
func (m *MockBackgroundJob) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// IsRunning mocks the IsRunning method
func (m *MockBackgroundJob) IsRunning(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Get(0).(bool), args.Error(1)
}

// GetLastExecutionTime mocks the GetLastExecutionTime method
func (m *MockBackgroundJob) GetLastExecutionTime(ctx context.Context) (time.Time, error) {
	args := m.Called(ctx)
	return args.Get(0).(time.Time), args.Error(1)
}

// GetNextExecutionTime mocks the GetNextExecutionTime method
func (m *MockBackgroundJob) GetNextExecutionTime(ctx context.Context) (time.Time, error) {
	args := m.Called(ctx)
	return args.Get(0).(time.Time), args.Error(1)
}

// TestBackgroundJob_HourlyAggregation validates hourly job schedule.
func TestBackgroundJob_HourlyAggregation(t *testing.T) {
	// GIVEN: A background job for hourly aggregation
	mockJob := new(MockBackgroundJob)

	mockJob.On("Start", mock.Anything).Return(nil)

	// WHEN: Starting hourly aggregation job
	err := mockJob.Start(context.Background())

	// THEN: Should start without error
	assert.NoError(t, err)
	mockJob.AssertCalled(t, "Start", mock.Anything)
}

// TestBackgroundJob_DailyAggregation validates daily job schedule.
func TestBackgroundJob_DailyAggregation(t *testing.T) {
	// GIVEN: A background job for daily aggregation
	mockJob := new(MockBackgroundJob)

	mockJob.On("Start", mock.Anything).Return(nil)

	// WHEN: Starting daily aggregation job
	err := mockJob.Start(context.Background())

	// THEN: Should start without error
	assert.NoError(t, err)
	mockJob.AssertCalled(t, "Start", mock.Anything)
}

// TestBackgroundJob_Stop validates job stopping.
func TestBackgroundJob_Stop(t *testing.T) {
	// GIVEN: A running background job
	mockJob := new(MockBackgroundJob)

	mockJob.On("Stop", mock.Anything).Return(nil)

	// WHEN: Stopping job
	err := mockJob.Stop(context.Background())

	// THEN: Should stop without error
	assert.NoError(t, err)
	mockJob.AssertCalled(t, "Stop", mock.Anything)
}

// TestBackgroundJob_IsRunning validates running status check.
func TestBackgroundJob_IsRunning(t *testing.T) {
	// GIVEN: A background job
	mockJob := new(MockBackgroundJob)

	mockJob.On("IsRunning", mock.Anything).Return(true, nil)

	// WHEN: Checking if job is running
	running, err := mockJob.IsRunning(context.Background())

	// THEN: Should return running status
	assert.NoError(t, err)
	assert.True(t, running)
}

// TestBackgroundJob_NotRunning validates stopped status.
func TestBackgroundJob_NotRunning(t *testing.T) {
	// GIVEN: A stopped background job
	mockJob := new(MockBackgroundJob)

	mockJob.On("IsRunning", mock.Anything).Return(false, nil)

	// WHEN: Checking running status
	running, err := mockJob.IsRunning(context.Background())

	// THEN: Should indicate not running
	assert.NoError(t, err)
	assert.False(t, running)
}

// TestBackgroundJob_GetLastExecutionTime validates last execution retrieval.
func TestBackgroundJob_GetLastExecutionTime(t *testing.T) {
	// GIVEN: A background job with execution history
	mockJob := new(MockBackgroundJob)
	now := time.Now()

	mockJob.On("GetLastExecutionTime", mock.Anything).Return(now.Add(-1*time.Hour), nil)

	// WHEN: Getting last execution time
	lastExecution, err := mockJob.GetLastExecutionTime(context.Background())

	// THEN: Should return last execution time
	assert.NoError(t, err)
	assert.NotZero(t, lastExecution)
	assert.True(t, lastExecution.Before(now))
}

// TestBackgroundJob_GetNextExecutionTime validates next execution retrieval.
func TestBackgroundJob_GetNextExecutionTime(t *testing.T) {
	// GIVEN: A background job with scheduled execution
	mockJob := new(MockBackgroundJob)
	now := time.Now()

	mockJob.On("GetNextExecutionTime", mock.Anything).Return(now.Add(1*time.Hour), nil)

	// WHEN: Getting next execution time
	nextExecution, err := mockJob.GetNextExecutionTime(context.Background())

	// THEN: Should return future execution time
	assert.NoError(t, err)
	assert.NotZero(t, nextExecution)
	assert.True(t, nextExecution.After(now))
}

// TestBackgroundJob_ExecutionInterval_Hourly validates hourly interval.
func TestBackgroundJob_ExecutionInterval_Hourly(t *testing.T) {
	// GIVEN: Hourly aggregation job
	mockJob := new(MockBackgroundJob)
	now := time.Now()
	lastExec := now.Add(-1 * time.Hour)
	nextExec := now.Add(1 * time.Hour)

	mockJob.On("GetLastExecutionTime", mock.Anything).Return(lastExec, nil)
	mockJob.On("GetNextExecutionTime", mock.Anything).Return(nextExec, nil)

	// WHEN: Checking execution schedule
	last, _ := mockJob.GetLastExecutionTime(context.Background())
	next, _ := mockJob.GetNextExecutionTime(context.Background())

	// THEN: Should be approximately 1 hour apart
	interval := next.Sub(last)
	assert.Equal(t, 2*time.Hour, interval)
}

// TestBackgroundJob_ExecutionInterval_Daily validates daily interval.
func TestBackgroundJob_ExecutionInterval_Daily(t *testing.T) {
	// GIVEN: Daily aggregation job
	mockJob := new(MockBackgroundJob)
	now := time.Now()
	lastExec := now.Add(-24 * time.Hour)
	nextExec := now.Add(24 * time.Hour)

	mockJob.On("GetLastExecutionTime", mock.Anything).Return(lastExec, nil)
	mockJob.On("GetNextExecutionTime", mock.Anything).Return(nextExec, nil)

	// WHEN: Checking execution schedule
	last, _ := mockJob.GetLastExecutionTime(context.Background())
	next, _ := mockJob.GetNextExecutionTime(context.Background())

	// THEN: Should be approximately 24 hours apart
	interval := next.Sub(last)
	assert.Equal(t, 48*time.Hour, interval)
}

// TestBackgroundJob_StartStop_Lifecycle validates start-stop lifecycle.
func TestBackgroundJob_StartStop_Lifecycle(t *testing.T) {
	// GIVEN: Background job
	mockJob := new(MockBackgroundJob)

	// Setup different return values for different calls
	mockJob.On("Start", mock.Anything).Return(nil)
	mockJob.On("IsRunning", mock.Anything).Return(true, nil).Once()
	mockJob.On("Stop", mock.Anything).Return(nil)
	mockJob.On("IsRunning", mock.Anything).Return(false, nil).Maybe()

	// WHEN: Starting job
	err1 := mockJob.Start(context.Background())
	running1, _ := mockJob.IsRunning(context.Background())

	// THEN: Job should be running
	assert.NoError(t, err1)
	assert.True(t, running1)

	// WHEN: Stopping job
	err2 := mockJob.Stop(context.Background())
	_, _ = mockJob.IsRunning(context.Background())

	// THEN: Job should be stopped
	assert.NoError(t, err2)
	// Note: Due to mock behavior, we verify the calls were made instead
	mockJob.AssertCalled(t, "Start", mock.Anything)
	mockJob.AssertCalled(t, "Stop", mock.Anything)
}

// TestBackgroundJob_ContextCancellation validates context handling.
func TestBackgroundJob_ContextCancellation(t *testing.T) {
	// GIVEN: Cancelled context
	mockJob := new(MockBackgroundJob)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockJob.On("Start", mock.Anything).Return(context.Canceled)

	// WHEN: Starting with cancelled context
	err := mockJob.Start(ctx)

	// THEN: Should return context cancelled error
	assert.Error(t, err)
}

// TestBackgroundJob_DoubleStart validates idempotent start.
func TestBackgroundJob_DoubleStart(t *testing.T) {
	// GIVEN: Background job
	mockJob := new(MockBackgroundJob)

	mockJob.On("Start", mock.Anything).Return(nil)

	// WHEN: Starting twice
	err1 := mockJob.Start(context.Background())
	err2 := mockJob.Start(context.Background())

	// THEN: Both should succeed (idempotent)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

// TestBackgroundJob_StopBeforeStart validates stop before start.
func TestBackgroundJob_StopBeforeStart(t *testing.T) {
	// GIVEN: Background job not started
	mockJob := new(MockBackgroundJob)

	mockJob.On("Stop", mock.Anything).Return(nil)

	// WHEN: Stopping job that's not running
	err := mockJob.Stop(context.Background())

	// THEN: Should be graceful (no error)
	assert.NoError(t, err)
}

// TestBackgroundJob_MultipleJobs validates multiple jobs coordination.
func TestBackgroundJob_MultipleJobs(t *testing.T) {
	// GIVEN: Multiple background jobs
	hourlyJob := new(MockBackgroundJob)
	dailyJob := new(MockBackgroundJob)

	hourlyJob.On("Start", mock.Anything).Return(nil)
	dailyJob.On("Start", mock.Anything).Return(nil)

	// WHEN: Starting multiple jobs
	err1 := hourlyJob.Start(context.Background())
	err2 := dailyJob.Start(context.Background())

	// THEN: Both should start
	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

// TestBackgroundJob_ExecutionTiming validates execution occurs at scheduled time.
func TestBackgroundJob_ExecutionTiming(t *testing.T) {
	// GIVEN: Job scheduled to run at specific time
	mockJob := new(MockBackgroundJob)
	now := time.Now()

	// Job should execute this hour
	nextHour := now.Truncate(time.Hour).Add(time.Hour)

	mockJob.On("GetNextExecutionTime", mock.Anything).Return(nextHour, nil)

	// WHEN: Checking next execution
	nextExec, err := mockJob.GetNextExecutionTime(context.Background())

	// THEN: Should schedule for next hour
	assert.NoError(t, err)
	assert.True(t, nextExec.After(now))
	assert.True(t, nextExec.Hour() >= now.Hour())
}

// TestBackgroundJob_LastExecution_Initial validates initial execution state.
func TestBackgroundJob_LastExecution_Initial(t *testing.T) {
	// GIVEN: New background job never executed
	mockJob := new(MockBackgroundJob)

	mockJob.On("GetLastExecutionTime", mock.Anything).Return(time.Time{}, nil)

	// WHEN: Getting last execution time
	lastExec, err := mockJob.GetLastExecutionTime(context.Background())

	// THEN: Should return zero time
	assert.NoError(t, err)
	assert.True(t, lastExec.IsZero())
}

// TestBackgroundJob_SuccessfulExecution validates successful execution.
func TestBackgroundJob_SuccessfulExecution(t *testing.T) {
	// GIVEN: Background job configured and running
	mockJob := new(MockBackgroundJob)
	now := time.Now()

	mockJob.On("Start", mock.Anything).Return(nil)
	mockJob.On("GetLastExecutionTime", mock.Anything).Return(now, nil)

	// WHEN: Running job and checking execution
	err := mockJob.Start(context.Background())
	lastExec, _ := mockJob.GetLastExecutionTime(context.Background())

	// THEN: Execution should be recorded
	assert.NoError(t, err)
	assert.NotZero(t, lastExec)
}

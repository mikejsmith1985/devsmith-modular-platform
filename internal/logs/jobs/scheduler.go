// Package jobs provides background job scheduling and execution.
package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Job represents a background job definition.
type Job struct { //nolint:govet // struct alignment optimized for readability
	Fn       func(context.Context) error
	Name     string
	Interval time.Duration
}

// Scheduler manages background jobs.
type Scheduler struct { //nolint:govet // struct alignment optimized for readability
	stopChan chan struct{}
	logger   *logrus.Logger
	jobs     []Job
	wg       sync.WaitGroup
	mu       sync.RWMutex
	running  bool
}

// NewScheduler creates a new Scheduler.
func NewScheduler(logger *logrus.Logger) *Scheduler {
	return &Scheduler{
		jobs:     make([]Job, 0),
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// Register adds a job to the scheduler.
func (s *Scheduler) Register(job Job) error {
	if job.Name == "" {
		return fmt.Errorf("job name cannot be empty")
	}
	if job.Interval <= 0 {
		return fmt.Errorf("job interval must be positive")
	}
	if job.Fn == nil {
		return fmt.Errorf("job function cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs = append(s.jobs, job)
	s.logger.Infof("Registered background job: %s (interval: %v)", job.Name, job.Interval)

	return nil
}

// Start begins the scheduler loop.
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler already running")
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("Starting job scheduler")

	// Start each job in a goroutine
	for _, job := range s.jobs {
		s.wg.Add(1)
		go s.runJob(ctx, job)
	}

	return nil
}

// Stop stops the scheduler.
func (s *Scheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler not running")
	}
	s.running = false
	s.mu.Unlock()

	s.logger.Info("Stopping job scheduler")

	// Signal all goroutines to stop
	close(s.stopChan)

	// Wait for all goroutines to finish or timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("All jobs stopped successfully")
	case <-ctx.Done():
		s.logger.Warn("Job scheduler stop timeout")
		return ctx.Err()
	}

	return nil
}

// IsRunning returns whether the scheduler is running.
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// runJob executes a job repeatedly at the specified interval.
func (s *Scheduler) runJob(ctx context.Context, job Job) {
	defer s.wg.Done()

	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	s.logger.Infof("Started job: %s", job.Name)

	// Run job once immediately
	s.executeJob(ctx, job)

	// Run job repeatedly
	for {
		select {
		case <-s.stopChan:
			s.logger.Infof("Stopped job: %s", job.Name)
			return
		case <-ctx.Done():
			s.logger.Infof("Context cancelled for job: %s", job.Name)
			return
		case <-ticker.C:
			s.executeJob(ctx, job)
		}
	}
}

// executeJob runs a single job execution.
func (s *Scheduler) executeJob(ctx context.Context, job Job) {
	s.logger.Debugf("Executing job: %s", job.Name)

	start := time.Now()
	err := job.Fn(ctx)
	duration := time.Since(start)

	if err != nil {
		s.logger.WithError(err).Errorf("Job failed: %s (duration: %v)", job.Name, duration)
	} else {
		s.logger.Debugf("Job completed: %s (duration: %v)", job.Name, duration)
	}
}

// JobBuilder helps construct job definitions.
type JobBuilder struct { //nolint:govet // struct alignment optimized for readability
	fn       func(context.Context) error
	name     string
	interval time.Duration
}

// NewJobBuilder creates a new JobBuilder.
func NewJobBuilder(name string) *JobBuilder {
	return &JobBuilder{name: name}
}

// WithInterval sets the job interval.
func (jb *JobBuilder) WithInterval(interval time.Duration) *JobBuilder {
	jb.interval = interval
	return jb
}

// WithFunction sets the job function.
func (jb *JobBuilder) WithFunction(fn func(context.Context) error) *JobBuilder {
	jb.fn = fn
	return jb
}

// Build constructs the Job.
func (jb *JobBuilder) Build() Job {
	return Job{
		Name:     jb.name,
		Interval: jb.interval,
		Fn:       jb.fn,
	}
}

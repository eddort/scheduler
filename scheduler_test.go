package scheduler

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScheduler_SlowTask(t *testing.T) {
	const slowTaskInterval = 2 * time.Second
	const slowTaskDuration = 3 * time.Second

	taskExecutionCounter := int32(0)
	taskUnsuccessfulCounter := int32(0)

	action := func(payload Payload) error {
		atomic.AddInt32(&taskExecutionCounter, 1)

		time.Sleep(slowTaskDuration)

		return errors.New("Task unsuccessful")
	}

	taskMiddleware := func(next ActionFunc) ActionFunc {
		return func(payload Payload) error {
			err := next(payload)
			if err != nil {
				atomic.AddInt32(&taskUnsuccessfulCounter, 1)
			}
			return err
		}
	}

	cfg := TaskConfig{
		Name:        "slow_task",
		Interval:    slowTaskInterval,
		Action:      action,
		Deadline:    slowTaskDuration,
		Middlewares: []Middleware{taskMiddleware},
	}

	s := New()
	s.RegisterTask(cfg)

	// Use sync.WaitGroup for synchronization
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Start()
	}()

	// Wait for a few task intervals to pass
	time.Sleep(7 * time.Second)
	s.Stop()

	wg.Wait()

	executionCount := atomic.LoadInt32(&taskExecutionCounter)
	unsuccessfulCount := atomic.LoadInt32(&taskUnsuccessfulCounter)

	assert.GreaterOrEqual(t, executionCount, int32(2), "Expected at least 2 task executions")
	assert.Equal(t, executionCount, unsuccessfulCount, "Expected unsuccessful task executions to be equal to the total task executions")
}
func TestScheduler_SuccessfulTask(t *testing.T) {
	const taskInterval = 1 * time.Second
	const taskDuration = 500 * time.Millisecond

	taskExecutionCounter := int32(0)
	taskSuccessfulCounter := int32(0)

	action := func(payload Payload) error {
		atomic.AddInt32(&taskExecutionCounter, 1)
		time.Sleep(taskDuration)
		return nil
	}

	taskMiddleware := func(next ActionFunc) ActionFunc {
		return func(payload Payload) error {
			err := next(payload)
			if err == nil {
				atomic.AddInt32(&taskSuccessfulCounter, 1)
			}
			return err
		}
	}

	cfg := TaskConfig{
		Name:        "successful_task",
		Interval:    taskInterval,
		Action:      action,
		Deadline:    2 * taskDuration,
		Middlewares: []Middleware{taskMiddleware},
	}

	s := New()
	s.RegisterTask(cfg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Start()
	}()

	time.Sleep(5 * time.Second)
	s.Stop()

	wg.Wait()

	executionCount := atomic.LoadInt32(&taskExecutionCounter)
	successfulCount := atomic.LoadInt32(&taskSuccessfulCounter)

	assert.GreaterOrEqual(t, executionCount, int32(4), "Expected at least 4 task executions")
	assert.Equal(t, executionCount, successfulCount, "Expected successful task executions to be equal to the total task executions")
}

func TestScheduler_DeadlineExceeded(t *testing.T) {
	const taskInterval = 200 * time.Millisecond
	const taskDuration = 300 * time.Millisecond

	taskExecutionCounter := int32(0)
	taskDeadlineExceededCounter := int32(0)

	action := func(payload Payload) error {
		atomic.AddInt32(&taskExecutionCounter, 1)
		time.Sleep(taskDuration)
		return nil
	}

	taskMiddleware := func(next ActionFunc) ActionFunc {
		return func(payload Payload) error {
			err := next(payload)
			if errors.Is(err, ErrDeadlineExceeded) {
				atomic.AddInt32(&taskDeadlineExceededCounter, 1)
			}
			return err
		}
	}

	cfg := TaskConfig{
		Name:        "deadline_exceeded_task",
		Interval:    taskInterval,
		Action:      action,
		Deadline:    taskDuration - 100*time.Millisecond,
		Middlewares: []Middleware{taskMiddleware},
	}

	s := New()
	s.RegisterTask(cfg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Start()
	}()

	time.Sleep(700 * time.Millisecond)
	s.Stop()

	wg.Wait()

	executionCount := atomic.LoadInt32(&taskExecutionCounter)
	deadlineExceededCount := atomic.LoadInt32(&taskDeadlineExceededCounter)

	assert.GreaterOrEqual(t, executionCount, int32(3), "Expected at least 3 task executions")
	assert.Equal(t, executionCount, deadlineExceededCount, "Expected deadline exceeded task executions to be equal to the total task executions")
}

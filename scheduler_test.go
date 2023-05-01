package scheduler_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eddort/scheduler"
)

func TestScheduler(t *testing.T) {
	loggingMiddleware := func(next scheduler.ActionFunc) scheduler.ActionFunc {
		return func(payload scheduler.Payload) error {
			t.Logf("Running task: %s", payload.Name)
			return next(payload)
		}
	}

	taskExecuted := false

	taskAction := func(payload scheduler.Payload) error {
		taskExecuted = true
		return nil
	}

	s := scheduler.New(loggingMiddleware)

	s.RegisterTask(scheduler.TaskConfig{
		Name:     "test-task",
		Interval: time.Millisecond * 100,
		Action:   taskAction,
	})

	time.Sleep(time.Millisecond * 200)

	s.Stop()

	assert.True(t, taskExecuted, "Task should be executed")
}

func TestSchedulerWithDeadline(t *testing.T) {
	taskExecuted := false

	taskAction := func(payload scheduler.Payload) error {
		time.Sleep(time.Millisecond * 150)
		taskExecuted = true
		return nil
	}

	s := scheduler.New()

	s.RegisterTask(scheduler.TaskConfig{
		Name:     "test-task-deadline",
		Interval: time.Millisecond * 100,
		Action:   taskAction,
		Deadline: time.Millisecond * 50,
	})

	time.Sleep(time.Millisecond * 200)

	s.Stop()

	assert.False(t, taskExecuted, "Task should not be executed")
}

func TestSchedulerWithMiddlewareError(t *testing.T) {
	errorMiddleware := func(next scheduler.ActionFunc) scheduler.ActionFunc {
		return func(payload scheduler.Payload) error {
			return errors.New("middleware error")
		}
	}

	taskExecuted := false

	taskAction := func(payload scheduler.Payload) error {
		taskExecuted = true
		return nil
	}

	s := scheduler.New(errorMiddleware)

	s.RegisterTask(scheduler.TaskConfig{
		Name:     "test-task-error",
		Interval: time.Millisecond * 100,
		Action:   taskAction,
	})

	time.Sleep(time.Millisecond * 200)

	s.Stop()

	assert.False(t, taskExecuted, "Task should not be executed")
}

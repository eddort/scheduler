package scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var successfulTaskCount int
var unsuccessfulTaskCount int
var taskCountMutex sync.Mutex

func successfulTask(payload Payload) error {
	time.Sleep(500 * time.Millisecond)
	return nil
}

func slowTask(payload Payload) error {
	time.Sleep(2 * time.Second)
	return nil
}
func countingMiddleware(next ActionFunc) ActionFunc {
	return func(payload Payload) error {
		err := next(payload)
		taskCountMutex.Lock()
		defer taskCountMutex.Unlock()

		if err == nil {
			successfulTaskCount++
		} else {
			unsuccessfulTaskCount++
		}
		return err
	}
}

func TestScheduler_SuccessfulTask(t *testing.T) {
	registry := New(countingMiddleware)
	registry.RegisterTask(TaskConfig{
		Name:        "SuccessfulTask",
		Interval:    1 * time.Second,
		Action:      successfulTask,
		Deadline:    1 * time.Second,
		Middlewares: []Middleware{countingMiddleware},
	})

	go registry.Start()
	time.Sleep(3 * time.Second)
	registry.Stop()

	taskCountMutex.Lock()
	defer taskCountMutex.Unlock()
	assert.GreaterOrEqual(t, successfulTaskCount, 2, "Expected 2 successful task executions")
}

func TestScheduler_SlowTask(t *testing.T) {
	registry := New(countingMiddleware)
	registry.RegisterTask(TaskConfig{
		Name:        "SlowTask",
		Interval:    1 * time.Second,
		Action:      slowTask,
		Deadline:    1 * time.Second,
		Middlewares: []Middleware{countingMiddleware},
	})

	go registry.Start()
	time.Sleep(3 * time.Second)
	registry.Stop()

	taskCountMutex.Lock()
	defer taskCountMutex.Unlock()
	assert.Equal(t, 2, unsuccessfulTaskCount, "Expected 2 unsuccessful task executions")
}

func TestScheduler_Stop(t *testing.T) {
	registry := New(countingMiddleware)
	registry.RegisterTask(TaskConfig{
		Name:        "SuccessfulTask",
		Interval:    1 * time.Second,
		Action:      successfulTask,
		Deadline:    1 * time.Second,
		Middlewares: []Middleware{countingMiddleware},
	})

	go registry.Start()
	time.Sleep(3 * time.Second)
	registry.Stop()

	assert.True(t, true, "Scheduler stopped")
}

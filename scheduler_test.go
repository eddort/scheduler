package scheduler

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestScheduler(t *testing.T) {
	logger, _ := test.NewNullLogger()
	s := New(logger)

	var count1 int32
	s.RegisterTask("task1", 100*time.Millisecond, 500*time.Millisecond, func() {
		atomic.AddInt32(&count1, 1)
		time.Sleep(200 * time.Millisecond)
	})

	var count2 int32
	s.RegisterTask("task2", 50*time.Millisecond, 50*time.Millisecond, func() {
		atomic.AddInt32(&count2, 1)
	})

	s.Start()

	time.Sleep(1 * time.Second)

	c1 := atomic.LoadInt32(&count1)
	c2 := atomic.LoadInt32(&count2)

	assert.Greater(t, c1, int32(1), "Task 1 should have been executed more than once")
	assert.Less(t, c1, int32(10), "Task 1 should not be executed too many times")

	assert.Greater(t, c2, int32(1), "Task 2 should have been executed more than once")
	assert.Less(t, c2, int32(25), "Task 2 should not be executed too many times")
}

func TestSchedulerDeadline(t *testing.T) {
	logger, hook := test.NewNullLogger()
	s := New(logger)

	var count int32
	s.RegisterTask("task3", 100*time.Millisecond, 100*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
		time.Sleep(200 * time.Millisecond)
	})

	s.Start()

	time.Sleep(1 * time.Second)

	c := atomic.LoadInt32(&count)

	assert.Greater(t, c, int32(1), "Task 3 should have been executed more than once")
	assert.Less(t, c, int32(10), "Task 3 should not be executed too many times")

	found := false
	for _, entry := range hook.AllEntries() {
		if entry.Message == "Task task3 reached its deadline and was terminated" {
			found = true
			break
		}
	}
	assert.True(t, found, "Task 3 should have a deadline termination log entry")
}

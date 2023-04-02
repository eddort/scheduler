package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type TaskFunc func()

type Task struct {
	Name     string
	Interval time.Duration
	Deadline time.Duration
	Function TaskFunc
}

type Scheduler struct {
	tasks   map[string]*Task
	running map[string]bool
	mu      sync.Mutex
	logger  *logrus.Logger
}

func New(logger *logrus.Logger) *Scheduler {
	return &Scheduler{
		tasks:   make(map[string]*Task),
		running: make(map[string]bool),
		logger:  logger,
	}
}

func (s *Scheduler) RegisterTask(name string, interval, deadline time.Duration, function TaskFunc) {
	s.tasks[name] = &Task{
		Name:     name,
		Interval: interval,
		Deadline: deadline,
		Function: function,
	}
}

func (s *Scheduler) Start() {
	for _, task := range s.tasks {
		go func(task *Task) {
			ticker := time.NewTicker(task.Interval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					s.executeTask(task)
				}
			}
		}(task)
	}
}

func (s *Scheduler) executeTask(task *Task) {
	s.mu.Lock()
	if s.running[task.Name] {
		s.mu.Unlock()
		return
	}

	s.running[task.Name] = true
	s.mu.Unlock()

	log := s.logger.WithFields(logrus.Fields{
		"name":     task.Name,
		"interval": task.Interval,
		"deadline": task.Deadline,
	})

	log.Infof("Starting task: %s", task.Name)

	ctx, cancel := context.WithTimeout(context.Background(), task.Deadline)
	defer cancel()

	done := make(chan struct{})
	go func() {
		task.Function()
		close(done)
	}()

	select {
	case <-done:
		log.Infof("Finished task: %s", task.Name)
	case <-ctx.Done():
		log.Warnf("Task reached its deadline and was terminated")
	}

	s.mu.Lock()
	s.running[task.Name] = false
	s.mu.Unlock()
}

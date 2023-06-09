package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrDeadlineExceeded = errors.New("Deadline exceeded")

type Registry struct {
	wg          sync.WaitGroup
	tasks       []*Task
	middlewares *[]Middleware
}

type Task struct {
	Name        string
	Interval    time.Duration
	Deadline    time.Duration
	Action      ActionFunc
	Cancel      context.CancelFunc
	Middlewares []Middleware
	Ctx         context.Context
}

type Payload struct {
	Name     string
	Interval time.Duration
	Deadline time.Duration
}

type ActionFunc func(Payload) error
type Middleware func(ActionFunc) ActionFunc

func runWithTimeout(fn func(Payload) error, timeout time.Duration) ActionFunc {
	return func(payload Payload) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		runtimeError := make(chan error, 1)
		go func() {
			runtimeError <- fn(payload)
		}()

		select {
		case <-ctx.Done():
			return ErrDeadlineExceeded
		case result := <-runtimeError:
			return result
		}
	}
}

func New(middlewares ...Middleware) *Registry {
	s := &Registry{
		tasks:       make([]*Task, 0),
		middlewares: &middlewares,
	}

	return s
}

type TaskConfig struct {
	Name        string
	Interval    time.Duration
	Action      ActionFunc
	Deadline    time.Duration
	Middlewares []Middleware
}

func (s *Registry) RegisterTask(cfg TaskConfig) {
	ctx, cancel := context.WithCancel(context.Background())

	if cfg.Deadline < 0 {
		cfg.Deadline = 1 * time.Hour
	}

	task := &Task{
		Name:        cfg.Name,
		Interval:    cfg.Interval,
		Deadline:    cfg.Deadline,
		Action:      cfg.Action,
		Middlewares: cfg.Middlewares,
		Cancel:      cancel,
		Ctx:         ctx,
	}

	s.tasks = append(s.tasks, task)

}

func (s *Registry) Start() {
	for _, task := range s.tasks {
		s.wg.Add(1)
		go s.watch(task)
	}
	s.wg.Wait()
}

func (s *Registry) watch(task *Task) {
	defer s.wg.Done()

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	action := runWithTimeout(task.Action, task.Deadline)
	for _, middleware := range task.Middlewares {
		action = middleware(action)
	}

	for {
		select {
		case <-ticker.C:
			done := make(chan struct{})

			go func() {
				payload := Payload{
					Name:     task.Name,
					Interval: task.Interval,
					Deadline: task.Deadline,
				}

				action(payload)
				close(done)
			}()

			<-done

		case <-task.Ctx.Done():
			return
		}
	}

}

func (s *Registry) Stop() {
	for _, task := range s.tasks {
		task.Cancel()
	}
	s.wg.Wait()
}

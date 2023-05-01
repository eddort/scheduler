package scheduler

import (
	"context"
	"sync"
	"time"
)

type Scheduler struct {
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

func BuildMiddlewareChain(action ActionFunc, middlewares *[]Middleware) ActionFunc {
	if middlewares == nil {
		return action
	}

	chain := action
	for i := len(*middlewares) - 1; i >= 0; i-- {
		chain = (*middlewares)[i](chain)
	}

	return chain
}

func New(middlewares ...Middleware) *Scheduler {
	s := &Scheduler{
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

func (s *Scheduler) RegisterTask(cfg TaskConfig) {
	ctx, cancel := context.WithCancel(context.Background())

	task := &Task{
		Name:        cfg.Name,
		Interval:    cfg.Interval,
		Action:      cfg.Action,
		Middlewares: cfg.Middlewares,
		Cancel:      cancel,
		Ctx:         ctx,
	}

	s.tasks = append(s.tasks, task)

}

func (s *Scheduler) Start() {
	for _, task := range s.tasks {
		s.wg.Add(1)
		go s.watch(task)
	}
	s.wg.Wait()
}

func (s *Scheduler) watch(task *Task) {
	defer s.wg.Done()

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	action := task.Action
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

			if task.Deadline > 0 {
				select {
				case <-done:
				case <-time.After(task.Deadline):
				}
			} else {
				<-done
			}
		case <-task.Ctx.Done():
			return
		}
	}

}

func (s *Scheduler) Stop() {
	for _, task := range s.tasks {
		task.Cancel()
	}
	s.wg.Wait()
}

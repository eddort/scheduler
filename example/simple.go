package main

import (
	"fmt"
	"time"

	"github.com/eddort/scheduler"
)

func main() {
	s := scheduler.New()

	taskConfig := scheduler.TaskConfig{
		Name:     "PrintTime",
		Interval: 2 * time.Second,
		Deadline: 3 * time.Second,
		Action: func(payload scheduler.Payload) error {
			fmt.Println("Current time:", time.Now())
			return nil
		},
		Middlewares: []scheduler.Middleware{func(next scheduler.ActionFunc) scheduler.ActionFunc {
			return func(payload scheduler.Payload) error {
				fmt.Println("Before task execution:", payload.Name)
				err := next(payload)
				fmt.Println("After task execution:", payload.Name)
				return err
			}
		}},
	}

	// Register and start the task
	s.RegisterTask(taskConfig)
	s.Start()
}

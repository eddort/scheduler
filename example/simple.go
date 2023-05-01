package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/eddort/scheduler"
)

func main() {

	counter := 0

	registry := scheduler.New()

	taskConfig := scheduler.TaskConfig{
		Name:     "PrintTime",
		Interval: 10 * time.Microsecond,
		Deadline: 3 * time.Second,
		Action: func(payload scheduler.Payload) error {
			fmt.Println("Current time:", time.Now())

			if counter%2 == 0 {
				time.Sleep(1 * time.Second)
				return nil
			}
			return errors.New("what's up")
		},
		Middlewares: []scheduler.Middleware{func(next scheduler.ActionFunc) scheduler.ActionFunc {
			return func(payload scheduler.Payload) error {
				counter++

				fmt.Println("Before task execution:", payload.Name)

				err := next(payload)

				if err != nil {
					fmt.Println("Task finished with an error:", err)
				} else {
					fmt.Println("Task finished correctly")
				}

				return err
			}
		}},
	}

	// Register and start the task
	registry.RegisterTask(taskConfig)
	go func() {
		time.Sleep(20 * time.Second)
		registry.Stop()
	}()
	registry.Start()
}

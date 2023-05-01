# Scheduler

A simple, concurrent task scheduler written in Go.

## Features

- Register tasks with unique names, intervals, and deadlines
- Concurrently execute tasks in separate goroutines
- Automatically prevent multiple instances of the same task from running simultaneously
- Gracefully handle task deadlines with context cancellation

## Usage

```go
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
		Interval: 2 * time.Second,
		Deadline: 3 * time.Second,
		Action: func(payload scheduler.Payload) error {
			fmt.Println("Current time:", time.Now())
			time.Sleep(1 * time.Second)
			if counter%2 == 0 {
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

```

## Installation

```bash
go get -u github.com/eddort/scheduler
```

## Testing

We use the [testify](https://github.com/stretchr/testify) library for testing. Install it with:

```bash
go get -u github.com/stretchr/testify
```

Then run tests with:

```bash
go test
```

## License

This project is licensed under the MIT License.
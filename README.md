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
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/eddort/scheduler"
)

func main() {
	logger := logrus.New()
	s := scheduler.New(logger)

	s.RegisterTask("task1", 1*time.Second, 5*time.Second, func() {
		fmt.Println("Executing task 1")
		time.Sleep(3 * time.Second)
	})

	s.RegisterTask("task2", 2*time.Second, 1*time.Second, func() {
		fmt.Println("Executing task 2")
		time.Sleep(500 * time.Millisecond)
	})

	s.Start()
	time.Sleep(10 * time.Second)
	s.Stop()
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
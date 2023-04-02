package main

import (
	"fmt"
	"time"

	"github.com/eddort/scheduler"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	s := scheduler.New(logger)

	s.RegisterTask("task1", 1*time.Second, 1*time.Second, func() {
		fmt.Println("Executing task 1")
		time.Sleep(2 * time.Second)
	})

	s.RegisterTask("task2", 2*time.Second, 1*time.Second, func() {
		fmt.Println("Executing task with deadline error 2")
		time.Sleep(1 * time.Second)
	})

	s.Start()

	select {}
}

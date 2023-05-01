package scheduler_test

import (
	"errors"
	"testing"

	scheduler "github.com/eddort/scheduler"

	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	var counter int

	action := func(payload scheduler.Payload) error {
		counter++
		return nil
	}

	middleware := func(next scheduler.ActionFunc) scheduler.ActionFunc {
		return func(payload scheduler.Payload) error {
			counter++
			return next(payload)
		}
	}

	middlewares := []scheduler.Middleware{middleware}

	chain := scheduler.BuildMiddlewareChain(action, &middlewares)
	err := chain(scheduler.Payload{})

	require.NoError(t, err)
	require.Equal(t, 2, counter)
}

func TestMiddlewareErrorHandling(t *testing.T) {
	expectedError := errors.New("some error")

	action := func(payload scheduler.Payload) error {
		return expectedError
	}

	var handledError error
	middleware := func(next scheduler.ActionFunc) scheduler.ActionFunc {
		return func(payload scheduler.Payload) error {
			err := next(payload)
			if err != nil {
				handledError = err
			}
			return err
		}
	}

	middlewares := []scheduler.Middleware{middleware}

	chain := scheduler.BuildMiddlewareChain(action, &middlewares)
	err := chain(scheduler.Payload{})

	require.Error(t, err)
	require.Equal(t, expectedError, handledError)
}

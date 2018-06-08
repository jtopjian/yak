package connections

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/mitchellh/go-linereader"
)

const (
	initialBackoffDelay = 2 * time.Second
	maxBackoffDelay     = 10 * time.Second
)

// Based off of Terraform's remote and local provisioners.
func printOutput(output io.Writer, r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)

	lr := linereader.New(r)
	for line := range lr.Ch {
		fmt.Fprintln(output, line)
	}
}

// Based off of Terraform's remote-exec provisioner.
// This will run a function one time. If the length of time
// specified as `timeout` is reached, the function is cancelled.
func timeoutFunc(timeout int, f func() error) error {
	t := time.Duration(timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	type errWrap struct {
		E error
	}

	var errVal atomic.Value
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		err := f()
		errVal.Store(&errWrap{err})
		return
	}()

	select {
	case <-ctx.Done():
	case <-doneCh:
	}

	switch ctx.Err() {
	case context.Canceled:
		return fmt.Errorf("interrupted")
	case context.DeadlineExceeded:
		return fmt.Errorf("timeout")
	}

	if ev, ok := errVal.Load().(*errWrap); ok {
		return ev.E
	}

	return nil
}

// Again, based off of Terraform's remote-exec provisioner.
// This will retry a function several times until a timeout
// is reached. Each attempt of the function will be delayed
// incrementally.
func retryFunc(timeout int, f func() error) error {
	t := time.Duration(timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	type errWrap struct {
		E error
	}

	var errVal atomic.Value
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)

		delay := time.Duration(0)

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}

			err := f()
			errVal.Store(&errWrap{err})

			if err == nil {
				return
			}

			delay *= 2
			if delay == 0 {
				delay = initialBackoffDelay
			}

			if delay > maxBackoffDelay {
				delay = maxBackoffDelay
			}
		}
	}()

	select {
	case <-ctx.Done():
	case <-doneCh:
	}

	switch ctx.Err() {
	case context.Canceled:
		return fmt.Errorf("interrupted")
	case context.DeadlineExceeded:
		return fmt.Errorf("timeout")
	}

	if ev, ok := errVal.Load().(*errWrap); ok {
		return ev.E
	}

	return nil
}

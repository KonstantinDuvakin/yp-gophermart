package retry

import (
	"context"
	"time"
)

var delay = []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond}

func Do(ctx context.Context, isRetryable func(error) bool, handler func() error) error {
	err := handler()

	if err == nil || !isRetryable(err) {
		return err
	}

	for _, d := range delay {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d):
			err = handler()
			if err == nil || !isRetryable(err) {
				return err
			}
		}
	}

	return err
}

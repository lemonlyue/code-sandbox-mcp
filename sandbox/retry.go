package sandbox

import (
	"context"
	"fmt"
	"time"
)

type RetryableFunc func(ctx context.Context) error

// WithRetry Create a retry decorator
func WithRetry(maxAttempts int, delay time.Duration) func(retryableFunc RetryableFunc) RetryableFunc {
	return func(fn RetryableFunc) RetryableFunc {
		return func(ctx context.Context) error {
			var err error
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				err = fn(ctx)
				if err == nil {
					return nil
				}

				select {
				case <-time.After(delay):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return fmt.Errorf("after %d attempts, last error: %w", maxAttempts, err)
		}
	}
}

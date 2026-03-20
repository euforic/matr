package matr

import (
	"context"
	"fmt"
	"time"
)

const (
	timeoutEnvVar      = "MATR_TIMEOUT"
	defaultTaskTimeout = 5 * time.Minute
)

// ContextWithTimeoutValue builds the runner context from the configured timeout value.
// An empty value uses the default timeout and zero disables the deadline.
func ContextWithTimeoutValue(timeoutValue string) (context.Context, context.CancelFunc, error) {
	if timeoutValue == "" {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTaskTimeout)
		return ctx, cancel, nil
	}

	timeout, err := time.ParseDuration(timeoutValue)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid timeout %q: %w", timeoutValue, err)
	}

	if timeout < 0 {
		return nil, nil, fmt.Errorf("invalid timeout %q: must be >= 0", timeoutValue)
	}

	if timeout == 0 {
		ctx, cancel := context.WithCancel(context.Background())
		return ctx, cancel, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return ctx, cancel, nil
}

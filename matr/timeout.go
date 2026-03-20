package matr

import (
	"fmt"
	"os"
	"time"
)

const (
	timeoutEnvVar           = "MATR_TIMEOUT"
	defaultExecutionTimeout = 5 * time.Minute
)

// ExecutionTimeout returns the configured task timeout or the default.
func ExecutionTimeout() (time.Duration, error) {
	value := os.Getenv(timeoutEnvVar)
	if value == "" {
		return defaultExecutionTimeout, nil
	}

	timeout, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q: %w", timeoutEnvVar, value, err)
	}

	return timeout, nil
}

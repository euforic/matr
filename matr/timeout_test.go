package matr

import (
	"testing"
	"time"
)

func TestExecutionTimeoutDefaultsToFiveMinutes(t *testing.T) {
	t.Setenv(timeoutEnvVar, "")

	timeout, err := ExecutionTimeout()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if timeout != 5*time.Minute {
		t.Fatalf("expected 5m, got %s", timeout)
	}
}

func TestExecutionTimeoutUsesEnvOverride(t *testing.T) {
	t.Setenv(timeoutEnvVar, "30s")

	timeout, err := ExecutionTimeout()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if timeout != 30*time.Second {
		t.Fatalf("expected 30s, got %s", timeout)
	}
}

func TestExecutionTimeoutRejectsInvalidDuration(t *testing.T) {
	t.Setenv(timeoutEnvVar, "not-a-duration")

	_, err := ExecutionTimeout()
	if err == nil {
		t.Fatalf("expected invalid duration error")
	}
}

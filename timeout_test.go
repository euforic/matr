package matr

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestContextWithTimeoutValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		value        string
		wantErr      string
		wantDeadline bool
		minUntil     time.Duration
		maxUntil     time.Duration
	}{
		{
			name:         "default timeout",
			value:        "",
			wantDeadline: true,
			minUntil:     4 * time.Minute,
			maxUntil:     6 * time.Minute,
		},
		{
			name:         "custom timeout",
			value:        "30s",
			wantDeadline: true,
			minUntil:     29 * time.Second,
			maxUntil:     31 * time.Second,
		},
		{
			name:         "disabled timeout",
			value:        "0s",
			wantDeadline: false,
		},
		{
			name:    "invalid timeout",
			value:   "later",
			wantErr: "invalid timeout",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel, err := ContextWithTimeoutValue(tc.value)
			if cancel != nil {
				defer cancel()
			}

			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if ctx == nil {
				t.Fatal("expected context")
			}

			assertDeadline(t, ctx, tc.wantDeadline, tc.minUntil, tc.maxUntil)
		})
	}
}

func assertDeadline(t *testing.T, ctx context.Context, wantDeadline bool, minUntil, maxUntil time.Duration) {
	t.Helper()

	deadline, ok := ctx.Deadline()
	if ok != wantDeadline {
		t.Fatalf("deadline presence mismatch: want %v got %v", wantDeadline, ok)
	}

	if !wantDeadline {
		return
	}

	until := time.Until(deadline)
	if until < minUntil || until > maxUntil {
		t.Fatalf("deadline out of range: got %v want between %v and %v", until, minUntil, maxUntil)
	}
}

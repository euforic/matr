//go:build matr

package main

import (
	"context"
	"time"
)

// Sleepy waits for six minutes unless the matr execution context times out first.
func Sleepy(ctx context.Context, args []string) error {
	select {
	case <-time.After(6 * time.Minute):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

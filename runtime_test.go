package matr

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestRunParsesFlagsAndDependencies(t *testing.T) {
	t.Parallel()

	var order []string

	m := New()
	m.Handle(&Task{
		Name:      "proto",
		Summary:   "Generate code",
		DependsOn: []string{"test"},
		Handler: func(ctx context.Context, cmd *Invocation, args []string) error {
			order = append(order, "proto")
			if len(args) != 0 {
				t.Fatalf("proto args = %v, want none", args)
			}
			return nil
		},
	})
	m.Handle(&Task{
		Name:    "test",
		Summary: "Run tests",
		Handler: func(ctx context.Context, cmd *Invocation, args []string) error {
			order = append(order, "test")
			return nil
		},
	})
	m.Handle(&Task{
		Name:      "build",
		Summary:   "Build project",
		Aliases:   []string{"b"},
		DependsOn: []string{"proto", "test"},
		Flags: []Flag{
			{Name: "release", Type: FlagBool, Short: "r", Usage: "build with release settings"},
			{Name: "timeout", Type: FlagDuration, Default: "30s", Usage: "override timeout"},
			{Name: "token", Type: FlagString, Required: true, Usage: "API token"},
		},
		Handler: func(ctx context.Context, cmd *Invocation, args []string) error {
			order = append(order, "build")
			if !cmd.Bool("release") {
				t.Fatal("release flag not parsed")
			}
			if got := cmd.Duration("timeout"); got != 45*time.Second {
				t.Fatalf("timeout = %v, want %v", got, 45*time.Second)
			}
			if got := cmd.String("token"); got != "secret" {
				t.Fatalf("token = %q, want %q", got, "secret")
			}
			if !reflect.DeepEqual(args, []string{"target"}) {
				t.Fatalf("args = %v, want %v", args, []string{"target"})
			}
			return nil
		},
	})

	err := m.Run(context.Background(), "b", "--release", "--timeout=45s", "--token=secret", "target")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if got, want := order, []string{"test", "proto", "build"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("execution order = %v, want %v", got, want)
	}
}

func TestRunFailsForMissingRequiredFlag(t *testing.T) {
	t.Parallel()

	m := New()
	m.Handle(&Task{
		Name: "deploy",
		Flags: []Flag{
			{Name: "token", Type: FlagString, Required: true, Usage: "API token"},
		},
		Handler: func(ctx context.Context, cmd *Invocation, args []string) error {
			return nil
		},
	})

	err := m.Run(context.Background(), "deploy")
	if err == nil || err.Error() == "" {
		t.Fatalf("expected missing required flag error, got %v", err)
	}
}

func TestRunStopsOnDependencyError(t *testing.T) {
	t.Parallel()

	m := New()
	m.Handle(&Task{
		Name: "prep",
		Handler: func(ctx context.Context, cmd *Invocation, args []string) error {
			return errors.New("boom")
		},
	})
	m.Handle(&Task{
		Name:      "build",
		DependsOn: []string{"prep"},
		Handler: func(ctx context.Context, cmd *Invocation, args []string) error {
			t.Fatal("build should not run")
			return nil
		},
	})

	err := m.Run(context.Background(), "build")
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected dependency error, got %v", err)
	}
}

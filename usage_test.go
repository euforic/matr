package matr

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestPrintUsageIncludesAliasesDependenciesAndFlagMetadata(t *testing.T) {
	t.Parallel()

	m := New()
	var help bytes.Buffer
	var errs bytes.Buffer
	m.SetOutputs(&help, &errs)

	m.Handle(&Task{
		Name:      "build",
		Summary:   "Build the project",
		Doc:       "Build the project.",
		Aliases:   []string{"b"},
		DependsOn: []string{"test", "proto"},
		Flags: []Flag{
			{Name: "release", Type: FlagBool, Short: "r", Usage: "build with release settings"},
			{Name: "token", Type: FlagString, Required: true, Usage: "API token"},
			{Name: "timeout", Type: FlagDuration, Default: "30s", Usage: "override build timeout"},
		},
		Handler: func(ctx context.Context, cmd *Invocation, args []string) error {
			return nil
		},
	})

	m.PrintUsage("build")

	output := help.String()
	for _, want := range []string{
		"Aliases: b",
		"Depends on: test, proto",
		"--release, -r",
		"--token",
		"required",
		"default: 30s",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("help output missing %q:\n%s", want, output)
		}
	}
}

func TestRunInvokesDefaultTaskWithoutArgs(t *testing.T) {
	t.Parallel()

	called := false
	m := New()
	m.Handle(&Task{
		Name: "default",
		Handler: func(ctx context.Context, cmd *Invocation, args []string) error {
			called = true
			if len(args) != 0 {
				t.Fatalf("default args = %v, want none", args)
			}
			return nil
		},
	})

	if err := m.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !called {
		t.Fatal("default task was not called")
	}
}

func TestRunFailsOnDependencyCycle(t *testing.T) {
	t.Parallel()

	m := New()
	m.Handle(&Task{
		Name:      "build",
		DependsOn: []string{"test"},
		Handler: func(ctx context.Context, cmd *Invocation, args []string) error {
			return nil
		},
	})
	m.Handle(&Task{
		Name:      "test",
		DependsOn: []string{"build"},
		Handler: func(ctx context.Context, cmd *Invocation, args []string) error {
			return nil
		},
	})

	err := m.Run(context.Background(), "build")
	if err == nil || !strings.Contains(err.Error(), "dependency cycle") {
		t.Fatalf("expected dependency cycle error, got %v", err)
	}
}

func TestRunReturnsErrorForUnknownTask(t *testing.T) {
	t.Parallel()

	m := New()
	var help bytes.Buffer
	var errs bytes.Buffer
	m.SetOutputs(&help, &errs)

	err := m.Run(context.Background(), "missing")
	if err == nil || !strings.Contains(err.Error(), `no handler found for target "missing"`) {
		t.Fatalf("expected unknown task error, got %v", err)
	}
	if !strings.Contains(errs.String(), `ERROR: no handler found for target "missing"`) {
		t.Fatalf("stderr missing unknown task message:\n%s", errs.String())
	}
	if !strings.Contains(help.String(), "Targets:") {
		t.Fatalf("usage output missing command list:\n%s", help.String())
	}
}

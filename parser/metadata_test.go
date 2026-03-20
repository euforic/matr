package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseCommandMetadata(t *testing.T) {
	t.Parallel()

	path := writeTempGoFile(t, `//go:build matr

package main

import "context"

// Build compiles the project.
// It produces an artifact.
// @alias:b
// @depends:test,proto
// @release:bool,short=r;build with release settings
// @timeout:duration,short=t,default=30s;override build timeout
// @token:string,required;API token used for deploys
func Build(ctx context.Context, cmd any, args []string) error {
	return nil
}

// Default runs when no command is provided.
func Default(ctx context.Context, cmd any, args []string) error {
	return nil
}
`)

	cmds, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := []Command{
		{
			Name:       "Build",
			Summary:    "Build compiles the project.",
			Doc:        "Build compiles the project.\nIt produces an artifact.",
			IsExported: true,
			Aliases:    []string{"b"},
			DependsOn:  []string{"test", "proto"},
			Flags: []Flag{
				{Name: "release", Type: "bool", Short: "r", Usage: "build with release settings"},
				{Name: "timeout", Type: "duration", Short: "t", Default: "30s", Usage: "override build timeout"},
				{Name: "token", Type: "string", Required: true, Usage: "API token used for deploys"},
			},
		},
		{
			Name:       "Default",
			Summary:    "Default runs when no command is provided.",
			Doc:        "Default runs when no command is provided.",
			IsExported: true,
			IsDefault:  true,
		},
	}

	if diff := cmp.Diff(want, cmds); diff != "" {
		t.Fatalf("Parse mismatch (-want +got):\n%s", diff)
	}
}

func TestParseRejectsInvalidFlagMetadata(t *testing.T) {
	t.Parallel()

	path := writeTempGoFile(t, `//go:build matr

package main

import "context"

	// Build compiles the project.
	// @timeout:nope;invalid type
	func Build(ctx context.Context, cmd any, args []string) error {
		return nil
	}
`)

	_, err := Parse(path)
	if err == nil || !strings.Contains(err.Error(), "invalid flag type") {
		t.Fatalf("expected invalid flag type error, got %v", err)
	}
}

func TestParseRejectsInvalidDependencyMetadata(t *testing.T) {
	t.Parallel()

	path := writeTempGoFile(t, `//go:build matr

package main

import "context"

// Build compiles the project.
// @depends:test,
func Build(ctx context.Context, cmd any, args []string) error {
	return nil
}
`)

	_, err := Parse(path)
	if err == nil || !strings.Contains(err.Error(), "invalid dependency metadata") {
		t.Fatalf("expected invalid dependency metadata error, got %v", err)
	}
}

func TestParseRejectsInvalidFlagOption(t *testing.T) {
	t.Parallel()

	path := writeTempGoFile(t, `//go:build matr

package main

import "context"

// Build compiles the project.
// @timeout:duration,wat;invalid option
func Build(ctx context.Context, cmd any, args []string) error {
	return nil
}
`)

	_, err := Parse(path)
	if err == nil || !strings.Contains(err.Error(), "invalid flag option") {
		t.Fatalf("expected invalid flag option error, got %v", err)
	}
}

func TestParseReturnsErrorForInvalidGo(t *testing.T) {
	t.Parallel()

	path := writeTempGoFile(t, `//go:build matr

package main

func Build(`)

	_, err := Parse(path)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func writeTempGoFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "parser_metadata_test_*.go")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(tmpFile.Name())
	})

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	return tmpFile.Name()
}

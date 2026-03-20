package matr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunPassesTimeoutEnv(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	outputPath := filepath.Join(cacheDir, "timeout.txt")
	execPath := filepath.Join(cacheDir, "matr")

	script := "#!/bin/sh\nprintf '%s' \"$MATR_TIMEOUT\" > \"$1\"\n"
	if err := os.WriteFile(execPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	if err := run(cacheDir, 45*time.Second, outputPath); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	if string(got) != "45s" {
		t.Fatalf("timeout mismatch: got %q want %q", got, "45s")
	}
}

func TestBuildRebuildsWhenBinaryMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	matrfilePath := filepath.Join(dir, "Matrfile.go")
	content := `//go:build matr

package main

import (
	"context"

	"github.com/euforic/matr"
)

// Test runs the smoke task.
func Test(ctx context.Context, _ *matr.Invocation, args []string) error {
	return nil
}
`
	if err := os.WriteFile(matrfilePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write matrfile: %v", err)
	}

	cacheDir := filepath.Join(dir, defaultCacheFolder)
	if err := os.Mkdir(cacheDir, 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	hash, err := getSha256(matrfilePath)
	if err != nil {
		t.Fatalf("hash matrfile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "matrfile.sha256"), hash, 0o644); err != nil {
		t.Fatalf("write hash: %v", err)
	}

	outDir, err := build(matrfilePath, false)
	if err != nil {
		t.Fatalf("build returned error: %v", err)
	}

	if outDir != cacheDir {
		t.Fatalf("cache dir = %q, want %q", outDir, cacheDir)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "matr")); err != nil {
		t.Fatalf("expected compiled binary: %v", err)
	}
}

func TestHashFilesChangesWhenAnyInputChanges(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	first := filepath.Join(dir, "first.txt")
	second := filepath.Join(dir, "second.txt")

	if err := os.WriteFile(first, []byte("a"), 0o644); err != nil {
		t.Fatalf("write first: %v", err)
	}
	if err := os.WriteFile(second, []byte("b"), 0o644); err != nil {
		t.Fatalf("write second: %v", err)
	}

	one, err := hashFiles(first, second)
	if err != nil {
		t.Fatalf("hash files: %v", err)
	}

	if err := os.WriteFile(second, []byte("changed"), 0o644); err != nil {
		t.Fatalf("rewrite second: %v", err)
	}

	two, err := hashFiles(first, second)
	if err != nil {
		t.Fatalf("rehash files: %v", err)
	}

	if string(one) == string(two) {
		t.Fatal("expected hash to change when any input changes")
	}
}

func TestGetMatrfilePathResolvesDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	want := filepath.Join(dir, "Matrfile.go")
	if err := os.WriteFile(want, []byte("//go:build matr\npackage main\n"), 0o644); err != nil {
		t.Fatalf("write matrfile: %v", err)
	}

	got, err := getMatrfilePath(dir)
	if err != nil {
		t.Fatalf("getMatrfilePath returned error: %v", err)
	}
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGetMatrfilePathMissingReturnsError(t *testing.T) {
	t.Parallel()

	_, err := getMatrfilePath(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunReturnsErrorWhenGeneratedTaskFails(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	matrfilePath := filepath.Join(dir, "Matrfile.go")
	content := `//go:build matr

package main

import (
	"context"
	"errors"

	"github.com/euforic/matr"
)

// Fail always returns an error.
func Fail(ctx context.Context, _ *matr.Invocation, args []string) error {
	return errors.New("boom")
}
`
	if err := os.WriteFile(matrfilePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write matrfile: %v", err)
	}

	cacheDir, err := build(matrfilePath, true)
	if err != nil {
		t.Fatalf("build returned error: %v", err)
	}

	err = run(cacheDir, time.Second, "fail")
	if err == nil {
		t.Fatal("expected generated runner error")
	}
	if !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status error, got %v", err)
	}
}

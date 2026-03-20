//go:build matr

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/euforic/matr/internal/harness"
	"github.com/euforic/matr/matr"
)

// Setup installs module dependencies for local development.
func Setup(ctx context.Context, args []string) error {
	return runAndPrint("go mod download")
}

// Validate runs fast repository checks before the full test suite.
func Validate(ctx context.Context, args []string) error {
	if err := runAndPrint("go build ./..."); err != nil {
		return err
	}
	return runAndPrint("go test ./internal/harness ./parser")
}

// Test runs the full Go test suite for the repository.
func Test(ctx context.Context, args []string) error {
	return runAndPrint("go test -v ./...")
}

// Review runs the full local review path, including commit-message validation.
func Review(ctx context.Context, args []string) error {
	if err := Validate(ctx, args); err != nil {
		return err
	}
	if err := Test(ctx, args); err != nil {
		return err
	}
	return checkCommitSubject(args)
}

// Run demonstrates a basic user-defined task in the repository's own Matrfile.
func Run(ctx context.Context, args []string) error {
	fmt.Println("Running matr")
	return nil
}

func checkCommitSubject(args []string) error {
	subject := strings.TrimSpace(strings.Join(args, " "))
	if subject == "" {
		out, err := matr.Sh("git log -1 --pretty=%s HEAD").Output()
		if err != nil {
			return err
		}
		subject = strings.TrimSpace(string(out))
	}

	if err := harness.ValidateConventionalCommitSubject(subject); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "commit subject ok: %s\n", subject)
	return nil
}

func runAndPrint(cmd string) error {
	out, err := matr.Sh(cmd).CombinedOutput()
	if len(out) > 0 {
		fmt.Print(string(out))
	}
	if err != nil {
		return errors.New(strings.TrimSpace(string(out)))
	}
	return nil
}

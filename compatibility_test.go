package matr

import (
	"context"
	"testing"

	cliruntime "github.com/euforic/matr/cli"
)

func TestRootAPIRemainsCompatibleWithCLIPackage(t *testing.T) {
	t.Parallel()

	acceptRuntime(New())
	acceptTask(&Task{})
	acceptInvocation(&Invocation{})
	acceptFlagType(FlagBool)
	acceptHandler(nil)

	called := false
	m := New()
	m.Handle(&Task{
		Name: "default",
		Handler: func(ctx context.Context, cmd *Invocation, args []string) error {
			called = true
			return nil
		},
	})

	if err := m.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !called {
		t.Fatal("expected default handler to run")
	}
}

func acceptRuntime(*cliruntime.Matr) {}

func acceptTask(*cliruntime.Task) {}

func acceptInvocation(*cliruntime.Invocation) {}

func acceptFlagType(cliruntime.FlagType) {}

func acceptHandler(cliruntime.HandlerFunc) {}

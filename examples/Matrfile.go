// +build matr

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/matr-builder/matr/matr"
	"github.com/matr-builder/tlkn"
)

// Default is an example of overriding the default handler
func Default(ctx context.Context) (context.Context, error) {
	fmt.Println("Running Custom Default HandlerFunc...")
	Build(ctx)
	return ctx, nil
}

// Build is used as and example handler
func Build(ctx context.Context) (context.Context, error) {
	matr.Deps(ctx, Proto, Test)
	fmt.Println("Building...")

	return ctx, tlkn.Bash(`
		ls -la
		echo $GOPATH
		ls -l
	`)
}

// Run is used as and example handler
func Run(ctx context.Context) (context.Context, error) {
	matr.Deps(ctx, Build)
	fmt.Println("Running...")
	for {
	}
}

// notExported will run the project
func notExported(ctx context.Context) (context.Context, error) {
	fmt.Println("NotExported...")
	time.Sleep(1 * time.Second)
	return ctx, nil
}

// Proto will build the protobuf files into golang files
func Proto(ctx context.Context) (context.Context, error) {
	err := tlkn.Bash("echo \"build some proto file\"")
	return ctx, err
}

// Test is used as and example handler
func Test(ctx context.Context) (context.Context, error) {
	err := tlkn.Bash(`echo "Run unit tests..."`)
	time.Sleep(1 * time.Second)
	return ctx, err
}

// Bench is used as and example handler
func Bench(ctx context.Context) (context.Context, error) {
	return ctx, tlkn.Bash(`
		echo "Run benchmark......"
	`)
}

// Docker is used as and example handler
func Docker(ctx context.Context) (context.Context, error) {
	return ctx, tlkn.Bash(`
		echo "Build some docker file...."
	`)
}

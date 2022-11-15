//go:build matr

package main

import (
	"context"
	"fmt"

	"github.com/euforic/matr/matr"
)

func Test(ctx context.Context, args []string) error {
	out, err := matr.Sh("go test -v ./...").Output()
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}

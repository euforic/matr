package matr

import "github.com/euforic/matr/cli"

var Version = "v0.2.0"

type Matr = cli.Matr

// New creates a new CLI runtime instance.
func New() *Matr {
	return cli.New()
}

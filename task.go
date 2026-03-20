package matr

import "github.com/euforic/matr/cli"

type HandlerFunc = cli.HandlerFunc
type FlagType = cli.FlagType

const (
	FlagBool     = cli.FlagBool
	FlagString   = cli.FlagString
	FlagInt      = cli.FlagInt
	FlagDuration = cli.FlagDuration
)

type Task = cli.Task
type Flag = cli.Flag
type Invocation = cli.Invocation

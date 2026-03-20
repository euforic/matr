# matr

[![GoDoc](https://godoc.org/github.com/euforic/matr?status.svg)](https://godoc.org/github.com/euforic/matr)
![](https://img.shields.io/badge/license-MIT-blue.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/euforic/matr)](https://goreportcard.com/report/github.com/euforic/matr)

## Install

```bash
$ go install github.com/euforic/matr/cmd/matr@latest
```

## Usage

Execute a command

```bash
$ matr [target]
```

Show launcher flags

```bash
$ matr -h
Usage of matr:
  -clean
    	clean the matr cache
  -h	display usage info
  -matrfile string
    	path to Matrfile (default "./Matrfile.go")
  -no-cache
    	don't use the matr cache
  -timeout duration
    	task execution timeout; 0 disables the timeout (default 5m0s)
  -v	display version
```

Get expanded help for a command

```bash
$ matr build -h
```

Override the task timeout

```bash
$ matr -timeout 30s test
```

Disable the timeout entirely

```bash
$ matr -timeout 0 run
```

## Matrfile

A matr file is any regular go file. Matrfiles must be marked with a build target of "matr"
and be a package main file. The default Matrfile is `./Matrfile.go` or `./Matrfile` to avoid
conflicts. A custom Matrfile path can be defined with the `-matrfile` flag (example: `matr -matrfile /somepath/yourfile.go`)

Example Matrfile header

```go
//go:build matr

package main
```

## Targets

Any exported function that is `func(ctx context.Context, cmd *matr.Invocation, args []string) error` is considered a
matr target. If the function returns an error it will print to stdout and cause the matrfile
to exit with an exit with a non 0 exit code.

The generated runner now executes through `github.com/euforic/matr/cli`. The root package
continues to re-export the runtime types and constructor, so existing Matrfiles can keep using
`github.com/euforic/matr` unchanged. If you want to embed the command runtime directly, use the
`cli` package.

Comments on the target function will become documentation accessible by running
`matr <target> -h`. This will show the command docs, aliases, dependencies, and command-scoped
flags for that target.

A target may be designated the default target, which is run when the user runs
matr with no target specified. To denote the default, create a function named `Default`.
If no default target is specified, running `matr` with no target will print the list of targets
and docs.

## Command Metadata

Use comment metadata between the command docs and the function definition.

```go
// @alias:b
// @depends:test,proto
// @release:bool,short=r;build with release settings
// @timeout:duration,short=t,default=30s;override build timeout
// @token:string,required;API token used for deploys
```

Supported flag types:

- `bool`
- `string`
- `int`
- `duration`

Metadata rules:

- `@alias:<name>` adds a command alias.
- `@depends:<name[,name...]>` declares make-like dependencies.
- `@<flag>:<type>[,short=<char>][,default=<value>][,required];<description>` declares a local flag.
- `required` is optional; flags are optional by default.
- command flags are command-scoped, not global.
- after parsing command flags, `args` contains only positional arguments.

Dependencies are resolved as a serial DAG:

- each dependency runs at most once per invocation
- execution stops on the first dependency error
- dependency cycles fail with an explicit error

Use `cmd` inside handlers to read parsed flag values:

- `cmd.Bool(name)`
- `cmd.String(name)`
- `cmd.Int(name)`
- `cmd.Duration(name)`

Command help includes aliases, dependencies, defaults, and required flags. Example:

```bash
$ matr build -h
matr build:

Build compiles the project.

Aliases: b
Depends on: test

Flags:
  --release, -r    build with release settings
  --timeout, -t    override build timeout (default: 30s)
  --token          API token used for deploys (required)
```

## Example

```go
//go:build matr

package main

import (
    "context"
    "fmt"

    "github.com/euforic/matr"
)

// Build compiles the project.
// @alias:b
// @depends:test
// @release:bool,short=r;build with release settings
// @timeout:duration,short=t,default=30s;override build timeout
// @token:string,required;API token used for deploys
func Build(ctx context.Context, cmd *matr.Invocation, args []string) error {
    fmt.Println("release:", cmd.Bool("release"))
    fmt.Println("timeout:", cmd.Duration("timeout"))
    fmt.Println("token:", cmd.String("token"))
    fmt.Println("args:", args)
    return nil
}

// Test runs unit tests.
func Test(ctx context.Context, _ *matr.Invocation, args []string) error {
    return matr.Sh("go test ./...").Run()
}
```

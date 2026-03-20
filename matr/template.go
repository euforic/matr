package matr

import (
	"io"
	"strings"
	"text/template"

	"github.com/euforic/matr/parser"
	"golang.org/x/text/cases"
)

const defaultTemplate = `//go:build matr

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/euforic/matr/matr"
)

func main() {
	// Create new Matr instance
	m := matr.New()

	{{- range .}}
	{{if .IsExported }}
	// {{if .Summary}}{{.Summary}}{{else}}{{.Name}}{{end}}
	m.Handle(&matr.Task{
		Name: "{{cmdname .Name}}",
		Summary: "{{trim .Summary}}",
		Doc: ` + "`{{trim .Doc}}`," + `
		Handler: {{.Name}},
	})
	{{- end -}}
	{{- end}}

	timeout := 5 * time.Minute
	if value := os.Getenv("MATR_TIMEOUT"); value != "" {
		parsedTimeout, err := time.ParseDuration(value)
		if err != nil {
			os.Stderr.WriteString("ERROR: " + fmt.Sprintf("invalid MATR_TIMEOUT value %q: %v", value, err) + "\n")
			os.Exit(1)
		}
		timeout = parsedTimeout
	}

	// Setup context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Setup signal handling for SIGINT and SIGTERM
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Run Matr in a separate goroutine
	errChan := make(chan error)
	go func() {
		errChan <- m.Run(ctx, os.Args[1:]...)
	}()

	// Wait for Matr to finish, a timeout, or a signal
	select {
	case err := <-errChan:
		if err != nil {
			os.Stderr.WriteString("ERROR: " + err.Error() + "\n")
		}
	case <-ctx.Done():
		os.Stderr.WriteString("ERROR: Context timed out\n")
	case <-sig:
		cancel()
		os.Stderr.WriteString("INFO: Received signal, shutting down\n")
	}
}`

// generate ...
func generate(cmds []parser.Command, w io.Writer) error {
	// Create a new template and parse the letter into it.
	t := template.Must(template.New("matr").Funcs(template.FuncMap{
		"title": cases.Title,
		"trim":  strings.TrimSpace,
		"cmdname": func(name string) string {
			return parser.LowerFirst(parser.CamelToHyphen(name))
		},
	}).Parse(defaultTemplate))
	return t.Execute(w, cmds)
}

package matr

import (
	"io"
	"strings"
	"text/template"

	"github.com/euforic/matr/parser"
)

const defaultTemplate = `//go:build matr

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/euforic/matr"
)

func main() {
	m := matr.New()

	{{- range .}}
	{{if .IsExported }}
	m.Handle(&matr.Task{
		Name: "{{cmdname .Name}}",
		Summary: "{{trim .Summary}}",
		Doc: ` + "`{{trim .Doc}}`," + `
		Aliases: []string{ {{- range $i, $a := .Aliases}}{{if $i}}, {{end}}"{{$a}}"{{end}} },
		DependsOn: []string{ {{- range $i, $d := .DependsOn}}{{if $i}}, {{end}}"{{$d}}"{{end}} },
		Flags: []matr.Flag{
			{{- range .Flags }}
			{
				Name: {{ printf "%q" .Name }},
				Type: {{ flagType .Type }},
				Short: {{ printf "%q" .Short }},
				Default: {{ printf "%q" .Default }},
				Required: {{ .Required }},
				Usage: {{ printf "%q" .Usage }},
			},
			{{- end }}
		},
		Handler: {{.Name}},
	})
	{{- end -}}
	{{- end}}

	ctx, cancel, err := matr.ContextWithTimeoutValue(os.Getenv("` + timeoutEnvVar + `"))
	if err != nil {
		_, _ = os.Stderr.WriteString("ERROR: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error)
	go func() {
		errChan <- m.Run(ctx, os.Args[1:]...)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			_, _ = os.Stderr.WriteString("ERROR: " + err.Error() + "\n")
			os.Exit(1)
		}
	case <-ctx.Done():
		_, _ = os.Stderr.WriteString("ERROR: Context timed out\n")
		os.Exit(1)
	case <-sig:
		cancel()
		_, _ = os.Stderr.WriteString("INFO: Received signal, shutting down\n")
		os.Exit(130)
	}
}`

func generate(cmds []parser.Command, w io.Writer) error {
	t := template.Must(template.New("matr").Funcs(template.FuncMap{
		"trim": strings.TrimSpace,
		"cmdname": func(name string) string {
			return parser.LowerFirst(parser.CamelToHyphen(name))
		},
		"flagType": func(name string) string {
			switch name {
			case "bool":
				return "matr.FlagBool"
			case "string":
				return "matr.FlagString"
			case "int":
				return "matr.FlagInt"
			case "duration":
				return "matr.FlagDuration"
			default:
				return `matr.FlagType("` + name + `")`
			}
		},
	}).Parse(defaultTemplate))
	return t.Execute(w, cmds)
}

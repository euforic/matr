package matr

import (
	"bytes"
	"strings"
	"testing"

	"github.com/euforic/matr/parser"
)

func TestGenerateUsesCLIPackage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := generate([]parser.Command{
		{
			Name:       "Build",
			Summary:    "Build the project.",
			Doc:        "Build the project.",
			IsExported: true,
			Flags: []parser.Flag{
				{Name: "release", Type: "bool", Usage: "build with release settings"},
			},
		},
	}, &buf)
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		`"github.com/euforic/matr/cli"`,
		"m := cli.New()",
		"&cli.Task{",
		"[]cli.Flag{",
		"Type: cli.FlagBool",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("generated output missing %q:\n%s", want, output)
		}
	}
}

package parser

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Command defines a matr cmd
type Command struct {
	Name       string
	Summary    string
	Doc        string
	IsExported bool
}

// Param defines a matr HandlerFunc parameter
type Param struct {
	Name string
	Type string
}

// Parse parses a matr file and returns a list of commands
func Parse(file string) ([]Command, error) {
	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, file, nil, 4)
	if err != nil {
		panic(err)
	}

	funcs := []Command{}
	if len(f.Comments) == 0 ||
		f.Comments[0].Pos() != 1 ||
		len(f.Comments[0].List) == 0 ||
		f.Comments[0].List[0].Text != "//go:build matr" {
		return funcs, errors.New("invalid Matrfile: matr build tag missing or incorrect")
	}

	for _, d := range f.Decls {
		t, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		funcs = append(funcs, parseCmd(t))
	}

	return funcs, nil
}

func parseCmd(t *ast.FuncDecl) Command {
	cmd := Command{
		Name:       t.Name.String(),
		IsExported: ast.IsExported(t.Name.String()),
	}

	if t.Doc != nil && len(t.Doc.List) > 0 {
		d := []string{}
		for _, ds := range t.Doc.List {
			d = append(d, strings.Replace(ds.Text, "//", "", 1))
		}

		cmd.Summary = strings.TrimLeft(d[0], " ")
		cmd.Doc = strings.TrimLeft(strings.Join(d, "\n"), " ")
	}

	return cmd
}

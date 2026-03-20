package parser

import (
	"errors"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"slices"
	"strings"
)

// Command defines a parsed matr command.
type Command struct {
	Name       string
	Summary    string
	Doc        string
	IsExported bool
	IsDefault  bool
	Aliases    []string
	DependsOn  []string
	Flags      []Flag
}

// Flag describes a command-scoped flag extracted from Matrfile comments.
type Flag struct {
	Name     string
	Type     string
	Short    string
	Default  string
	Required bool
	Usage    string
}

var validFlagTypes = []string{"bool", "string", "int", "duration"}

// Parse parses a matr file and returns a list of commands.
func Parse(file string) ([]Command, error) {
	cmds := []Command{}
	fset := token.NewFileSet()
	f, err := goparser.ParseFile(fset, file, nil, goparser.ParseComments)
	if err != nil {
		return cmds, err
	}

	if len(f.Comments) == 0 ||
		f.Comments[0].Pos() != 1 ||
		len(f.Comments[0].List) == 0 ||
		f.Comments[0].List[0].Text != "//go:build matr" {
		return cmds, errors.New("invalid Matrfile: matr build tag missing or incorrect")
	}

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		cmd, err := parseCmd(fn)
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

func parseCmd(fn *ast.FuncDecl) (Command, error) {
	cmd := Command{
		Name:       fn.Name.Name,
		IsExported: ast.IsExported(fn.Name.Name),
		IsDefault:  fn.Name.Name == "Default",
	}

	if fn.Doc == nil {
		return cmd, nil
	}

	var docLines []string
	for _, comment := range fn.Doc.List {
		line := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "@") {
			if err := applyMetadata(&cmd, line); err != nil {
				return Command{}, fmt.Errorf("%s: %w", cmd.Name, err)
			}
			continue
		}
		docLines = append(docLines, line)
	}

	if len(docLines) > 0 {
		cmd.Summary = docLines[0]
		cmd.Doc = strings.Join(docLines, "\n")
	}

	return cmd, nil
}

func applyMetadata(cmd *Command, line string) error {
	switch {
	case strings.HasPrefix(line, "@alias:"):
		alias := strings.TrimSpace(strings.TrimPrefix(line, "@alias:"))
		if alias == "" {
			return errors.New("invalid alias metadata")
		}
		cmd.Aliases = append(cmd.Aliases, alias)
		return nil
	case strings.HasPrefix(line, "@depends:"):
		raw := strings.TrimSpace(strings.TrimPrefix(line, "@depends:"))
		if raw == "" {
			return errors.New("invalid dependency metadata")
		}
		for dep := range strings.SplitSeq(raw, ",") {
			dep = strings.TrimSpace(dep)
			if dep == "" {
				return errors.New("invalid dependency metadata")
			}
			cmd.DependsOn = append(cmd.DependsOn, dep)
		}
		return nil
	default:
		flag, err := parseFlag(line)
		if err != nil {
			return err
		}
		cmd.Flags = append(cmd.Flags, flag)
		return nil
	}
}

func parseFlag(line string) (Flag, error) {
	meta, usage, _ := strings.Cut(strings.TrimPrefix(line, "@"), ";")
	name, rest, ok := strings.Cut(meta, ":")
	if !ok || strings.TrimSpace(name) == "" {
		return Flag{}, errors.New("invalid flag metadata")
	}

	parts := strings.Split(rest, ",")
	flag := Flag{
		Name:  strings.TrimSpace(name),
		Type:  strings.TrimSpace(parts[0]),
		Usage: strings.TrimSpace(usage),
	}
	if !slices.Contains(validFlagTypes, flag.Type) {
		return Flag{}, fmt.Errorf("invalid flag type %q", flag.Type)
	}

	for _, raw := range parts[1:] {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		switch {
		case raw == "required":
			flag.Required = true
		case strings.HasPrefix(raw, "short="):
			flag.Short = strings.TrimSpace(strings.TrimPrefix(raw, "short="))
			if flag.Short == "" {
				return Flag{}, errors.New("invalid short flag metadata")
			}
			if len(flag.Short) != 1 {
				return Flag{}, errors.New("short flag must be a single character")
			}
			if flag.Short[0] == '-' {
				return Flag{}, errors.New("short flag cannot start with '-'")
			}
		case strings.HasPrefix(raw, "default="):
			flag.Default = strings.TrimSpace(strings.TrimPrefix(raw, "default="))
		default:
			return Flag{}, fmt.Errorf("invalid flag option %q", raw)
		}
	}

	return flag, nil
}

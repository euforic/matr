package matr

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

var Version = "v0.2.0"

// Matr is the root structure.
type Matr struct {
	tasks    map[string]*Task
	aliases  map[string]string
	order    []string
	onExit   func(context.Context, error)
	helpOut  io.Writer
	errorOut io.Writer
}

// New creates a new Matr struct instance and returns a point to it.
func New() *Matr {
	return &Matr{
		tasks:    map[string]*Task{},
		aliases:  map[string]string{},
		helpOut:  os.Stdout,
		errorOut: os.Stderr,
	}
}

// TaskNames returns the available task names in deterministic order.
func (m *Matr) TaskNames() []string {
	names := append([]string(nil), m.order...)
	sort.Strings(names)
	return names
}

// PrintUsage outputs usage docs to stdout.
func (m *Matr) PrintUsage(cmd string) {
	var err error

	if cmd != "" {
		task, ok := m.lookupTask(cmd)
		if ok {
			_, _ = fmt.Fprintf(m.helpOut, "matr %s:\n\n", task.Name)
			_, _ = fmt.Fprintln(m.helpOut, task.Doc)
			if len(task.Aliases) > 0 {
				_, _ = fmt.Fprintf(m.helpOut, "\nAliases: %s\n", strings.Join(task.Aliases, ", "))
			}
			if len(task.DependsOn) > 0 {
				_, _ = fmt.Fprintf(m.helpOut, "Depends on: %s\n", strings.Join(task.DependsOn, ", "))
			}
			if len(task.Flags) > 0 {
				_, _ = fmt.Fprintln(m.helpOut, "\nFlags:")
				for _, fl := range task.Flags {
					_, _ = fmt.Fprintf(m.helpOut, "  --%s", fl.Name)
					if fl.Short != "" {
						_, _ = fmt.Fprintf(m.helpOut, ", -%s", fl.Short)
					}
					if fl.Usage != "" {
						_, _ = fmt.Fprintf(m.helpOut, "\t%s", fl.Usage)
					}
					var extras []string
					if fl.Required {
						extras = append(extras, "required")
					}
					if fl.Default != "" {
						extras = append(extras, "default: "+fl.Default)
					}
					if len(extras) > 0 {
						_, _ = fmt.Fprintf(m.helpOut, " (%s)", strings.Join(extras, ", "))
					}
					_, _ = fmt.Fprintln(m.helpOut)
				}
			}
			_, _ = fmt.Fprintln(m.helpOut)
			return
		}
		err = fmt.Errorf("ERROR: no handler found for target %q", cmd)
	}

	_, _ = fmt.Fprintln(m.helpOut, "\nRun Task: matr <opts> [target] [command flags] args...")
	_, _ = fmt.Fprintln(m.helpOut, "\nTargets:")
	tw := tabwriter.NewWriter(m.helpOut, 0, 0, 3, ' ', 0)
	for _, name := range m.TaskNames() {
		task := m.tasks[name]
		if task.Name == "default" {
			continue
		}
		_, _ = fmt.Fprintf(tw, "\t%s\t%s\n", task.Name, task.Summary)
	}
	_ = tw.Flush()
	_, _ = fmt.Fprintln(m.helpOut, " ")
	if err != nil {
		_, _ = fmt.Fprintln(m.errorOut, err.Error())
	}
}

// Handle registers a new task handler with matr.
func (m *Matr) Handle(task *Task) {
	if task.Name == "" {
		task.Name = "default"
	}
	key := strings.ToLower(task.Name)
	if _, exists := m.tasks[key]; !exists {
		m.order = append(m.order, key)
	}
	m.tasks[key] = task
	for _, alias := range task.Aliases {
		m.aliases[strings.ToLower(alias)] = key
	}
}

// Run executes the requested task function with the provided context and arguments.
func (m *Matr) Run(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		if task, ok := m.lookupTask("default"); ok {
			return m.runTask(ctx, task, nil, map[string]bool{}, map[string]bool{})
		}
		m.PrintUsage("")
		return nil
	}

	if isHelpArg(args[0]) {
		m.PrintUsage("")
		return nil
	}

	task, ok := m.lookupTask(args[0])
	if !ok {
		_, _ = fmt.Fprintf(m.errorOut, "ERROR: no handler found for target %q\n", args[0])
		m.PrintUsage("")
		return nil
	}

	if len(args) > 1 && isHelpArg(args[1]) {
		m.PrintUsage(task.Name)
		return nil
	}

	return m.runTask(ctx, task, args[1:], map[string]bool{}, map[string]bool{})
}

func (m *Matr) runTask(ctx context.Context, task *Task, argv []string, ran, visiting map[string]bool) error {
	key := strings.ToLower(task.Name)
	if ran[key] {
		return nil
	}
	if visiting[key] {
		return fmt.Errorf("dependency cycle detected at %q", task.Name)
	}
	visiting[key] = true

	for _, depName := range task.DependsOn {
		dep, ok := m.lookupTask(depName)
		if !ok {
			return fmt.Errorf("unknown dependency %q for %q", depName, task.Name)
		}
		if err := m.runTask(ctx, dep, nil, ran, visiting); err != nil {
			return err
		}
	}

	delete(visiting, key)

	invocation, err := newInvocation(task, argv)
	if err != nil {
		return err
	}

	err = task.Handler(ctx, invocation, invocation.Args())
	ran[key] = true
	if m.onExit != nil {
		m.onExit(ctx, err)
	}
	return err
}

func newInvocation(task *Task, argv []string) (*Invocation, error) {
	fs := flag.NewFlagSet(task.Name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	values := make(map[string]any, len(task.Flags))
	postParse := make([]func(), 0, len(task.Flags))
	for _, fl := range task.Flags {
		after, err := registerFlag(fs, values, fl)
		if err != nil {
			return nil, err
		}
		postParse = append(postParse, after)
	}

	if err := fs.Parse(argv); err != nil {
		return nil, err
	}
	for _, fn := range postParse {
		fn()
	}

	for _, fl := range task.Flags {
		if fl.Required && isZeroFlagValue(values[fl.Name]) {
			return nil, fmt.Errorf("missing required flag --%s", fl.Name)
		}
	}

	return &Invocation{
		task:   task,
		values: values,
		args:   fs.Args(),
	}, nil
}

func registerFlag(fs *flag.FlagSet, values map[string]any, fl Flag) (func(), error) {
	switch fl.Type {
	case FlagBool:
		def := false
		if fl.Default != "" {
			parsed, err := strconv.ParseBool(fl.Default)
			if err != nil {
				return nil, err
			}
			def = parsed
		}
		ptr := new(bool)
		*ptr = def
		fs.BoolVar(ptr, fl.Name, def, fl.Usage)
		if fl.Short != "" {
			fs.BoolVar(ptr, fl.Short, def, fl.Usage)
		}
		return func() { values[fl.Name] = *ptr }, nil
	case FlagString:
		ptr := new(string)
		*ptr = fl.Default
		fs.StringVar(ptr, fl.Name, fl.Default, fl.Usage)
		if fl.Short != "" {
			fs.StringVar(ptr, fl.Short, fl.Default, fl.Usage)
		}
		return func() { values[fl.Name] = *ptr }, nil
	case FlagInt:
		def := 0
		if fl.Default != "" {
			parsed, err := strconv.Atoi(fl.Default)
			if err != nil {
				return nil, err
			}
			def = parsed
		}
		ptr := new(int)
		*ptr = def
		fs.IntVar(ptr, fl.Name, def, fl.Usage)
		if fl.Short != "" {
			fs.IntVar(ptr, fl.Short, def, fl.Usage)
		}
		return func() { values[fl.Name] = *ptr }, nil
	case FlagDuration:
		def := time.Duration(0)
		if fl.Default != "" {
			parsed, err := time.ParseDuration(fl.Default)
			if err != nil {
				return nil, err
			}
			def = parsed
		}
		ptr := new(time.Duration)
		*ptr = def
		fs.DurationVar(ptr, fl.Name, def, fl.Usage)
		if fl.Short != "" {
			fs.DurationVar(ptr, fl.Short, def, fl.Usage)
		}
		return func() { values[fl.Name] = *ptr }, nil
	default:
		return nil, fmt.Errorf("unsupported flag type %q", fl.Type)
	}
}

func isZeroFlagValue(v any) bool {
	switch value := v.(type) {
	case bool:
		return !value
	case string:
		return value == ""
	case int:
		return value == 0
	case time.Duration:
		return value == 0
	default:
		return v == nil
	}
}

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help"
}

func (m *Matr) lookupTask(name string) (*Task, bool) {
	key := strings.ToLower(name)
	if target, ok := m.aliases[key]; ok {
		key = target
	}
	task, ok := m.tasks[key]
	return task, ok
}

// OnExit executes a final function before matr exits.
func (m *Matr) OnExit(fn func(ctx context.Context, err error)) {
	m.onExit = fn
}

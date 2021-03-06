package matr

import (
	"context"
	"errors"
)

// Matr is the root structure
type Matr struct {
	tasks  map[string]Task
	onExit func(context.Context, error)
}

// New creates a new Matr struct instance and returns a point to it
func New() *Matr {
	return &Matr{
		tasks: map[string]Task{},
	}
}

// TaskNames returns an []string of the available task names
func (m *Matr) TaskNames() []string {
	names := []string{}
	for n := range m.tasks {
		names = append(names, n)
	}
	return names
}

// Handle registers a new task handler with matr. The Handler will then be referenceable by the provided name,
// if a task is named "default" or "" that function will be called if no function name is provided. The
// default function is also a good place to output usage information for the available tasks.
// CallOptions can be used to allow for before and after Handler middleware functions.
func (m *Matr) Handle(name string, fn HandlerFunc) Task {
	if name == "" {
		name = "default"
	}
	task := Task{
		Name:    name,
		Handler: fn,
	}
	m.tasks[name] = task
	return m.tasks[name]
}

// Run will execute the requested task function with the provided context and arguments.
func (m *Matr) Run(ctx context.Context, cmd ...string) error {
	var taskName string

	if len(cmd) > 0 {
		taskName = cmd[0]
	}

	ctx, err := m.execTask(ctx, taskName)
	if m.onExit != nil {
		m.onExit(ctx, err)
	}

	return err
}

// OnExit executes a final function before matr exits
func (m *Matr) OnExit(fn func(ctx context.Context, err error)) {
	m.onExit = fn
	return
}

func (m *Matr) execTask(ctx context.Context, name string) (context.Context, error) {
	var err error
	if name == "" {
		name = "default"
	}

	task, ok := m.tasks[name]
	if !ok {
		t, ok := m.tasks["default"]
		if !ok {
			return ctx, errors.New("No Default handler defined")
		}
		return t.Handler(ctx)
	}

	ctx, err = task.Handler(ctx)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

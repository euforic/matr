package matr

import (
	"context"
	"time"
)

// Task struct that holds registered handler.
type Task struct {
	Name      string
	Summary   string
	Doc       string
	Aliases   []string
	DependsOn []string
	Flags     []Flag
	Handler   HandlerFunc
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as a matr task Handler.
type HandlerFunc func(c context.Context, cmd *Invocation, args []string) error

type FlagType string

const (
	FlagBool     FlagType = "bool"
	FlagString   FlagType = "string"
	FlagInt      FlagType = "int"
	FlagDuration FlagType = "duration"
)

type Flag struct {
	Name     string
	Type     FlagType
	Short    string
	Default  string
	Required bool
	Usage    string
}

type Invocation struct {
	task   *Task
	values map[string]any
	args   []string
}

func (i *Invocation) Name() string {
	if i == nil || i.task == nil {
		return ""
	}
	return i.task.Name
}

func (i *Invocation) Args() []string {
	if i == nil {
		return nil
	}
	return append([]string(nil), i.args...)
}

func (i *Invocation) Bool(name string) bool {
	v, _ := i.values[name].(bool)
	return v
}

func (i *Invocation) String(name string) string {
	v, _ := i.values[name].(string)
	return v
}

func (i *Invocation) Int(name string) int {
	v, _ := i.values[name].(int)
	return v
}

func (i *Invocation) Duration(name string) time.Duration {
	v, _ := i.values[name].(time.Duration)
	return v
}

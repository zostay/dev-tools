package plugin

import (
	"context"
	"fmt"
	"strings"
)

var ErrUnsupportedTask = fmt.Errorf("this plugin does not support that task")

type Ordering int

type OperationFunc func(context.Context) error

func (op OperationFunc) Call(ctx context.Context) error {
	return op(ctx)
}

type OperationHandler interface {
	Call(ctx context.Context) error
}

type Operations []Operation

type Operation struct {
	Order  Ordering
	Action OperationHandler
}

// Task provides some operations to help perform a task. The task is executed
// in a series of stages and if multiple plugins implement a given task, they
// may be run in parallel. The task operations are executed in the following
// order:
// * Setup
// * Check
// * Begin (in ascending Order)
// * Run (in ascending Order)
// * End (in ascending Order)
// * Finishing
// * Teardown
type Task interface {
	Setup(context.Context) error
	Check(context.Context) error
	Begin(context.Context) (Operations, error)
	Run(context.Context) (Operations, error)
	End(context.Context) (Operations, error)
	Finishing(context.Context) error
	Teardown(context.Context) error
}

type Tasks []Task

type Config struct {
	Values      map[string]string
	SubSections map[string]*Config
}

func (c *Config) Get(key string) string {
	parts := strings.SplitN(key, ".", 2)
	thisKey := parts[0]
	if len(parts) == 1 {
		return c.Values[thisKey]
	}
	return c.SubSections[thisKey].Get(parts[1])
}

// TaskInterface is the base interface that all plugins implement.
type TaskInterface interface {
	// Implements will list the names of the tasks that this plugin TaskInterface
	// implements.
	Implements() (taskNames []string, err error)

	// Prepare should return an initialized Task object that is configured using
	// the given global configuration as well as the task configuration. The
	// object passed for task configuration is specific to the given taskName.
	Prepare(taskName string, globalCfg *Config) (task Task, err error)
}

type Boilerplate struct{}

func (Boilerplate) Setup(context.Context) error               { return nil }
func (Boilerplate) Check(context.Context) error               { return nil }
func (Boilerplate) Begin(context.Context) (Operations, error) { return nil, nil }
func (Boilerplate) Run(context.Context) (Operations, error)   { return nil, nil }
func (Boilerplate) End(context.Context) (Operations, error)   { return nil, nil }
func (Boilerplate) Finishing(context.Context) error           { return nil }
func (Boilerplate) Teardown(context.Context) error            { return nil }

type CommandInterface interface {
	List() []CommandDescriptor
	Run(tag string, args []string, opts map[string]string) int
}

type CommandDescriptor struct {
	Tag               string
	Parents           []string
	Use               string
	Short             string
	MinPositionalArgs int
	MaxPositionalargs int
	Options           []OptionDescriptor
}

type OptionType int

const (
	OptionBool OptionType = iota + 1
	OptionaString
)

type OptionDescriptor struct {
	Name      string
	Shorthand string
	Usage     string
	Default   string
	ValueType OptionType
}

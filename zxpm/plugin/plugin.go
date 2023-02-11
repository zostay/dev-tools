package plugin

import (
	"context"

	"github.com/zostay/dev-tools/pkg/config"
)

type Ordering int
type OperationFunc func(ctx context.Context) error

type Operations []Operation

type Operation struct {
	Order  Ordering
	Action OperationFunc
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
	Begin() Operations
	Run() Operations
	End() Operations
	Finishing(context.Context) error
	Teardown(context.Context) error
}

// Interface is the base interface that all plugins implement.
type Interface interface {
	// Implements will list the names of the tasks that this plugin Interface
	// implements.
	Implements() (taskNames []string)

	// Prepare should return an initialized Task object that is configured using
	// the given global configuration as well as the task configuration. The
	// object passed for task configuration is specific to the given taskName.
	Prepare(taskName string, cfg *config.Config, taskCfg any) (task Task)
}

type Boilerplate struct{}

func (Boilerplate) Setup(context.Context) error     { return nil }
func (Boilerplate) Check(context.Context) error     { return nil }
func (Boilerplate) Begin() Operations               { return nil }
func (Boilerplate) Run() Operations                 { return nil }
func (Boilerplate) End() Operations                 { return nil }
func (Boilerplate) Finishing(context.Context) error { return nil }
func (Boilerplate) Teardown(context.Context) error  { return nil }

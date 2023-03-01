package plugin

import (
	"context"
)

// Task provides some operations to help perform a task. The task is executed
// in a series of stages and if multiple plugins implement a given task, they
// may be run in parallel. The task operations are executed in the following
// order:
// * Setup
// * Check
// * Begin (in ascending Order)
// * Run (in ascending Order)
// * End (in ascending Order)
// * Finish
// * Teardown
type Task interface {
	Setup(context.Context) error
	Check(context.Context) error
	Begin(context.Context) (Operations, error)
	Run(context.Context) (Operations, error)
	End(context.Context) (Operations, error)
	Finish(context.Context) error
	Teardown(context.Context) error
}

type Tasks []Task

type TaskBoilerplate struct{}

func (TaskBoilerplate) Setup(context.Context) error               { return nil }
func (TaskBoilerplate) Check(context.Context) error               { return nil }
func (TaskBoilerplate) Begin(context.Context) (Operations, error) { return nil, nil }
func (TaskBoilerplate) Run(context.Context) (Operations, error)   { return nil, nil }
func (TaskBoilerplate) End(context.Context) (Operations, error)   { return nil, nil }
func (TaskBoilerplate) Finish(context.Context) error              { return nil }
func (TaskBoilerplate) Teardown(context.Context) error            { return nil }

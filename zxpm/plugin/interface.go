package plugin

import (
	"context"
	"fmt"
	"strings"
)

var (
	// ErrUnsupportedTask is returned by Interface.Prepare when the named
	// task is not implemented by the plugin.
	ErrUnsupportedTask = fmt.Errorf("this plugin does not support that task")

	// ErrUnsupportedGoal is returned by Interface.Goal when the named goal
	// is not defined by the plugin.
	ErrUnsupportedGoal = fmt.Errorf("this plugin does not support that goal")

	// ErrBadTaskName is returned when a badly formatted task name is detected.
	ErrBadTaskName = fmt.Errorf("the task name is badly formatted")
)

// GoalDescription describes a top-level goal.
type GoalDescription interface {
	// Name is the top-level name of the task. This is preferable a single,
	// short verb. It must be all lowercase letters or may contain a hyphen.
	// This will be the default command-name to use to execute this task and
	// associated sub-tasks and also must match the first path element of a
	// SubTaskDescription Name.
	Name() string

	// Short is the short description of the task to show the user when showing
	// usage information for the task.
	Short() string

	// Aliases returns other names (possibly shortened names) to grant
	// implementations of this task.
	Aliases() []string
}

// TaskDescription describes a sub-task.
type TaskDescription interface {
	// Name is a path starting with a leading slash naming the sub-task. The
	// first element of this path must match a TaskDescription defined by the
	// given Plugin (or a built-in task). The remaining path elements define a
	// sub-task of the level above. This is usually of the form of /verb/noun
	// where short words a preferred and the names must be lowercase and may
	// contain hyphens.
	//
	// The path elements here will become the names of the sub-commands on the
	// command-line if a user wants to execute a sub-step of a larger task.
	Name() string

	// Short is the short description of what this sub-task does. It will be
	// combined with all other like-named sub-tasks to from the description text
	// shown when usage help is requested. This description should be very
	// concise.
	Short() string

	// Requires names zero or more sub-task names on which this task depends.
	// Usually, these will be other sub-tasks that are defined by this plugin.
	//
	// Actual sub-task dependencies are calculated by calculating the
	// requirements of all like-named sub-tasks together. For example, if
	// /release/publish in plugin A depends on /release/mint and it dapends on
	// /release/wait in plugin B and both plugins will be executed, then this
	// task will not be executed until after both /release/wait and
	// /release/mint have been executed.
	Requires() []string
}

func GoalName(task TaskDescription) (string, error) {
	name := task.Name()
	if name[0] != '/' {
		return "", fmt.Errorf("%s: %w", task.Name(), ErrBadTaskName)
	}

	name = name[1:]

	idx := strings.IndexRune(name, '/')
	if idx < 0 {
		return "", fmt.Errorf("%s: %w", task.Name(), ErrBadTaskName)
	}

	return name[:idx], nil
}

// Interface is the base interface that all plugins implement.
type Interface interface {
	// Implements will list the tasks that this plugin implements. It may return
	// an empty list if no task is implemented.
	Implements(ctx context.Context) (tasks []TaskDescription, err error)

	// Goal will return the GoalDefinition for the given goal name, which
	// provides documentation for the named top-level goal. This method should
	// return ErrUnsupportedGoal when no such top-level task is defined by this
	// plugin.
	Goal(ctx context.Context, name string) (def GoalDescription, err error)

	// Prepare should return an initialized Task object that is configured using
	// the given global configuration as well as the task configuration. The
	// object passed for task configuration is specific to the given taskName.
	//
	// The lifecycle of this method needs to be handled such that the return
	// value from this method must be handled similar to the following:
	//
	//  task, err := taskInterface.Prepare(ctx, taskName, globalCfg)
	//  if err != nil {
	//    if task != nil {
	//      cancelErr := taskInterface.Cancel(ctx, task)
	//      if cancelErr != nil {
	//        log.Print(cancelErr)
	//      }
	//    }
	//    return err
	//  }
	//
	// Once a task is returned from this method, you must either call Cancel
	// or Close on the Interface with the given task object or risk leaking
	// resources or having other unfinished working and unexpected results.
	//
	// This should return ErrUnsupportedTask if the given taskName does not
	// match a supported task.
	Prepare(
		ctx context.Context,
		taskName string,
	) (task Task, err error)

	// Cancel must be called when a task is not going to be completed in full.
	// The implementation will use this opportunity to call teardown and perform
	// any cleanup actions that have been queued up to try to undo the work done
	// so far.
	Cancel(ctx context.Context, task Task) (err error)

	// Complete must be called when a task has been run to completion. This
	// allows the task to perform any final teardown and cleanup resources.
	Complete(ctx context.Context, task Task) (err error)
}

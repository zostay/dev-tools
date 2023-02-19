package plugin

import (
	"context"
	"fmt"

	"github.com/zostay/dev-tools/pkg/config"
)

var ErrUnsupportedTask = fmt.Errorf("this plugin does not support that task")

// TaskInterface is the base interface that all plugins implement.
type TaskInterface interface {
	// Implements will list the names of the tasks that this plugin TaskInterface
	// implements.
	Implements(ctx context.Context) (taskNames []string, err error)

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
	// or Close on the TaskInterface with the given task object or risk leaking
	// resources or having other unfinished working and unexpected results.
	Prepare(
		ctx context.Context,
		taskName string,
		globalCfg *config.Config,
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

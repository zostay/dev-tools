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
	Prepare(
		ctx context.Context,
		taskName string,
		globalCfg *config.Config,
	) (task Task, err error)
}

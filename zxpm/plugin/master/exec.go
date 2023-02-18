package master

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/zostay/dev-tools/pkg/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

type Error []error

func (e Error) Error() string {
	msgs := make([]string, len(e))
	for i, err := range e {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "; ")
}

func executeTaskOperation(
	ctx context.Context,
	ts plugin.Tasks,
	op func(context.Context, plugin.Task) error,
) error {
	opfs := make([]plugin.OperationFunc, len(ts))
	for i := range ts {
		t := ts[i]
		opfs = append(opfs, func(ctx context.Context) error {
			return op(ctx, t)
		})
	}

	return executeOperationFuncs(ctx, opfs)
}

func executeOperationFuncs(
	ctx context.Context,
	opfs []plugin.OperationFunc,
) error {
	wg := &sync.WaitGroup{}
	errs := make(chan error, len(opfs))

	for _, op := range opfs {
		wg.Add(1)
		go func() {
			defer wg.Done()

			errs <- op(ctx)
		}()
	}

	resultErr := make(Error, 0, len(opfs))
	for i := 0; i < len(opfs); i++ {
		err := <-errs
		if err != nil {
			resultErr = append(resultErr, err)
		}
	}

	if len(resultErr) > 0 {
		return resultErr
	}
	return nil
}

func evaluateOperations(
	ctx context.Context,
	ts plugin.Tasks,
	op func(plugin.Task, context.Context) (plugin.Operations, error),
) (plugin.Operations, error) {
	ops := make(plugin.Operations, 0, len(ts))
	for _, t := range ts {
		thisOps, err := op(t, ctx)
		if err != nil {
			return nil, err
		}
		ops = append(ops, thisOps...)
	}

	sort.Slice(ops, func(i, j int) bool {
		return ops[i].Order < ops[j].Order
	})

	return ops, nil
}

func gatherOperationFuncs(
	ops plugin.Operations,
) []plugin.Operations {
	opss := make([]plugin.Operations, 0, len(ops))
	var lastOrder plugin.Ordering = -1
	for _, op := range ops {
		order := op.Order
		if order < 0 {
			order = 0
		} else if order > 100 {
			order = 100
		}

		if order > lastOrder {
			opss = append(opss, make(plugin.Operations, 0, len(ops)))
		}

		opss[len(opss)-1] = append(opss[len(opss)-1], op)

		lastOrder = order
	}

	return opss
}

func executeOperationGroup(
	ctx context.Context,
	ops plugin.Operations,
) error {
	opfs := make([]plugin.OperationFunc, len(ops))
	for i, op := range ops {
		opfs[i] = op.Action.Call
	}
	return executeOperationFuncs(ctx, opfs)
}

func executeOperationGroups(
	ctx context.Context,
	opss []plugin.Operations,
) error {
	for _, ops := range opss {
		err := executeOperationGroup(ctx, ops)
		if err != nil {
			return fmt.Errorf("order %d: %w", ops[0].Order, err)
		}
	}
	return nil
}

func executePrioritizedOperations(
	ctx context.Context,
	stage string,
	ts plugin.Tasks,
	op func(plugin.Task, context.Context) (plugin.Operations, error),
) error {
	ops, err := evaluateOperations(ctx, ts, op)
	if err != nil {
		return err
	}
	opfss := gatherOperationFuncs(ops)
	err = executeOperationGroups(ctx, opfss)
	if err != nil {
		return fmt.Errorf("failed %s stage %w", stage, err)
	}
	return err
}

type taskOperationFunc func(plugin.Task, context.Context) error

func executeBasicStage(
	opFunc taskOperationFunc,
	stage string,
) func(context.Context, plugin.Task) error {
	return func(ctx context.Context, t plugin.Task) error {
		err := opFunc(t, ctx)
		if err != nil {
			return fmt.Errorf("failed %s stage: %w", stage, err)
		}
		return nil
	}
}

var (
	executeSetup     = executeBasicStage(plugin.Task.Setup, "setup")
	executeCheck     = executeBasicStage(plugin.Task.Check, "check")
	executeFinishing = executeBasicStage(plugin.Task.Finishing, "finishing")
	executeTeardown  = executeBasicStage(plugin.Task.Teardown, "teardown")
)

func Execute(
	ctx context.Context,
	cfg *config.Config,
	ts plugin.Tasks,
) error {
	pctx := plugin.NewPluginContext(cfg)
	plugin.InitializeContext(ctx, pctx)

	err := executeTaskOperation(ctx, ts, executeSetup)
	if err != nil {
		return err
	}

	err = executeTaskOperation(ctx, ts, executeCheck)
	if err != nil {
		return err
	}

	err = executePrioritizedOperations(ctx, "begin", ts, plugin.Task.Begin)
	if err != nil {
		return err
	}

	err = executePrioritizedOperations(ctx, "run", ts, plugin.Task.Run)
	if err != nil {
		return err
	}

	err = executePrioritizedOperations(ctx, "end", ts, plugin.Task.End)
	if err != nil {
		return err
	}

	err = executeTaskOperation(ctx, ts, executeFinishing)
	if err != nil {
		return err
	}

	// TODO figure out how to ensure that teardown is always run
	err = executeTaskOperation(ctx, ts, executeTeardown)
	if err != nil {
		return err
	}

	return nil
}

package plugin

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/zostay/dev-tools/pkg/config"
	"github.com/zostay/dev-tools/zxpm/plugin/tools"
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
	ts Tasks,
	op func(context.Context, Task) error,
) error {
	opfs := make([]OperationFunc, len(ts))
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
	opfs []OperationFunc,
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
	ts Tasks,
	op func(Task) Operations,
) Operations {
	ops := make(Operations, 0, len(ts))
	for _, t := range ts {
		ops = append(ops, op(t)...)
	}

	sort.Slice(ops, func(i, j int) bool {
		return ops[i].Order < ops[j].Order
	})

	return ops
}

func gatherOperationFuncs(
	ops Operations,
) []Operations {
	opss := make([]Operations, 0, len(ops))
	var lastOrder Ordering = -1
	for _, op := range ops {
		order := op.Order
		if order < 0 {
			order = 0
		} else if order > 100 {
			order = 100
		}

		if order > lastOrder {
			opss = append(opss, make(Operations, 0, len(ops)))
		}

		opss[len(opss)-1] = append(opss[len(opss)-1], op)

		lastOrder = order
	}

	return opss
}

func executeOperationGroup(
	ctx context.Context,
	ops Operations,
) error {
	opfs := make([]OperationFunc, len(ops))
	for i, op := range ops {
		opfs[i] = op.Action
	}
	return executeOperationFuncs(ctx, opfs)
}

func executeOperationGroups(
	ctx context.Context,
	opss []Operations,
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
	ts Tasks,
	op func(Task) Operations,
) error {
	ops := evaluateOperations(ts, op)
	opfss := gatherOperationFuncs(ops)
	err := executeOperationGroups(ctx, opfss)
	if err != nil {
		return fmt.Errorf("failed %s stage %w", stage, err)
	}
	return err
}

type taskOperationFunc func(Task, context.Context) error

func executeBasicStage(
	opFunc taskOperationFunc,
	stage string,
) func(context.Context, Task) error {
	return func(ctx context.Context, t Task) error {
		err := opFunc(t, ctx)
		if err != nil {
			return fmt.Errorf("failed %s stage: %w", stage, err)
		}
		return nil
	}
}

var (
	executeSetup     = executeBasicStage(Task.Setup, "setup")
	executeCheck     = executeBasicStage(Task.Check, "check")
	executeFinishing = executeBasicStage(Task.Finishing, "finishing")
	executeTeardown  = executeBasicStage(Task.Teardown, "teardown")
)

func Execute(
	ctx context.Context,
	cfg *config.Config,
	ts Tasks,
) error {
	tools.InitializeContext(ctx, cfg)

	err := executeTaskOperation(ctx, ts, executeSetup)
	if err != nil {
		return err
	}

	err = executeTaskOperation(ctx, ts, executeCheck)
	if err != nil {
		return err
	}

	err = executePrioritizedOperations(ctx, "begin", ts, Task.Begin)
	if err != nil {
		return err
	}

	err = executePrioritizedOperations(ctx, "run", ts, Task.Run)
	if err != nil {
		return err
	}

	err = executePrioritizedOperations(ctx, "end", ts, Task.End)
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

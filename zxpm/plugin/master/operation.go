package master

import (
	"context"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type OperationHandler struct {
	ops plugin.Operations
}

func (h *OperationHandler) Call(ctx context.Context) error {
	return RunTasksAndAccumulateErrors[int, plugin.Operation](
		ctx,
		NewSliceIterator[plugin.Operation](h.ops),
		func(ctx context.Context, _ int, op plugin.Operation) error {
			return op.Action.Call(ctx)
		})
}

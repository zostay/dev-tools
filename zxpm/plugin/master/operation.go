package master

import (
	"context"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type OperationHandler struct {
	ops plugin.Operations
}

func (h *OperationHandler) Call(ctx context.Context) error {
	return RunTasksAndAccumulateErrors[plugin.Operation](ctx, h.ops,
		func(ctx context.Context, op plugin.Operation) error {
			return op.Action.Call(ctx)
		})
}

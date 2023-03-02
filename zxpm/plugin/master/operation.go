package master

import (
	"context"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type operationInfo struct {
	pluginName string
	op         plugin.Operation
}

func newOperationInfo(pluginName string, op plugin.Operation) *operationInfo {
	return &operationInfo{pluginName, op}
}

func operationInfoLess(opInfo []*operationInfo) func(i, j int) bool {
	return func(i, j int) bool {
		return opInfo[i].op.Order < opInfo[j].op.Order
	}
}

type OperationHandler struct {
	taskName string
	ti       *Interface
	opInfo   []*operationInfo
}

func newOperationHandler(
	taskName string,
	ti *Interface,
	opInfo []*operationInfo,
) *OperationHandler {
	return &OperationHandler{
		taskName: taskName,
		ti:       ti,
		opInfo:   opInfo,
	}
}

func (h *OperationHandler) Call(ctx context.Context) error {
	return RunTasksAndAccumulateErrors[int, *operationInfo](
		ctx,
		NewSliceIterator[*operationInfo](h.opInfo),
		func(ctx context.Context, _ int, info *operationInfo) error {
			ctx, pctx := h.ti.ctxFor(ctx, h.taskName, info.pluginName)
			err := info.op.Action.Call(ctx)
			if err != nil {
				return err
			}

			pctx.UpdateStorage(pctx.StorageChanges())

			return nil
		})
}

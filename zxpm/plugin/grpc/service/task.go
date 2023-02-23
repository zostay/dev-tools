package service

import (
	"context"

	"github.com/spf13/cast"

	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin/api"
)

func (s *TaskExecution) executeStage(
	ctx context.Context,
	request *api.Task_Operation_Request,
	op func(plugin.Task, context.Context) error,
) (*api.Task_Operation_Response, error) {
	state, err := s.deref(request.GetTask())
	if err != nil {
		return nil, err
	}

	state.Context.UpdateStorage(cast.ToStringMap(request.GetStorage()))
	ctx = plugin.InitializeContext(ctx, state.Context)

	err = op(state.Task, ctx)
	if err != nil {
		return nil, err
	}

	return &api.Task_Operation_Response{
		StorageUpdate: state.Context.StorageChanges(),
	}, nil
}

func (s *TaskExecution) ExecuteCheck(
	ctx context.Context,
	request *api.Task_Operation_Request,
) (*api.Task_Operation_Response, error) {
	return s.executeStage(ctx, request, plugin.Task.Check)
}

func (s *TaskExecution) prepareStage(
	ctx context.Context,
	ref *api.Task_Ref,
	prepare func(plugin.Task, context.Context) (plugin.Operations, error),
) (*api.Task_SubStage_Response, error) {
	state, err := s.deref(ref)
	if err != nil {
		return nil, err
	}

	ops, err := prepare(state.Task, ctx)
	if err != nil {
		return nil, err
	}

	orders := make([]int32, len(ops))
	for i, op := range ops {
		orders[i] = int32(op.Order)
	}

	return &api.Task_SubStage_Response{
		ProvidedOrders: orders,
	}, nil

}

func (s *TaskExecution) PrepareBegin(
	ctx context.Context,
	ref *api.Task_Ref,
) (*api.Task_SubStage_Response, error) {
	return s.prepareStage(ctx, ref, plugin.Task.Begin)
}

func (s *TaskExecution) PrepareRun(
	ctx context.Context,
	ref *api.Task_Ref,
) (*api.Task_SubStage_Response, error) {
	return s.prepareStage(ctx, ref, plugin.Task.Run)
}

func (s *TaskExecution) PrepareEnd(
	ctx context.Context,
	ref *api.Task_Ref,
) (*api.Task_SubStage_Response, error) {
	return s.prepareStage(ctx, ref, plugin.Task.End)
}

func (s *TaskExecution) ExecuteFinish(
	ctx context.Context,
	request *api.Task_Operation_Request,
) (*api.Task_Operation_Response, error) {
	return s.executeStage(ctx, request, plugin.Task.Finish)
}

package service

import (
	"context"

	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin/api"
	"github.com/zostay/dev-tools/zxpm/storage"
)

func (s *TaskExecution) executeSubStage(
	ctx context.Context,
	request *api.Task_Operation_Request,
	op plugin.OperationFunc,
) (*api.Task_Operation_Response, error) {
	return s.executeStage(ctx, request,
		func(_ plugin.Task, ctx context.Context) error {
			return op(ctx)
		},
	)
}

func (s *TaskExecution) executePrioritizedStage(
	ctx context.Context,
	request *api.Task_SubStage_Request,
	opList func(plugin.Task, context.Context) (plugin.Operations, error),
) (*api.Task_Operation_Response, error) {
	opRequest := request.GetRequest()
	state, err := s.deref(opRequest.GetTask())
	if err != nil {
		return nil, err
	}

	ops, err := opList(state.Task, ctx)
	if err != nil {
		return nil, err
	}

	// This is used to accumulate changes because we want to ensure that
	// operations that are supposed to be able run concurrently don't interfere
	// with each other. If the plugin author really wants to depend on the
	// values saved, they need to prioritize the operations as appropriate to
	// make that work.
	accChanges := storage.New()

	// TODO Should we run plugin-side operations concurrently in goroutines?
	var res *api.Task_Operation_Response
	for _, op := range ops {
		if op.Order == plugin.Ordering(request.SubStage) {
			res, err = s.executeStage(ctx,
				&api.Task_Operation_Request{
					Task:    opRequest.GetTask(),
					Storage: opRequest.GetStorage(),
				},
				func(_ plugin.Task, ctx context.Context) error {
					return op.Action.Call(ctx)
				},
			)

			theseChanges := res.GetStorageUpdate()
			accChanges.UpdateStrings(theseChanges)
			if err != nil {
				return nil, err
			}
		}
	}

	return &api.Task_Operation_Response{
		StorageUpdate: accChanges.AllSettingsStrings(),
	}, nil
}

func (s *TaskExecution) ExecuteBegin(
	ctx context.Context,
	request *api.Task_SubStage_Request,
) (*api.Task_Operation_Response, error) {
	return s.executePrioritizedStage(ctx, request, plugin.Task.Begin)
}

func (s *TaskExecution) ExecuteRun(
	ctx context.Context,
	request *api.Task_SubStage_Request,
) (*api.Task_Operation_Response, error) {
	return s.executePrioritizedStage(ctx, request, plugin.Task.Run)
}

func (s *TaskExecution) ExecuteEnd(
	ctx context.Context,
	request *api.Task_SubStage_Request,
) (*api.Task_Operation_Response, error) {
	return s.executePrioritizedStage(ctx, request, plugin.Task.End)
}

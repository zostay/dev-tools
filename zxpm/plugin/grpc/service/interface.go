package service

import (
	"context"

	"github.com/oklog/ulid/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin/api"
	"github.com/zostay/dev-tools/zxpm/plugin/translate"
)

type TaskState struct {
	Task    plugin.Task
	Context *plugin.Context
}

type TaskExecution struct {
	Impl  plugin.Interface
	state map[string]map[string]*TaskState
}

func NewGRPCTaskInterfaceClient(impl plugin.Interface) *TaskExecution {
	names, err := impl.Implements(context.Background())
	if err != nil {
		return nil
	}

	state := make(map[string]map[string]*TaskState, len(names))
	for _, name := range names {
		state[name] = make(map[string]*TaskState, 1)
	}

	return &TaskExecution{
		Impl:  impl,
		state: state,
	}
}

func generateStateId() string {
	return ulid.Make().String()
}

func (s *TaskExecution) Implements(
	ctx context.Context,
	_ *api.Task_Implements_Request,
) (*api.Task_Implements_Response, error) {
	names, err := s.Impl.Implements(ctx)
	if err != nil {
		return nil, err
	}
	return &api.Task_Implements_Response{
		Names: names,
	}, nil
}

func (s *TaskExecution) Prepare(
	ctx context.Context,
	request *api.Task_Prepare_Request,
) (*api.Task_Prepare_Response, error) {
	globalConfig := translate.APIConfigToConfig(request.GetGlobalConfig())

	pctx := plugin.NewContext(globalConfig)
	ctx = plugin.InitializeContext(ctx, pctx)

	task, err := s.Impl.Prepare(ctx, request.GetName(), globalConfig)
	if err != nil {
		return nil, err
	}

	state := &TaskState{
		Task:    task,
		Context: pctx,
	}

	name := request.GetName()
	id := generateStateId()
	s.state[name][id] = state

	res, err := s.executeStage(ctx, &api.Task_Operation_Request{
		Task: &api.Task_Ref{
			Name:    name,
			StateId: id,
		},
		Storage: map[string]string{},
	}, plugin.Task.Setup)
	if err != nil {
		return nil, err
	}

	return &api.Task_Prepare_Response{
		Task: &api.Task_Ref{
			Name:    request.GetName(),
			StateId: id,
		},
		Storage: res.GetStorageUpdate(),
	}, nil
}

func (s *TaskExecution) deref(ref *api.Task_Ref) (*TaskState, error) {
	name := ref.GetName()
	id := ref.GetStateId()
	task := s.state[name][id]
	if task == nil {
		return nil, status.Errorf(codes.NotFound, "the task named %q with state ID %q not found", name, id)
	}

	return task, nil
}

func (s *TaskExecution) closeTask(
	ctx context.Context,
	taskRef *api.Task_Ref,
) error {
	_, err := s.deref(taskRef)
	if err != nil {
		return err
	}

	_, err = s.executeStage(ctx, &api.Task_Operation_Request{
		Task:    taskRef,
		Storage: map[string]string{},
	}, plugin.Task.Teardown)

	delete(s.state[taskRef.GetName()], taskRef.GetStateId())

	return err
}

func (s *TaskExecution) Cancel(
	ctx context.Context,
	request *api.Task_Cancel_Request,
) (*api.Task_Cancel_Response, error) {
	err := s.closeTask(ctx, request.GetTask())
	if err != nil {
		return nil, err
	}
	return &api.Task_Cancel_Response{}, nil
}

func (s *TaskExecution) Complete(
	ctx context.Context,
	request *api.Task_Complete_Request,
) (*api.Task_Complete_Response, error) {
	err := s.closeTask(ctx, request.GetTask())
	if err != nil {
		return nil, err
	}
	return &api.Task_Complete_Response{}, nil
}

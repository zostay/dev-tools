package plugin

import (
	"context"

	"github.com/oklog/ulid/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zostay/dev-tools/zxpm/plugin/api"
	"github.com/zostay/dev-tools/zxpm/plugin/translate"
)

type GRPCTaskClientState struct {
	Task    Task
	Context *PluginContext
}

type GRPCTaskInterfaceServer struct {
	Impl  TaskInterface
	state map[string]map[string]*GRPCTaskClientState
}

func NewGRPCTaskInterfaceClient(impl TaskInterface) *GRPCTaskInterfaceServer {
	names, err := impl.Implements()
	if err != nil {
		return nil
	}

	state := make(map[string]map[string]*GRPCTaskClientState, len(names))
	for _, name := range names {
		state[name] = make(map[string]*GRPCTaskClientState, 1)
	}

	return &GRPCTaskInterfaceServer{
		Impl:  impl,
		state: state,
	}
}

func generateStateId() string {
	return ulid.Make().String()
}

func (s *GRPCTaskInterfaceServer) Implements(
	_ context.Context,
	_ *api.Task_Implements_Request,
) (*api.Task_Implements_Response, error) {
	names, err := s.Impl.Implements()
	if err != nil {
		return nil, err
	}
	return &api.Task_Implements_Response{
		Names: names,
	}, nil
}

func (s *GRPCTaskInterfaceServer) Prepare(
	ctx context.Context,
	request *api.Task_Prepare_Request,
) (*api.Task_Prepare_Response, error) {
	globalConfig := translate.APIConfigToPluginConfig(request.GetGlobalConfig())

	task, err := s.Impl.Prepare(request.GetName(), globalConfig)
	if err != nil {
		return nil, err
	}

	state := &GRPCTaskClientState{
		Task:    task,
		Context: NewPluginContext(globalConfig),
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
	}, Task.Setup)
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

func (s *GRPCTaskInterfaceServer) deref(ref *api.Task_Ref) (*GRPCTaskClientState, error) {
	name := ref.GetName()
	id := ref.GetStateId()
	task := s.state[name][id]
	if task == nil {
		return nil, status.Errorf(codes.NotFound, "the task named %q with state ID %q not found", name, id)
	}

	return task, nil
}

func (s *GRPCTaskInterfaceServer) closeTask(
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
	}, Task.Teardown)

	delete(s.state[taskRef.GetName()], taskRef.GetStateId())

	return err
}

func (s *GRPCTaskInterfaceServer) Cancel(
	ctx context.Context,
	request *api.Task_Cancel_Request,
) (*api.Task_Cancel_Response, error) {
	err := s.closeTask(ctx, request.GetTask())
	if err != nil {
		return nil, err
	}
	return &api.Task_Cancel_Response{}, nil
}

func (s *GRPCTaskInterfaceServer) Complete(
	ctx context.Context,
	request *api.Task_Complete_Request,
) (*api.Task_Complete_Response, error) {
	err := s.closeTask(ctx, request.GetTask())
	if err != nil {
		return nil, err
	}
	return &api.Task_Complete_Response{}, nil
}

func (s *GRPCTaskInterfaceServer) executeStage(
	ctx context.Context,
	request *api.Task_Operation_Request,
	op func(Task, context.Context) error,
) (*api.Task_Operation_Response, error) {
	state, err := s.deref(request.GetTask())
	if err != nil {
		return nil, err
	}

	state.Context.UpdateStorage(request.GetStorage())
	ctx = InitializeContext(ctx, state.Context)

	err = op(state.Task, ctx)
	if err != nil {
		return nil, err
	}

	return &api.Task_Operation_Response{
		StorageUpdate: state.Context.StorageChanges(),
	}, nil
}

func (s *GRPCTaskInterfaceServer) ExecuteCheck(
	ctx context.Context,
	request *api.Task_Operation_Request,
) (*api.Task_Operation_Response, error) {
	return s.executeStage(ctx, request, Task.Check)
}

func (s *GRPCTaskInterfaceServer) prepareStage(
	ctx context.Context,
	ref *api.Task_Ref,
	prepare func(Task, context.Context) (Operations, error),
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

func (s *GRPCTaskInterfaceServer) PrepareBegin(
	ctx context.Context,
	ref *api.Task_Ref,
) (*api.Task_SubStage_Response, error) {
	return s.prepareStage(ctx, ref, Task.Begin)
}

func (s *GRPCTaskInterfaceServer) executeSubStage(
	ctx context.Context,
	request *api.Task_Operation_Request,
	op OperationFunc,
) (*api.Task_Operation_Response, error) {
	return s.executeStage(ctx, request,
		func(_ Task, ctx context.Context) error {
			return op(ctx)
		},
	)
}

func (s *GRPCTaskInterfaceServer) executePrioritizedStage(
	ctx context.Context,
	request *api.Task_SubStage_Request,
	opList func(Task, context.Context) (Operations, error),
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

	var res *api.Task_Operation_Response
	accChanges := make(map[string]string, 10)
	for _, op := range ops {
		if op.Order == Ordering(request.SubStage) {
			res, err = s.executeStage(ctx,
				&api.Task_Operation_Request{
					Task:    opRequest.GetTask(),
					Storage: opRequest.GetStorage(),
				},
				func(_ Task, ctx context.Context) error {
					return op.Action(ctx)
				},
			)

			theseChanges := res.GetStorageUpdate()
			for k, v := range theseChanges {
				accChanges[k] = v
			}

			if err != nil {
				return nil, err
			}
		}
	}

	return &api.Task_Operation_Response{
		StorageUpdate: accChanges,
	}, nil
}

func (s *GRPCTaskInterfaceServer) ExecuteBegin(
	ctx context.Context,
	request *api.Task_SubStage_Request,
) (*api.Task_Operation_Response, error) {
	return s.executePrioritizedStage(ctx, request, Task.Begin)
}

func (s *GRPCTaskInterfaceServer) PrepareRun(
	ctx context.Context,
	ref *api.Task_Ref,
) (*api.Task_SubStage_Response, error) {
	return s.prepareStage(ctx, ref, Task.Run)
}

func (s *GRPCTaskInterfaceServer) ExecuteRun(
	ctx context.Context,
	request *api.Task_SubStage_Request,
) (*api.Task_Operation_Response, error) {
	return s.executePrioritizedStage(ctx, request, Task.Run)
}

func (s *GRPCTaskInterfaceServer) PrepareEnd(
	ctx context.Context,
	ref *api.Task_Ref,
) (*api.Task_SubStage_Response, error) {
	return s.prepareStage(ctx, ref, Task.End)
}

func (s *GRPCTaskInterfaceServer) ExecuteEnd(ctx context.Context, request *api.Task_SubStage_Request) (*api.Task_Operation_Response, error) {
	return s.executePrioritizedStage(ctx, request, Task.End)
}

func (s *GRPCTaskInterfaceServer) ExecuteFinishing(
	ctx context.Context,
	request *api.Task_Operation_Request,
) (*api.Task_Operation_Response, error) {
	return s.executeStage(ctx, request, Task.Finishing)
}

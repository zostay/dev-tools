package plugin

import (
	"context"

	"github.com/hashicorp/go-plugin"
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

type GRPCTaskInterfaceClient struct {
	plugin.Plugin

	Impl  TaskInterface
	state map[string]map[string]*GRPCTaskClientState
}

func NewGRPCTaskInterfaceClient(impl TaskInterface) *GRPCTaskInterfaceClient {
	names := impl.Implements()
	state := make(map[string]map[string]*GRPCTaskClientState, len(names))
	for _, name := range names {
		state[name] = make(map[string]*GRPCTaskClientState, 1)
	}

	return &GRPCTaskInterfaceClient{
		Impl:  impl,
		state: state,
	}
}

func generateStateId() string {
	return ulid.Make().String()
}

func (c *GRPCTaskInterfaceClient) Implements(
	_ context.Context,
	_ *api.Task_Implements_Request,
) (*api.Task_Implements_Response, error) {
	names := c.Impl.Implements()
	return &api.Task_Implements_Response{
		Names: names,
	}, nil
}

func (c *GRPCTaskInterfaceClient) Prepare(
	ctx context.Context,
	request *api.Task_Prepare_Request,
) (*api.Task_Prepare_Response, error) {
	globalConfig := translate.APIConfigToPluginConfig(request.GetGlobalConfig())

	task := c.Impl.Prepare(request.GetName(), globalConfig)

	state := &GRPCTaskClientState{
		Task:    task,
		Context: NewPluginContext(globalConfig),
	}

	name := request.GetName()
	id := generateStateId()
	c.state[name][id] = state

	res, err := c.executeStage(ctx, &api.Task_Operation_Request{
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

func (c *GRPCTaskInterfaceClient) deref(ref *api.Task_Ref) (*GRPCTaskClientState, error) {
	name := ref.GetName()
	id := ref.GetStateId()
	task := c.state[name][id]
	if task == nil {
		return nil, status.Errorf(codes.NotFound, "the task named %q with state ID %q not found", name, id)
	}

	return task, nil
}

func (c *GRPCTaskInterfaceClient) closeTask(
	ctx context.Context,
	taskRef *api.Task_Ref,
) error {
	_, err := c.deref(taskRef)
	if err != nil {
		return err
	}

	_, err = c.executeStage(ctx, &api.Task_Operation_Request{
		Task:    taskRef,
		Storage: map[string]string{},
	}, Task.Teardown)

	delete(c.state[taskRef.GetName()], taskRef.GetStateId())

	return err
}

func (c *GRPCTaskInterfaceClient) Cancel(
	ctx context.Context,
	request *api.Task_Cancel_Request,
) (*api.Task_Cancel_Response, error) {
	err := c.closeTask(ctx, request.GetTask())
	if err != nil {
		return nil, err
	}
	return &api.Task_Cancel_Response{}, nil
}

func (c *GRPCTaskInterfaceClient) Complete(
	ctx context.Context,
	request *api.Task_Complete_Request,
) (*api.Task_Complete_Response, error) {
	err := c.closeTask(ctx, request.GetTask())
	if err != nil {
		return nil, err
	}
	return &api.Task_Complete_Response{}, nil
}

func (c *GRPCTaskInterfaceClient) executeStage(
	ctx context.Context,
	request *api.Task_Operation_Request,
	op func(Task, context.Context) error,
) (*api.Task_Operation_Response, error) {
	state, err := c.deref(request.GetTask())
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

func (c *GRPCTaskInterfaceClient) ExecuteCheck(
	ctx context.Context,
	request *api.Task_Operation_Request,
) (*api.Task_Operation_Response, error) {
	return c.executeStage(ctx, request, Task.Check)
}

func (c *GRPCTaskInterfaceClient) prepareStage(
	ref *api.Task_Ref,
	prepare func(Task) Operations,
) (*api.Task_SubStage_Response, error) {
	state, err := c.deref(ref)
	if err != nil {
		return nil, err
	}

	ops := prepare(state.Task)
	orders := make([]int32, len(ops))
	for i, op := range ops {
		orders[i] = int32(op.Order)
	}

	return &api.Task_SubStage_Response{
		ProvidedOrders: orders,
	}, nil

}

func (c *GRPCTaskInterfaceClient) PrepareBegin(
	_ context.Context,
	ref *api.Task_Ref,
) (*api.Task_SubStage_Response, error) {
	return c.prepareStage(ref, Task.Begin)
}

func (c *GRPCTaskInterfaceClient) executeSubStage(
	ctx context.Context,
	request *api.Task_Operation_Request,
	op OperationFunc,
) (*api.Task_Operation_Response, error) {
	return c.executeStage(ctx, request,
		func(_ Task, ctx context.Context) error {
			return op(ctx)
		},
	)
}

func (c *GRPCTaskInterfaceClient) executePrioritizedStage(
	ctx context.Context,
	request *api.Task_SubStage_Request,
	opList func(Task) Operations,
) (*api.Task_Operation_Response, error) {
	opRequest := request.GetRequest()
	state, err := c.deref(opRequest.GetTask())
	if err != nil {
		return nil, err
	}

	ops := opList(state.Task)
	var res *api.Task_Operation_Response
	accChanges := make(map[string]string, 10)
	for _, op := range ops {
		if op.Order == Ordering(request.SubStage) {
			res, err = c.executeStage(ctx,
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

func (c *GRPCTaskInterfaceClient) ExecuteBegin(
	ctx context.Context,
	request *api.Task_SubStage_Request,
) (*api.Task_Operation_Response, error) {
	return c.executePrioritizedStage(ctx, request, Task.Begin)
}

func (c *GRPCTaskInterfaceClient) PrepareRun(
	_ context.Context,
	ref *api.Task_Ref,
) (*api.Task_SubStage_Response, error) {
	return c.prepareStage(ref, Task.Run)
}

func (c *GRPCTaskInterfaceClient) ExecuteRun(
	ctx context.Context,
	request *api.Task_SubStage_Request,
) (*api.Task_Operation_Response, error) {
	return c.executePrioritizedStage(ctx, request, Task.Run)
}

func (c *GRPCTaskInterfaceClient) PrepareEnd(
	_ context.Context,
	ref *api.Task_Ref,
) (*api.Task_SubStage_Response, error) {
	return c.prepareStage(ref, Task.End)
}

func (c *GRPCTaskInterfaceClient) ExecuteEnd(ctx context.Context, request *api.Task_SubStage_Request) (*api.Task_Operation_Response, error) {
	return c.executePrioritizedStage(ctx, request, Task.End)
}

func (c *GRPCTaskInterfaceClient) ExecuteFinishing(
	ctx context.Context,
	request *api.Task_Operation_Request,
) (*api.Task_Operation_Response, error) {
	return c.executeStage(ctx, request, Task.Finishing)
}

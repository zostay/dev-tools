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
	state := make(map[string]map[string]Task, len(names))
	for _, name := range names {
		state[name] = make(map[string]Task, 1)
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
	_ *api.Implements_Request,
) (*api.Implements_Response, error) {
	names := c.Impl.Implements()
	return &api.Implements_Response{
		Names: names,
	}, nil
}

func (c *GRPCTaskInterfaceClient) Prepare(
	_ context.Context,
	request *api.Prepare_Request,
) (*api.Prepare_Response, error) {
	globalConfig := translate.APIConfigToPluginConfig(request.GetGlobalConfig())

	task := c.Impl.Prepare(request.GetName(), globalConfig)

	state := &GRPCTaskClientState{
		Task:    task,
		Context: NewPluginContext(globalConfig),
	}

	name := request.GetName()
	id := generateStateId()
	c.state[name][id] = state

	return &api.Prepare_Response{
		Task: &api.Task_Ref{
			Name:    request.GetName(),
			StateId: id,
		},
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

func (c *GRPCTaskInterfaceClient) Cancel(
	_ context.Context,
	request *api.Cancel_Request,
) (*api.Cancel_Response, error) {
	_, err := c.deref(request.GetTask())
	if err != nil {
		return nil, err
	}

	delete(c.state[request.GetTask().GetName()], request.GetTask().GetStateId())

	return &api.Cancel_Response{}, nil
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

func (c *GRPCTaskInterfaceClient) PrepareBegin(
	ctx context.Context,
	ref *api.Task_Ref,
) (*api.Task_SubStage_Response, error) {

	// TODO implement me
	panic("implement me")
}

func (c *GRPCTaskInterfaceClient) ExecuteBegin(ctx context.Context, request *api.Task_SubStage_Request) (*api.Task_Operation_Response, error) {
	// TODO implement me
	panic("implement me")
}

func (c *GRPCTaskInterfaceClient) PrepareRun(ctx context.Context, ref *api.Task_Ref) (*api.Task_SubStage_Response, error) {
	// TODO implement me
	panic("implement me")
}

func (c *GRPCTaskInterfaceClient) ExecuteRun(ctx context.Context, response *api.Task_SubStage_Response) (*api.Task_Operation_Response, error) {
	// TODO implement me
	panic("implement me")
}

func (c *GRPCTaskInterfaceClient) PrepareEnd(ctx context.Context, ref *api.Task_Ref) (*api.Task_SubStage_Response, error) {
	// TODO implement me
	panic("implement me")
}

func (c *GRPCTaskInterfaceClient) ExecuteEnd(ctx context.Context, request *api.Task_SubStage_Request) (*api.Task_Operation_Response, error) {
	// TODO implement me
	panic("implement me")
}

func (c *GRPCTaskInterfaceClient) ExecuteFinishing(ctx context.Context, request *api.Task_Operation_Request) (*api.Task_Operation_Response, error) {
	// TODO implement me
	panic("implement me")
}

func (c *GRPCTaskInterfaceClient) mustEmbedUnimplementedTaskExecutionServer() {
	// TODO implement me
	panic("implement me")
}

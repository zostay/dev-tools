package plugin

import (
	"context"
	"sort"

	"google.golang.org/grpc"

	"github.com/zostay/dev-tools/zxpm/plugin/api"
	"github.com/zostay/dev-tools/zxpm/plugin/translate"
)

type GRPCTaskInterfaceClient struct {
	client api.TaskExecutionClient
}

type GRPCTaskClient struct {
	client  api.TaskExecutionClient
	ref     *api.Task_Ref
	storage map[string]string
}

type GRPCTaskOperationClient struct {
	parent *GRPCTaskClient
	call   func(context.Context, *api.Task_SubStage_Request, ...grpc.CallOption) (*api.Task_Operation_Response, error)
	order  int32
}

func (c *GRPCTaskInterfaceClient) Implements() ([]string, error) {
	res, err := c.client.Implements(context.Background(), &api.Task_Implements_Request{})
	if err != nil {
		return nil, err
	}
	return res.GetNames(), nil
}

func (c *GRPCTaskInterfaceClient) Prepare(
	taskName string,
	globalCfg *Config,
) (Task, error) {
	res, err := c.client.Prepare(context.Background(),
		&api.Task_Prepare_Request{
			Name:         taskName,
			GlobalConfig: translate.PluginConfigToAPIConfig(globalCfg),
		},
	)
	if err != nil {
		return nil, err
	}

	return &GRPCTaskClient{
		client:  c.client,
		ref:     res.GetTask(),
		storage: res.GetStorage(),
	}, nil
}

func (t *GRPCTaskClient) Setup(
	_ context.Context,
) error {
	return nil
}

func (t *GRPCTaskClient) applyStorageUpdate(changes map[string]string) {
	for k, v := range changes {
		t.storage[k] = v
	}
}

func (t *GRPCTaskClient) operation(
	ctx context.Context,
	op func(context.Context, *api.Task_Operation_Request, ...grpc.CallOption) (*api.Task_Operation_Response, error),
) error {
	res, err := op(ctx, &api.Task_Operation_Request{
		Task:    t.ref,
		Storage: t.storage,
	})

	if err != nil {
		return err
	}

	t.applyStorageUpdate(res.GetStorageUpdate())

	return nil
}

func (t *GRPCTaskClient) Check(
	ctx context.Context,
) error {
	return t.operation(ctx, t.client.ExecuteCheck)
}

func (t *GRPCTaskClient) operations(
	ctx context.Context,
	prepare func(context.Context, *api.Task_Ref, ...grpc.CallOption) (*api.Task_SubStage_Response, error),
) (Operations, error) {
	res, err := prepare(ctx, t.ref)
	if err != nil {
		return nil, err
	}

	orders := res.ProvidedOrders
	sort.Slice(orders, func(i, j int) bool { return orders[i] < orders[j] })

	ops := make(Operations, len(orders))
	for i, order := range orders {
		ops[i] = Operation{
			Order: Ordering(order),
			Action: &GRPCTaskOperationClient{
				parent: t,
				call:   t.client.ExecuteBegin,
				order:  order,
			},
		}
	}

	return ops, nil
}

func (t *GRPCTaskClient) Begin(ctx context.Context) (Operations, error) {
	return t.operations(ctx, t.client.PrepareBegin)
}

func (t *GRPCTaskClient) Run(ctx context.Context) (Operations, error) {
	return t.operations(ctx, t.client.PrepareRun)
}

func (t *GRPCTaskClient) End(ctx context.Context) (Operations, error) {
	return t.operations(ctx, t.client.PrepareEnd)
}

func (t *GRPCTaskClient) Finishing(
	ctx context.Context,
) error {
	return t.operation(ctx, t.client.ExecuteCheck)
}

func (t *GRPCTaskClient) Teardown(
	_ context.Context,
) error {
	return nil
}

func (o *GRPCTaskOperationClient) Call(ctx context.Context) error {
	res, err := o.call(ctx, &api.Task_SubStage_Request{
		Request: &api.Task_Operation_Request{
			Task:    o.parent.ref,
			Storage: o.parent.storage,
		},
		SubStage: o.order,
	})

	if err != nil {
		return err
	}

	o.parent.applyStorageUpdate(res.GetStorageUpdate())

	return nil
}

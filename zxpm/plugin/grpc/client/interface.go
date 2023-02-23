package client

import (
	"context"

	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin/api"
	"github.com/zostay/dev-tools/zxpm/plugin/translate"
)

var _ plugin.Interface = &TaskInterface{}

type TaskInterface struct {
	client api.TaskExecutionClient
}

func NewGRPCTaskInterface(client api.TaskExecutionClient) *TaskInterface {
	return &TaskInterface{client}
}

func (c *TaskInterface) Implements(
	ctx context.Context,
) ([]plugin.TaskDescription, error) {
	res, err := c.client.Implements(ctx, &api.Task_Implements_Request{})
	if err != nil {
		return nil, err
	}
	return translate.APITaskDescriptorsToPluginTaskDescriptions(res.GetTasks()), nil
}

func (c *TaskInterface) Goal(
	ctx context.Context,
	goalName string,
) (plugin.GoalDescription, error) {
	res, err := c.client.Goal(ctx, &api.Task_Goal_Request{
		Name: goalName,
	})
	if err != nil {
		return nil, err
	}
	return translate.APIGoalDescriptorToPluginGoalDescription(res.GetDefinition()), nil
}

func (c *TaskInterface) Prepare(
	ctx context.Context,
	taskName string,
) (plugin.Task, error) {
	res, err := c.client.Prepare(ctx,
		&api.Task_Prepare_Request{
			Name:         taskName,
			GlobalConfig: translate.KVToAPIConfig(plugin.KV(ctx)),
		},
	)
	if err != nil {
		return nil, err
	}

	return &Task{
		client:  c.client,
		ref:     res.GetTask(),
		storage: res.GetStorage(),
	}, nil
}

func (c *TaskInterface) Cancel(
	ctx context.Context,
	task plugin.Task,
) error {
	ref := task.(*Task).ref
	_, err := c.client.Cancel(ctx, &api.Task_Cancel_Request{
		Task: ref,
	})
	return err
}

func (c *TaskInterface) Complete(
	ctx context.Context,
	task plugin.Task,
) error {
	ref := task.(*Task).ref
	_, err := c.client.Complete(ctx, &api.Task_Complete_Request{
		Task: ref,
	})
	return err
}

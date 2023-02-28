package client

import (
	"context"
	"strings"

	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin/api"
	"github.com/zostay/dev-tools/zxpm/plugin/translate"
)

var _ plugin.Interface = &Interface{}

type Interface struct {
	client api.TaskExecutionClient
}

func NewGRPCTaskInterface(client api.TaskExecutionClient) *Interface {
	return &Interface{client}
}

func (c *Interface) Implements(
	ctx context.Context,
) ([]plugin.TaskDescription, error) {
	res, err := c.client.Implements(ctx, &api.Task_Implements_Request{})
	if err != nil {
		return nil, err
	}
	return translate.APITaskDescriptorsToPluginTaskDescriptions(res.GetTasks()), nil
}

func (c *Interface) Goal(
	ctx context.Context,
	goalName string,
) (plugin.GoalDescription, error) {
	res, err := c.client.Goal(ctx, &api.Task_Goal_Request{
		Name: goalName,
	})
	if err != nil {
		if strings.Contains(err.Error(), plugin.ErrUnsupportedGoal.Error()) {
			return nil, plugin.ErrUnsupportedGoal
		}
		return nil, err
	}
	return translate.APIGoalDescriptorToPluginGoalDescription(res.GetDefinition()), nil
}

func (c *Interface) Prepare(
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
		if strings.Contains(err.Error(), plugin.ErrUnsupportedTask.Error()) {
			return nil, plugin.ErrUnsupportedTask
		}
		return nil, err
	}

	chgs := res.GetStorage()
	plugin.UpdateKV(ctx, chgs)

	return &Task{
		client: c.client,
		ref:    res.GetTask(),
	}, nil
}

func (c *Interface) Cancel(
	ctx context.Context,
	task plugin.Task,
) error {
	ref := task.(*Task).ref
	_, err := c.client.Cancel(ctx, &api.Task_Cancel_Request{
		Task: ref,
	})
	return err
}

func (c *Interface) Complete(
	ctx context.Context,
	task plugin.Task,
) error {
	ref := task.(*Task).ref
	_, err := c.client.Complete(ctx, &api.Task_Complete_Request{
		Task: ref,
	})
	return err
}

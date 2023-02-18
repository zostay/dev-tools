package client

import (
	"context"

	"github.com/zostay/dev-tools/pkg/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin/api"
	"github.com/zostay/dev-tools/zxpm/plugin/translate"
)

type TaskInterface struct {
	client api.TaskExecutionClient
}

func (c *TaskInterface) Implements() ([]string, error) {
	res, err := c.client.Implements(context.Background(), &api.Task_Implements_Request{})
	if err != nil {
		return nil, err
	}
	return res.GetNames(), nil
}

func (c *TaskInterface) Prepare(
	taskName string,
	globalCfg *config.Config,
) (plugin.Task, error) {
	res, err := c.client.Prepare(context.Background(),
		&api.Task_Prepare_Request{
			Name:         taskName,
			GlobalConfig: translate.ConfigToAPIConfig(globalCfg),
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

package client

import (
	"context"

	"google.golang.org/grpc"

	"github.com/zostay/dev-tools/zxpm/plugin/api"
)

type Operation struct {
	parent *Task
	call   func(context.Context, *api.Task_SubStage_Request, ...grpc.CallOption) (*api.Task_Operation_Response, error)
	order  int32
}

func (o *Operation) Call(ctx context.Context) error {
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

package plugin

import "context"

type Ordering int

type OperationFunc func(ctx context.Context) error

func (op OperationFunc) Call(ctx context.Context) error {
	return op(ctx)
}

type OperationHandler interface {
	Call(ctx context.Context) error
}

type Operations []Operation

type Operation struct {
	Order  Ordering
	Action OperationHandler
}

func OperationLess(ops Operations) func(int, int) bool {
	return func(i, j int) bool {
		return ops[i].Order < ops[j].Order
	}
}

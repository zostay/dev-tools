package master

import (
	"context"
	"sync"
)

func RunTasksAndAccumulate[In, Out any](
	ctx context.Context,
	inputs []In,
	task func(context.Context, In) (Out, error),
) ([]Out, error) {
	results := make([]Out, 0, len(inputs))
	accErr := make(Error, 0, len(inputs))

	resChan := make(chan Out)
	errChan := make(chan error)

	wg := &sync.WaitGroup{}
	done := make(chan bool)

	for i := range inputs {
		input := inputs[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			output, err := task(ctx, input)
			if err != nil {
				errChan <- err
			}
			if output != nil {
				resChan <- output
			}
		}()
	}

	go func() {
		wg.Wait()
		done <- true
	}()

WaitLoop:
	for {
		select {
		case out := <-resChan:
			results = append(results, out)
		case err := <-errChan:
			accErr = append(accErr, err)
		case <-ctx.Done():
			accErr = append(accErr, ctx.Err())
			break WaitLoop
		case <-done:
			break WaitLoop
		}
	}

	if len(accErr) == 0 {
		accErr = nil
	}
	return results, accErr
}

func RunTasksAndAccumulateErrors[In any](
	ctx context.Context,
	inputs []In,
	task func(context.Context, In) error,
) error {
	_, err := RunTasksAndAccumulate[In, struct{}](ctx, inputs,
		func(ctx context.Context, input In) (struct{}, error) {
			return struct{}{}, task(ctx, input)
		})
	return err
}

package master

import (
	"context"
	"sync"
)

func RunTasksAndAccumulate[Idx comparable, In, Out any](
	ctx context.Context,
	inputs Iterable[Idx, In],
	task func(context.Context, Idx, In) (Out, error),
) ([]Out, error) {
	results := make([]Out, 0, inputs.Len())
	accErr := make(Error, 0, inputs.Len())

	resChan := make(chan Out)
	errChan := make(chan error)

	wg := &sync.WaitGroup{}
	done := make(chan bool)

	for inputs.Next() {
		input := inputs.Value()
		idx := inputs.Id()
		wg.Add(1)
		go func() {
			defer wg.Done()
			output, err := task(ctx, idx, input)
			if err != nil {
				errChan <- err
			}
			resChan <- output
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
		return results, nil
	}
	return results, accErr
}

func RunTasksAndAccumulateErrors[Idx comparable, In any](
	ctx context.Context,
	inputs Iterable[Idx, In],
	task func(context.Context, Idx, In) error,
) error {
	_, err := RunTasksAndAccumulate[Idx, In, struct{}](ctx, inputs,
		func(ctx context.Context, idx Idx, input In) (struct{}, error) {
			return struct{}{}, task(ctx, idx, input)
		})
	return err
}

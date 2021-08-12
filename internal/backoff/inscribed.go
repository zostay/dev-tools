package backoff

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type InscribedState struct {
	failures int
}

func Inscribed(done *sync.WaitGroup, state *InscribedState, doit func() error) func() {
	var (
		backoffTime = 1 * time.Second << state.failures
		quitOnce    = new(sync.Once)
		quit        = make(chan struct{})
		quitter     = func() {
			quitOnce.Do(func() { quit <- struct{}{} })
		}
	)

	done.Add(1)
	go func() {
		defer done.Done()

		err := doit()
		if err == nil {
			go quitter()
			return
		}

		stop := false
		if state.failures > 0 {
			fmt.Fprintf(os.Stderr, "Try again in %v\n", backoffTime)

			select {
			case <-time.After(backoffTime):
			case <-quit:
				stop = true
			}
		}

		if stop {
			return
		}

		err = doit()
		if err != nil {
			state.failures++
			return
		}

		state.failures = 0
	}()

	return quitter
}

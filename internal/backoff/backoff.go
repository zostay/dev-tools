package backoff

import (
	"fmt"
	"os"
	"sync"
	"time"
)

func Generic(done *sync.WaitGroup, doit func() error) func() {
	var (
		backoffTime = 1 * time.Second
		quitOnce    = new(sync.Once)
		quit        = make(chan struct{})
		quitter     = func() {
			quitOnce.Do(func() { quit <- struct{}{} })
		}
	)

	done.Add(1)
	go func() {
		defer done.Done()
		for {
			err := doit()
			if err == nil {
				go quitter()
				return
			}

			fmt.Fprintf(os.Stderr, "Try again in %v\n", backoffTime)

			select {
			case <-time.After(backoffTime):
				continue
			case <-quit:
				return
			}

			if backoffTime < 15*time.Second {
				backoffTime = backoffTime * 2
			}
		}
	}()

	return quitter
}

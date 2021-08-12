// Package acmd provides an asynchronous command interface that automatically
// retries to start the command until it runs. It is intended to be pluggable
// for specializing.
package acmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/zostay/dev-tools/internal/backoff"
	"github.com/zostay/dev-tools/pkg/future"
)

type state int

const (
	stateStart state = iota + 1
	stateReady
	stateStarted
	stateQuit
	stateKill
)

type event struct {
	state state

	err error
	cmd *exec.Cmd
}

type StartHandler func() error
type ReadyHandler func(cmd *exec.Cmd) error
type StartedHandler func(cmd *exec.Cmd) error
type StopHandler func(error)

type Cmd struct {
	StartHandler   StartHandler
	ReadyHandler   ReadyHandler
	StartedHandler StartedHandler
	StopHandler    StopHandler

	cmdLine []string
	done    *sync.WaitGroup
	result  *future.DeferredPromise

	quitter       func(error)
	events        chan event
	readyState    backoff.InscribedState
	eventlooponce *sync.Once
	startonce     *sync.Once
	stoponce      *sync.Once
}

func Command(cmdLine []string, done *sync.WaitGroup) (*Cmd, error) {
	if len(cmdLine) < 1 {
		return nil, errors.New("cmdLine is too short")
	}

	c := Cmd{
		cmdLine:       cmdLine,
		done:          done,
		result:        future.Deferred(),
		events:        make(chan event),
		eventlooponce: new(sync.Once),
		startonce:     new(sync.Once),
		stoponce:      new(sync.Once),
	}

	c.makeQuitter(func() {})

	return &c, nil
}

func (c *Cmd) makeQuitter(q func()) {
	c.quitter = func(err error) {
		q()
		c.result.Keep(nil, err)
	}
}

func (c *Cmd) Get() (interface{}, error) {
	return c.result.Get()
}

func (c *Cmd) Then(f future.Followup) future.Promise {
	return c.result.Then(f)
}

func (c *Cmd) String() string {
	return strings.Join(c.cmdLine, " ")
}

func (c *Cmd) eventloop() {
	c.eventlooponce.Do(func() {
		c.done.Add(1)
		go func() {
			defer c.done.Done()
			for {
				e := <-c.events
				if quit := c.handle(&e); quit {
					return
				}
			}
		}()
	})
}

func (c *Cmd) Run() error {
	c.eventloop()
	c.Start()
	return c.Wait()
}

func (c *Cmd) Start() {
	c.eventloop()
	c.startonce.Do(func() {
		c.done.Add(1)
		go func() {
			defer c.done.Done()
			c.events <- event{
				state: stateStart,
			}
		}()
	})
}

func (c *Cmd) Wait() error {
	_, err := c.result.Get()
	return err
}

func (c *Cmd) Stop() {
	c.stoponce.Do(func() {
		c.done.Add(1)
		go func() {
			defer c.done.Done()
			c.events <- event{
				state: stateKill,
				err:   errors.New("stop called"),
			}
		}()
	})
}

func (c *Cmd) handle(e *event) bool {
	switch e.state {
	case stateStart:
		c.start()
	case stateReady:
		c.ready(e.cmd)
	case stateStarted:
		c.started(e.cmd)
	case stateKill:
		c.kill(e.err)
	case stateQuit:
		return false
	default:
		panic("unknown cmd state")
	}

	return true
}

func (c *Cmd) kill(err error) {
	c.quitter(err)
}

func (c *Cmd) start() {
	if c.StartHandler != nil {
		if err := c.StartHandler(); err != nil {
			c.done.Add(1)
			go func() {
				defer c.done.Done()
				c.result.Keep(nil, err)
				c.events <- event{
					state: stateQuit,
				}
			}()
			return
		}
	}

	q := backoff.Generic(c.done, func() error {
		cmd, err := c.buildCmd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Command %q failed restart: %v\n", c.String(), err)
			return err
		}

		c.events <- event{
			state: stateReady,
			cmd:   cmd,
		}

		return nil
	})

	c.makeQuitter(q)
}

func (c *Cmd) buildCmd() (*exec.Cmd, error) {
	fmt.Fprintf(os.Stderr, "Run> %s ...\n", c.String())

	var cmd *exec.Cmd
	if len(c.cmdLine) > 1 {
		cmd = exec.Command(c.cmdLine[0], c.cmdLine[1:]...)
	} else {
		cmd = exec.Command(c.cmdLine[0])
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd, nil
}

func (c *Cmd) ready(cmd *exec.Cmd) {
	if c.ReadyHandler != nil {
		if err := c.ReadyHandler(cmd); err != nil {
			c.done.Add(1)
			go func() {
				defer c.done.Done()
				c.result.Keep(nil, err)
				c.events <- event{
					state: stateQuit,
				}
			}()
			return
		}
	}

	q := backoff.Inscribed(c.done, &c.readyState, func() error {
		err := cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building with %q: %v", c.String(), err)
			return err
		}

		c.done.Add(1)
		go func() {
			defer c.done.Done()
			c.events <- event{
				state: stateStarted,
				cmd:   cmd,
			}
		}()

		return nil
	})

	c.makeQuitter(q)
}

func (c *Cmd) started(cmd *exec.Cmd) {
	once := new(sync.Once)
	q := func() {
		once.Do(func() {
			if err := cmd.Process.Signal(syscall.SIGHUP); err != nil {
				fmt.Fprintf(os.Stderr, "Error sending SIGHUP to %q: %v\n", c.String(), err)
				if err := cmd.Process.Kill(); err != nil {
					fmt.Fprintf(os.Stderr, "Unable to kill %q: %v\n", c.String(), err)
				}
			}
		})
	}

	c.done.Add(1)
	go func() {
		defer c.done.Done()
		err := cmd.Wait()
		c.result.Keep(struct{}{}, err)
	}()

	// from this point on, the user supplied error is ignored as we prefer to
	// get the error from .Wait()
	c.quitter = func(_ error) { q() }

	if c.StartedHandler != nil {
		if err := c.StartedHandler(cmd); err != nil {
			c.done.Add(1)
			go func() {
				defer c.done.Done()
				c.stoponce.Do(func() {
					c.events <- event{
						state: stateKill,
						err:   err,
					}
				})
			}()
			return
		}
	}
}

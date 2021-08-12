package server

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/zostay/dev-tools/internal/fswatch"
	"github.com/zostay/dev-tools/pkg/acmd"
	"github.com/zostay/dev-tools/pkg/config"
)

type state int

const (
	stateStart state = iota + 1
	stateRebuild
	stateRestart
	stateChange
	stateKill
)

type event struct {
	state state

	name string
}

type Worker struct {
	config *config.WebTarget
	done   *sync.WaitGroup

	addrs     chan net.Addr
	addrMatch *regexp.Regexp

	events chan event

	fsevents chan fswatch.Event
	fserrors chan error

	paused   bool
	quitting bool

	builder *acmd.Cmd
	runner  *RunCmd

	quitters []func()
}

func NewWorker(
	config *config.WebTarget,
	done *sync.WaitGroup,
) (*Worker, error) {
	addrMatch, err := regexp.Compile(config.ServerAddressMatch)
	if err != nil {
		return nil, err
	}

	w := Worker{
		config: config,
		done:   done,

		events: make(chan event),

		addrs:     make(chan net.Addr),
		addrMatch: addrMatch,

		fsevents: make(chan fswatch.Event),
		fserrors: make(chan error),

		paused: true,

		quitters: make([]func(), 0),
	}

	return &w, nil
}

func (w *Worker) setupBuilder() {
	if len(w.config.Build) == 0 {
		return
	}

	var err error
	w.builder, err = acmd.Command(w.config.Build, w.done)
	if err != nil {
		panic(err)
	}

	w.builder.Start()
	w.done.Add(1)
	go func() {
		defer w.done.Done()
		err := w.builder.Wait()
		if err != nil {
			<-time.After(5 * time.Second)
			w.events <- event{
				state: stateRebuild,
			}
		} else {
			w.events <- event{
				state: stateRestart,
			}
		}
	}()
}

func (w *Worker) setupRunner() {
	if len(w.config.Run) == 0 {
		return
	}

	var err error
	w.runner, err = RunCommand(w.config.Run, w.done, w.addrMatch)
	if err != nil {
		panic(err)
	}

	w.runner.Start()
	w.done.Add(2)
	go func() {
		addr, err := w.runner.Addr()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading server address: %v", err)
			return
		}
		w.addrs <- addr
	}()
	go func() {
		defer w.done.Done()
		err := w.runner.Wait()
		if err != nil && err.Error() != "signal: hangup" {
			fmt.Fprintf(os.Stderr, "unexpected quit: %v", err)
			<-time.After(5 * time.Second)
		} else {
			<-time.After(100 * time.Millisecond)
		}

		w.events <- event{
			state: stateRestart,
		}
	}()
}

func (w *Worker) Start() {
	w.done.Add(1)
	go func() {
		defer w.done.Done()
		for {
			select {
			case e := <-w.events:
				if quit := w.handle(&e); quit {
					return
				}

			case e := <-w.fsevents:
				w.done.Add(1)
				go func() {
					defer w.done.Done()
					w.events <- event{
						state: stateChange,
						name:  e.Name,
					}
				}()

			case err := <-w.fserrors:
				fmt.Fprintf(os.Stderr, "FSNotify Error: %v\n", err)
			}
		}
	}()

	w.events <- event{
		state: stateStart,
	}
}

func (w *Worker) handle(e *event) bool {
	switch e.state {
	case stateStart:
		w.setupBuilder()
		w.setupRunner()

	case stateChange:
		w.change(e.name)

	case stateRebuild:
		w.rebuild()

	case stateRestart:
		w.restart()

	case stateKill:
		w.kill()

	default:
		panic("unknown worker state")
	}
	return false
}

func (w *Worker) rebuild() {
	w.setupBuilder()
}

func (w *Worker) restart() {
	w.setupRunner()
}

func (w *Worker) change(name string) {
	if w.paused {
		return
	}

	w.paused = true
	fmt.Fprintf(os.Stderr, "FSNotify detected change: %s\n", name)

	w.done.Add(1)
	go func() {
		defer w.done.Done()
		w.events <- event{
			state: stateRebuild,
		}
	}()
}

func (w *Worker) kill() {
	w.quitting = true

	w.builder.Stop()
	w.runner.Stop()
}

func (w *Worker) AddrListener() chan net.Addr {
	return w.addrs
}

func (w *Worker) EventsListener() chan fswatch.Event {
	return w.fsevents
}

func (w *Worker) ErrorsListener() chan error {
	return w.fserrors
}

func (w *Worker) RegisterQuitter(q func()) {
	w.quitters = append(w.quitters, q)
}

func (w *Worker) Quit() {
	go func() {
		w.events <- event{
			state: stateKill,
		}
	}()
}

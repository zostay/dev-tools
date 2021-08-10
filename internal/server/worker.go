package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/zostay/dev-tools/internal/backoff"
	"github.com/zostay/dev-tools/internal/fswatch"
	"github.com/zostay/dev-tools/pkg/config"
)

type Worker struct {
	config           *config.WebTarget
	done             *sync.WaitGroup
	quit             chan struct{}
	fsevents         chan fswatch.Event
	fserrors         chan error
	paused           bool
	whilePaused      int
	rebuild          chan struct{}
	rebuilt          chan error
	runStart         chan struct{}
	runQuit          chan error
	runCmd           *exec.Cmd
	buildCmd         *exec.Cmd
	backoffTime      time.Duration
	buildBackoffQuit func()
	runBackoffQuit   func()
	limbo            sync.Locker
	quitters         []func()
	recvAddr         chan net.Addr
	addr             net.Addr
}

func NewWorker(
	config *config.WebTarget,
	done *sync.WaitGroup,
) *Worker {
	return &Worker{
		config:           config,
		done:             done,
		quit:             make(chan struct{}),
		fsevents:         make(chan fswatch.Event),
		fserrors:         make(chan error),
		paused:           true,
		whilePaused:      0,
		rebuild:          make(chan struct{}),
		rebuilt:          make(chan error),
		runStart:         make(chan struct{}),
		runQuit:          make(chan error),
		runCmd:           nil,
		buildCmd:         nil,
		backoffTime:      1 * time.Second,
		buildBackoffQuit: nil,
		runBackoffQuit:   nil,
		limbo:            new(sync.Mutex),
		quitters:         make([]func(), 0),
		recvAddr:         make(chan net.Addr),
		addr:             nil,
	}
}

func (w *Worker) RegisterQuitter(q func()) {
	w.quitters = append(w.quitters, q)
}

func (w *Worker) EventsListener() chan fswatch.Event {
	return w.fsevents
}

func (w *Worker) ErrorsListener() chan error {
	return w.fserrors
}

func (w *Worker) Run() {
	defer w.done.Done()

	go func() {
		// trigger startup for build a run
		w.rebuild <- struct{}{}
		w.runStart <- struct{}{}
	}()

	for {
		select {

		// on this signal we start a build
		case <-w.rebuild:
			w.startRebuild()

		// build is complete, so now we need to restart the runner
		case err := <-w.rebuilt:
			w.completeRebuild(err)

		// we received a file system event, so trigger a new build
		case event := <-w.fsevents:
			w.triggerRebuild(event)

		// show file system event errors
		case err := <-w.fserrors:
			fmt.Fprintf(os.Stderr, "FSNotify Error: %v\n", err)

		// this is going to start the server
		case <-w.runStart:
			w.startRun()

		// this triggers when the server quits to trigger restart
		case err := <-w.runQuit:
			w.completeRun(err)

		// we've been told to quit, so shutdown
		case <-w.quit:
			w.shutdown()
			return
		}
	}
}

func (w *Worker) Addr() net.Addr {
	if w.addr == nil {
		w.addr = <-w.recvAddr
	}

	return w.addr
}

func (w *Worker) Quit() {
	w.quit <- struct{}{}
}

func (w *Worker) runLineStr() string {
	return strings.Join(w.config.Run, " ")
}

func (w *Worker) buildLineStr() string {
	return strings.Join(w.config.Build, " ")
}

func (w *Worker) startRebuild() {
	if len(w.config.Build) == 0 {
		// this configuration has no build process
		return
	}

	w.limbo.Lock()
	w.buildBackoffQuit = backoff.Generic(w.done, func() error {
		var err error
		w.buildCmd, err = w.startCmd(w.config.Build)
		if err != nil {
			w.buildCmd = nil
			fmt.Fprintf(os.Stderr, "Error building with %q: %v", w.buildLineStr(), err)
			return err
		}

		w.done.Add(1)
		go func() {
			defer w.done.Done()
			w.rebuilt <- w.buildCmd.Wait()
		}()
		w.limbo.Unlock()

		return nil
	})
}

func (w *Worker) completeRebuild(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during build %q: %v", w.buildLineStr(), err)
	}

	//if w.whilePaused > 0 {
	//	w.whilePaused = 0
	//	go func() {
	//		<-time.After(1 * time.Second)
	//		w.rebuild <- struct{}{}
	//	}()
	//} else {
	w.paused = false
	w.buildCmd = nil
	w.quitCmd(w.runCmd)
	//}
}

func (w *Worker) triggerRebuild(event fswatch.Event) {
	if w.paused {
		w.whilePaused++
		return
	}

	w.paused = true
	fmt.Fprintf(os.Stderr, "FSNotify detected change: %s\n", event.Name)

	go func() { w.rebuild <- struct{}{} }()
}

func (w *Worker) startRun() {
	w.limbo.Lock()
	w.runBackoffQuit = backoff.Generic(w.done, func() error {
		var err error
		w.runCmd, err = w.startCmd(w.config.Run)
		if err != nil {
			w.runCmd = nil
			fmt.Fprintf(os.Stderr, "Command %q failed restart: %v\n", w.runLineStr(), err)
			return err
		}

		w.done.Add(1)
		go func() {
			defer w.done.Done()
			w.runQuit <- w.runCmd.Wait()
		}()
		w.limbo.Unlock()

		return nil
	})
}

func (w *Worker) completeRun(err error) {
	w.runCmd = nil
	if err != nil && err.Error() != "signal: hangup" {
		fmt.Fprintf(os.Stderr, "Command %q quit prematurely: %v\n", w.runLineStr(), err)
		if w.backoffTime < 15*time.Second {
			w.backoffTime = 2 * w.backoffTime
		}
	} else {
		w.backoffTime = 1 * time.Second
	}

	fmt.Fprintf(os.Stderr, "Restarting %q in %v ...\n", w.runLineStr(), w.backoffTime)
	go func() {
		<-time.After(w.backoffTime)
		w.runStart <- struct{}{}
	}()
}

func (w *Worker) shutdown() {
	w.limbo.Lock()
	if w.buildBackoffQuit != nil {
		w.buildBackoffQuit()
	}
	if w.runBackoffQuit != nil {
		w.runBackoffQuit()
	}
	w.quitCmd(w.runCmd)
	w.quitCmd(w.buildCmd)
	for _, quitter := range w.quitters {
		go quitter()
	}
	w.limbo.Unlock()
}

func (w *Worker) startCmd(cmdLine []string) (*exec.Cmd, error) {
	fmt.Fprintf(os.Stderr, "Run> %s ...\n", strings.Join(cmdLine, " "))

	var cmd *exec.Cmd
	if len(cmdLine) > 1 {
		cmd = exec.Command(cmdLine[0], cmdLine[1:]...)
	} else {
		cmd = exec.Command(cmdLine[0])
	}

	stdo, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stde, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdor := io.TeeReader(stdo, os.Stdout)
	stder := io.TeeReader(stde, os.Stderr)

	err = w.monitorForAddr(stdor, stder)
	if err != nil {
		return nil, err
	}

	err = cmd.Start()

	return cmd, err
}

type workerAddr struct {
	host string
}

func (wa *workerAddr) Network() string {
	return "tcp"
}

func (wa *workerAddr) String() string {
	return wa.host
}

func (w *Worker) monitorForAddr(rs ...io.Reader) error {
	if w.config.ServerAddressMatch == "" {
		return nil
	}

	m, err := regexp.Compile(w.config.ServerAddressMatch)
	if err != nil {
		return err
	}

	for _, r := range rs {
		s := bufio.NewScanner(r)
		go func(s *bufio.Scanner) {
			looking := true
			for s.Scan() {
				if looking {
					if gs := m.FindStringSubmatch(s.Text()); gs != nil {
						url, err := url.Parse(gs[1])
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error parsing URL %q to make address: %v", gs[1], err)
							continue
						}

						w.recvAddr <- &workerAddr{url.Host}
					}
				}
			}
		}(s)
	}

	return nil
}

func (w *Worker) quitCmd(cmd *exec.Cmd) {
	if cmd != nil {
		if err := cmd.Process.Signal(syscall.SIGHUP); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to hangup command: %v", err)
		}
	}
}

package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/pkg/config"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start API server(s)",
	RunE:  RunAPI,
}

func RunAPI(cmd *cobra.Command, args []string) error {
	config.Init(0)

	cfg, err := config.Get()
	if err != nil {
		return err
	}

	done := new(sync.WaitGroup)
	for _, target := range cfg.Web.Targets {
		if target.Type != config.APITarget {
			continue
		}

		if err := startAPITarget(done, &target); err != nil {
			go stopEverything()
			done.Wait()
			return err
		}
	}

	done.Wait()

	return nil
}

type worker struct {
	config           *config.WebTarget
	done             *sync.WaitGroup
	quit             chan struct{}
	fsevents         chan fsnotify.Event
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
}

func newWorker(
	config *config.WebTarget,
	done *sync.WaitGroup,
) *worker {
	return &worker{
		config:           config,
		done:             done,
		quit:             make(chan struct{}),
		fsevents:         make(chan fsnotify.Event),
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
	}
}

func (w *worker) run() {
	defer w.done.Done()

	go func() {
		// trigger startup for build a run
		w.rebuild <- struct{}{}
		w.runStart <- struct{}{}
	}()

	for {
		select {
		case <-w.rebuild:
			w.startRebuild()

		case err := <-w.rebuilt:
			w.completeRebuild(err)

		case event := <-w.fsevents:
			w.triggerRebuild(event)

		case err := <-w.fserrors:
			fmt.Fprintf(os.Stderr, "FSNotify Error: %v\n", err)

		case <-w.runStart:
			w.startRun()

		case err := <-w.runQuit:
			w.completeRun(err)

		case <-w.quit:
			w.shutdown()
			return
		}
	}
}

func (w *worker) runLineStr() string {
	return strings.Join(w.config.Run, " ")
}

func (w *worker) buildLineStr() string {
	return strings.Join(w.config.Build, " ")
}

func (w *worker) startRebuild() {
	w.limbo.Lock()
	w.buildBackoffQuit = backoff(w.done, func() error {
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

func (w *worker) completeRebuild(err error) {
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

func (w *worker) triggerRebuild(event fsnotify.Event) {
	if w.paused {
		w.whilePaused++
		return
	}

	w.paused = true
	fmt.Fprintf(os.Stderr, "FSNotify detected change: %s\n", event.Name)

	go func() { w.rebuild <- struct{}{} }()
}

func (w *worker) startRun() {
	w.limbo.Lock()
	w.runBackoffQuit = backoff(w.done, func() error {
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

func (w *worker) completeRun(err error) {
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

func (w *worker) shutdown() {
	w.limbo.Lock()
	if w.buildBackoffQuit != nil {
		w.buildBackoffQuit()
	}
	if w.runBackoffQuit != nil {
		w.runBackoffQuit()
	}
	w.quitCmd(w.runCmd)
	w.quitCmd(w.buildCmd)
	w.limbo.Unlock()
}

func (w *worker) startCmd(cmdLine []string) (*exec.Cmd, error) {
	fmt.Fprintf(os.Stderr, "Run> %s ...\n", strings.Join(cmdLine, " "))

	var cmd *exec.Cmd
	if len(cmdLine) > 1 {
		cmd = exec.Command(cmdLine[0], cmdLine[1:]...)
	} else {
		cmd = exec.Command(cmdLine[0])
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Start()

	return cmd, err
}

func (w *worker) quitCmd(cmd *exec.Cmd) {
	if cmd != nil {
		if err := cmd.Process.Signal(syscall.SIGHUP); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to hangup command: %v", err)
		}
	}
}

var workers = make([]worker, 0)

func startAPITarget(
	done *sync.WaitGroup,
	target *config.WebTarget,
) error {
	w := newWorker(target, done)

	for _, wcfg := range target.Watches {
		globs := make([]glob.Glob, len(wcfg.Filters))
		for i, f := range wcfg.Filters {
			g, err := glob.Compile(f)
			if err != nil {
				return err
			}

			globs[i] = g
		}

		err := setupFSNotifyRecursive(
			w,
			wcfg.Targets,
			globs,
			done,
		)

		if err != nil {
			return err
		}
	}

	done.Add(1)
	go w.run()

	return nil
}

func stopEverything() {
	for _, worker := range workers {
		worker.quit <- struct{}{}
	}
}

func backoff(done *sync.WaitGroup, doit func() error) func() {
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

func setupFSNotifyRecursive(
	w *worker,
	targets []string,
	globs []glob.Glob,
	done *sync.WaitGroup,
) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	for _, t := range targets {
		err := filepath.WalkDir(t, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() || path == t {
				if err := watcher.Add(path); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			watcher.Close()
			return err
		}
	}

	tlist := "[" + strings.Join(targets, ",") + "]"
	done.Add(1)
	go func() {
		defer done.Done()
		expected := false
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok && !expected {
					fmt.Fprintf(os.Stderr, "Unexpected clsoure of watcher event stream %s", tlist)
					return
				}

				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					if event.Op == fsnotify.Create || event.Op == fsnotify.Rename {
						watcher.Add(event.Name)
					}
				}

				if len(globs) > 0 {
					for _, g := range globs {
						if g.Match(event.Name) {
							w.fsevents <- event
						}
					}
				} else {
					w.fsevents <- event
				}

			case err, ok := <-watcher.Errors:
				if !ok && !expected {
					fmt.Fprintf(os.Stderr, "Unexpected closure of watcher error stream %s", tlist)
					return
				}
				w.fserrors <- err

			case <-w.quit:
				expected = true
				watcher.Close()
				return
			}
		}
	}()

	return nil
}

package fswatch

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"

	"github.com/zostay/dev-tools/pkg/config"
)

type Event fsnotify.Event

type Watcher interface {
	EventsListener() chan Event
	ErrorsListener() chan error
}

func SetupWatcher(w Watcher, done *sync.WaitGroup, config *config.FileWatch) (func(), error) {
	globs := make([]glob.Glob, len(config.Filters))
	for i, f := range config.Filters {
		g, err := glob.Compile(f)
		if err != nil {
			return nil, err
		}

		globs[i] = g
	}

	quitter, err := setupFSNotifyRecursive(
		w,
		config.Targets,
		globs,
		done,
	)

	return quitter, err
}

func setupFSNotifyRecursive(
	w Watcher,
	targets []string,
	globs []glob.Glob,
	done *sync.WaitGroup,
) (func(), error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
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
			return nil, err
		}
	}

	quit := make(chan struct{})
	tlist := "[" + strings.Join(targets, ",") + "]"
	done.Add(1)
	go func() {
		defer done.Done()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					fmt.Fprintf(os.Stderr, "Unexpected clsoure of watcher event stream %s\n", tlist)
					return
				}

				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					if event.Op == fsnotify.Create || event.Op == fsnotify.Rename {
						if err := watcher.Add(event.Name); err != nil {
							fmt.Fprintf(os.Stderr, "Failed to add new directory %q to event string %s: %v\n", event.Name, tlist, err)
							return
						}
					}
				}

				if len(globs) > 0 {
					for _, g := range globs {
						if g.Match(event.Name) {
							w.EventsListener() <- Event(event)
						}
					}
				} else {
					w.EventsListener() <- Event(event)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					fmt.Fprintf(os.Stderr, "Unexpected closure of watcher error stream %s\n", tlist)
					return
				}
				w.ErrorsListener() <- err

			case <-quit:
				watcher.Close()
				return
			}
		}
	}()

	return func() { quit <- struct{}{} }, nil
}

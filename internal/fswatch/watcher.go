package fswatch

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar"
	"github.com/fsnotify/fsnotify"

	"github.com/zostay/dev-tools/pkg/config"
)

// Event is just an fsnotify.Event.
type Event fsnotify.Event

// Watcher is a class that can receive event streams related to file system
// changes.
type Watcher interface {
	// EventsListener returns a channel related to the object that is ready to
	// receive file sytem change events.
	EventsListener() chan Event

	// ErrorsListener returns a channel related to the object that is ready to
	// receive file system change error events.
	ErrorsListener() chan error
}

// TODO Does this really need a WaitGroup? Do we really care about when the FS
// notifier has completely finished watching something? Should it be sent as
// part of the quitter instead or something?

// SetupWatcher is the entry point for setting up a watch on some set of files
// and to pass those events and errors onto the given Watcher. It returns a
// quit function, which can be used to tell the Watcher to exit and stop
// watching for file system changes. If an error is returned, the watcher is not
// setup and no events should be sent.
//
// If a sync.WaitGroup is given, then it will be notified when the watcher has
// quit.
func SetupWatcher(w Watcher, done *sync.WaitGroup, config *config.FileWatch) (func(), error) {
	// use a throw-away waitgroup if none is given.
	if done == nil {
		done = new(sync.WaitGroup)
	}

	quitter, err := setupFSNotifyRecursive(
		w,
		config.Targets,
		config.Filters,
		done,
	)

	return quitter, err
}

func setupFSNotifyRecursive(
	w Watcher,
	targets []string,
	patterns []string,
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

				if len(patterns) > 0 {
					for _, g := range patterns {
						matches, err := doublestar.PathMatch(g, event.Name)
						if err != nil {
							w.ErrorsListener() <- err
						}

						if matches {
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

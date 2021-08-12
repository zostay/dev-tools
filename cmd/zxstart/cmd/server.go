package cmd

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"sync"

	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/internal/fswatch"
	"github.com/zostay/dev-tools/internal/gohttp"
	"github.com/zostay/dev-tools/internal/server"
	"github.com/zostay/dev-tools/pkg/config"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start application server(s)",
	RunE:  RunServer,
}

func RunServer(cmd *cobra.Command, args []string) error {
	config.Init(0)

	cfg, err := config.Get()
	if err != nil {
		return err
	}

	done := new(sync.WaitGroup)
	for name, target := range cfg.Web.Targets {
		if target.Type != config.ServerTarget {
			continue
		}

		fmt.Fprintf(os.Stderr, "Starting worker %s ... \n", name)

		done.Add(1)
		s, err := startServerTarget(target, done)
		if err != nil {
			done.Done()
			go stopEverything()
			done.Wait()
			return err
		}

		workers[name] = s
	}

	for name, target := range cfg.Web.Targets {
		if target.Type != config.FrontendTarget {
			continue
		}

		fmt.Fprintf(os.Stderr, "Starting front-end %s ... \n", name)

		done.Add(1)
		s, err := startFrontendTarget(target, done, workers)
		if err != nil {
			done.Done()
			go stopEverything()
			done.Wait()
			return err
		}

		workers[name] = s
	}

	for _, target := range cfg.Web.Targets {
		if target.Type != config.FrontendTarget && target.Type != config.ServerTarget {
			fmt.Fprintf(os.Stderr, "Web target type %q is not supported.\n", target.Type)
		}
	}

	done.Wait()

	return nil
}

type Server interface {
	AddrListener() chan net.Addr
	Quit()
}

var workers = make(map[string]Server)

func startServerTarget(
	target config.WebTarget,
	done *sync.WaitGroup,
) (Server, error) {
	w, err := server.NewWorker(&target, done)
	if err != nil {
		return nil, err
	}

	for _, wcfg := range target.Watches {
		q, err := fswatch.SetupWatcher(w, done, &wcfg)
		if err != nil {
			stopEverything()
			return nil, err
		}

		w.RegisterQuitter(q)
	}

	w.Start()

	return w, nil
}

func startFrontendTarget(
	target config.WebTarget,
	done *sync.WaitGroup,
	workers map[string]Server,
) (Server, error) {
	f := gohttp.New(done)

	for _, dcfg := range target.Dispatch {
		var (
			tname = dcfg.Target
			path  = dcfg.Path
		)

		done.Add(1)
		go func(path string, w Server) {
			defer done.Done()
			for {
				addr := <-w.AddrListener()

				url := url.URL{
					Scheme: "http",
					Host:   addr.String(),
				}

				fmt.Fprintf(os.Stderr, "Updating proxy %q => %q\n", path, url.String())
				f.SetProxy(path, &url)
			}
		}(path, workers[tname])
	}

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	go func() {
		err := f.Serve(l)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gohttp server error: %v\n", err)
		}
	}()

	return f, nil
}

func stopEverything() {
	for _, worker := range workers {
		go worker.Quit()
	}
}

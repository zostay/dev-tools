package cmd

import (
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/internal/fswatch"
	"github.com/zostay/dev-tools/internal/gohttp"
	"github.com/zostay/dev-tools/internal/netx"
	"github.com/zostay/dev-tools/internal/server"
	"github.com/zostay/dev-tools/pkg/config"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start application server(s)",
	RunE:  RunServer,
}

// Server is the interface used to implement each type of application server
// that zxstart can manage.
type Server interface {
	// AddrListener tells the caller where to connect to this application
	// server.
	AddrListener() chan net.Addr

	// Start starts the application server.
	Start()

	// Quit tells teh application server to shutdown.
	Quit()
}

var logger = log.New(os.Stderr, "", 0)

type (
	InitServerFunc   func(config.WebTarget, *sync.WaitGroup) (Server, error)
	ConfigServerFunc func(string, config.WebTarget, *sync.WaitGroup, map[string]Server) error
)

func RunServer(cmd *cobra.Command, args []string) error {
	config.Init(0)

	cfg, err := config.Get()
	if err != nil {
		return err
	}

	var (
		workers = make(map[string]Server)
		done    = new(sync.WaitGroup)

		initialize InitServerFunc
		configure  ConfigServerFunc
	)

	// prepare
	for _, target := range cfg.Web.Targets {
		switch target.Type {
		// servers are complete app servers on their own port
		case config.ServerTarget:
			initialize = initServerTarget
			configure = configServerTarget

		// front-ends are ingress servers that route calls by prefix
		case config.FrontendTarget:
			initialize = initFrontendTarget
			configure = configFrontendTarget

		default:
			logger.Printf("Web target type %q is not supported.\n", target.Type)
			continue
		}
	}

	// init - construct and prep each server
	for name, target := range cfg.Web.Targets {
		s, err := initialize(target, done)
		if err != nil {
			go stopEverything(workers)
			done.Wait()
			return err
		}

		workers[name] = s
	}

	// config - tell each server what it is going to do
	for name, target := range cfg.Web.Targets {
		err := configure(name, target, done, workers)
		if err != nil {
			go stopEverything(workers)
			done.Wait()
			return err
		}
	}

	// process - start each server
	for name, server := range workers {
		logger.Printf("Starting server %s ... \n", name)

		server.Start()
	}

	// post-process - connect the server to the hoomin
	for name, target := range cfg.Web.Targets {
		if target.Type != config.FrontendTarget && target.Type != config.ServerTarget {
			logger.Printf("Web target type %q is not supported.\n", target.Type)
		}

		if target.OpenBrowser {
			s := workers[name]

			var openCmdName string
			switch runtime.GOOS {
			case "darwin":
				openCmdName = "open"
			case "linux":
				openCmdName = "xdg-open"
			default:
				panic("unsupported OS for open_browser")
			}

			go func(s Server) {
				for {
					addr := <-s.AddrListener()
					url, err := netx.AddrToURL(addr.String())
					if err != nil {
						logger.Printf("Failed to turn address %q into URL: %v", addr, err)
						return
					}

					logger.Printf("Opening browser to %q", url.String())

					openCmd := exec.Command(openCmdName, url.String())
					err = openCmd.Run()
					if err != nil {
						logger.Printf("Failed to open browser to %q: %v", addr, err)
					}
				}
			}(s)
		}
	}

	// sticks here
	done.Wait()

	return nil
}

func initServerTarget(
	target config.WebTarget,
	done *sync.WaitGroup,
) (Server, error) {
	w, err := server.NewWorker(logger, &target, done)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func configServerTarget(
	name string,
	target config.WebTarget,
	done *sync.WaitGroup,
	workers map[string]Server,
) error {
	w := workers[name].(*server.Worker)

	for _, wcfg := range target.Watches {
		q, err := fswatch.SetupWatcher(w, done, &wcfg)
		if err != nil {
			stopEverything(workers)
			return err
		}

		w.RegisterQuitter(q)
	}

	return nil
}

func initFrontendTarget(
	target config.WebTarget,
	done *sync.WaitGroup,
) (Server, error) {
	f := gohttp.New(done, logger)

	return f, nil
}

func configFrontendTarget(
	name string,
	target config.WebTarget,
	done *sync.WaitGroup,
	workers map[string]Server,
) error {
	f := workers[name].(*gohttp.Frontend)

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

				logger.Printf("Updating proxy %q => %q\n", path, url.String())
				f.SetProxy(path, &url)
			}
		}(path, workers[tname])
	}

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	done.Add(1)
	go func() {
		defer done.Done()
		err := f.Serve(l)
		if err != nil {
			logger.Printf("gohttp server error: %v\n", err)
		}
	}()

	return nil
}

func stopEverything(
	workers map[string]Server,
) {
	for _, worker := range workers {
		go worker.Quit()
	}
}

package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/internal/fswatch"
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
	for _, target := range cfg.Web.Targets {
		switch target.Type {
		case config.ServerTarget:
			if err := startServerTarget(done, &target); err != nil {
				go stopEverything()
				done.Wait()
				return err
			}

		case config.FrontendTarget:
			if err := startFrontendTarget(done, &target); err != nil {
				go stopEverything()
				done.Wait()
				return err
			}

		default:
			fmt.Fprintf(os.Stderr, "Web target type %q is not supported.\n", target.Type)
		}
	}

	done.Wait()

	return nil
}

var workers = make([]server.Worker, 0)

func startServerTarget(
	done *sync.WaitGroup,
	target *config.WebTarget,
) error {
	w := server.NewWorker(target, done)

	for _, wcfg := range target.Watches {
		q, err := fswatch.SetupWatcher(w, done, &wcfg)
		if err != nil {
			stopEverything()
			return err
		}

		w.RegisterQuitter(q)
	}

	done.Add(1)
	go w.Run()

	return nil
}

func startFrontendTarget(
	done *sync.WaitGroup,
	target *config.WebTarget,
) error {
	return nil
}

func stopEverything() {
	for _, worker := range workers {
		go worker.Quit()
	}
}

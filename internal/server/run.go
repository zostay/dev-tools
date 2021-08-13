package server

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sync"

	"github.com/zostay/dev-tools/pkg/acmd"
	"github.com/zostay/dev-tools/pkg/future"
)

type RunCmd struct {
	*acmd.Cmd

	AddrMatch *regexp.Regexp
	addr      *future.DeferredPromise
	logger    *log.Logger
}

func RunCommand(
	cmdLine []string,
	done *sync.WaitGroup,
	logger *log.Logger,
	addrMatch *regexp.Regexp,
) (*RunCmd, error) {
	c, err := acmd.Command(cmdLine, done, logger)
	if err != nil {
		return nil, err
	}

	r := RunCmd{
		Cmd:       c,
		AddrMatch: addrMatch,
		addr:      future.Deferred(),
		logger:    logger,
	}

	c.ReadyHandler = func(cmd *exec.Cmd) error {
		stdo, wo := io.Pipe()
		cmd.Stdout = wo

		stde, we := io.Pipe()
		cmd.Stderr = we

		stdor := io.TeeReader(stdo, os.Stdout)
		stder := io.TeeReader(stde, os.Stderr)

		err = r.monitorForAddr(stdor, stder)
		if err != nil {
			return err
		}

		return nil
	}

	return &r, nil
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

func (r *RunCmd) addrMatcher(s *bufio.Scanner) future.Actor {
	m := r.AddrMatch
	return func() (interface{}, error) {
		// TODO might want to apply a contextual timeout to limit how
		// long we wait for the address to show up.
		looking := true
		for s.Scan() {
			if looking {
				if gs := m.FindStringSubmatch(s.Text()); len(gs) == 2 {
					url, err := url.Parse(gs[1])
					if err != nil {
						r.logger.Printf("Error parsing URL %q to make address: %v", gs[1], err)
						return nil, err
					}

					return &workerAddr{url.Host}, nil
				}
			}
		}
		return nil, errors.New("address never found in server log output")
	}
}

func (r *RunCmd) monitorForAddr(rs ...io.Reader) error {
	for _, rd := range rs {
		s := bufio.NewScanner(rd)
		r.addr.When(
			future.Start(func(s *bufio.Scanner) future.Actor {
				return r.addrMatcher(s)
			}(s)),
		)
	}

	return nil
}

func (r *RunCmd) Addr() (net.Addr, error) {
	addr, err := r.addr.Get()
	if err != nil {
		return nil, err
	}
	return addr.(net.Addr), err
}

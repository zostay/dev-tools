package gohttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
)

type Frontend struct {
	addrs   chan net.Addr
	plock   *sync.RWMutex
	proxies map[string]*httputil.ReverseProxy
	s       *http.Server
	done    *sync.WaitGroup
}

func New(done *sync.WaitGroup) *Frontend {
	return &Frontend{
		addrs:   make(chan net.Addr),
		plock:   new(sync.RWMutex),
		proxies: make(map[string]*httputil.ReverseProxy),
		s:       new(http.Server),
		done:    done,
	}
}

func (f *Frontend) SetProxy(path string, to *url.URL) {
	f.plock.Lock()
	defer f.plock.Unlock()
	f.proxies[path] = httputil.NewSingleHostReverseProxy(to)
}

func (f *Frontend) Serve(l net.Listener) error {
	defer func() {
		go func(a net.Addr) {
			f.addrs <- a
		}(l.Addr())
	}()

	f.s.Handler = f.MakeHandler()
	return f.s.Serve(l)
}

func (f *Frontend) AddrListener() chan net.Addr {
	return f.addrs
}

func (f *Frontend) Quit() {
	err := f.s.Shutdown(context.TODO())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error shutting down gohttp: %v\n", err)
	}
}

func (f *Frontend) MakeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		proxy := f.FindBestProxy(path)

		if proxy == nil {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}

		proxy.ServeHTTP(w, r)
	}
}

func (f *Frontend) FindBestProxy(path string) *httputil.ReverseProxy {
	pp := strings.Split(path, "/")

	var bestlen = 0
	var bestproxy *httputil.ReverseProxy

	f.plock.RLock()
	defer f.plock.RUnlock()
	for trypath, proxy := range f.proxies {
		for i := len(pp); i > 0; i-- {
			if strings.Join(pp[:i], "/") == trypath {
				if bestlen < i {
					bestlen = i
					bestproxy = proxy
				}
				break
			}
		}
	}

	return bestproxy
}

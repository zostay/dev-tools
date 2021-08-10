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
	proxies map[string]*httputil.ReverseProxy
	l       net.Listener
	s       *http.Server
	done    *sync.WaitGroup
}

func New(done *sync.WaitGroup) *Frontend {
	return &Frontend{
		proxies: make(map[string]*httputil.ReverseProxy),
		s:       new(http.Server),
		done:    done,
	}
}

func (f *Frontend) AddProxy(path string, to *url.URL) {
	f.proxies[path] = httputil.NewSingleHostReverseProxy(to)
}

func (f *Frontend) Serve(l net.Listener) error {
	f.s.Handler = f.MakeHandler()
	return f.s.Serve(l)
}

func (f *Frontend) Addr() net.Addr {
	return f.l.Addr()
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

package gohttp

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

type proxyURL struct {
	to    *url.URL
	proxy *httputil.ReverseProxy
}

type Frontend struct {
	addrs   chan net.Addr
	plock   *sync.RWMutex
	proxies map[string]proxyURL
	s       *http.Server
	done    *sync.WaitGroup
	logger  *log.Logger
}

func New(done *sync.WaitGroup, logger *log.Logger) *Frontend {
	return &Frontend{
		addrs:   make(chan net.Addr),
		plock:   new(sync.RWMutex),
		proxies: make(map[string]proxyURL),
		s:       new(http.Server),
		done:    done,
		logger:  logger,
	}
}

func (f *Frontend) SetProxy(path string, to *url.URL) {
	f.plock.Lock()
	defer f.plock.Unlock()
	f.proxies[path] = proxyURL{
		to:    to,
		proxy: httputil.NewSingleHostReverseProxy(to),
	}
}

func (f *Frontend) Serve(l net.Listener) error {
	f.logger.Printf("Frontend is listening on %s ...", l.Addr().String())

	go func(a net.Addr) {
		f.addrs <- a
	}(l.Addr())

	f.s.Handler = f.MakeHandler()
	return f.s.Serve(l)
}

func (f *Frontend) AddrListener() chan net.Addr {
	return f.addrs
}

func (f *Frontend) Start() {
}

func (f *Frontend) Quit() {
	err := f.s.Shutdown(context.TODO())
	if err != nil {
		f.logger.Printf("Error shutting down gohttp: %v\n", err)
	}
}

type responseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
	wroteHeader  bool
}

func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriter) Write(bs []byte) (int, error) {
	bw, err := rw.ResponseWriter.Write(bs)
	rw.bytesWritten = rw.bytesWritten + bw
	return bw, err
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.wroteHeader = true
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (f *Frontend) MakeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ww := &responseWriter{ResponseWriter: w}

		path := r.URL.Path
		proxy := f.FindBestProxy(path)

		if proxy == nil {
			f.logger.Printf("%s %s [404] (- bytes) -> {nil}", r.Method, r.URL.String())
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}

		proxy.proxy.ServeHTTP(ww, r)

		if ww.wroteHeader {
			f.logger.Printf("%s %s [%d] (%d bytes) -> %s", r.Method, r.URL.String(), ww.status, ww.bytesWritten, proxy.to)
		} else {
			f.logger.Printf("%s %s [200] (%d bytes) -> %s", r.Method, r.URL.String(), ww.bytesWritten, proxy.to)
		}
	}
}

func (f *Frontend) FindBestProxy(path string) *proxyURL {
	pp := strings.Split(path, "/")

	bestlen := 0
	var bestproxy *proxyURL

	f.plock.RLock()
	defer f.plock.RUnlock()
	for trypath, proxy := range f.proxies {
		for i := len(pp); i > strings.Count(trypath, "/")-1; i-- {
			ppmatch := strings.Join(pp[:i], "/") + "/"
			if ppmatch == trypath {
				if bestlen < i {
					bestlen = i
					bestproxy = &proxy
				}
				break
			}
		}
	}

	return bestproxy
}

package netx

import (
	"fmt"
	"net"
	"net/url"
)

func AddrToHostPort(hostport string) (string, error) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return "", err
	}

	switch host {
	case "::":
		host = "::1"
	case "0.0.0.0":
		host = "127.0.0.1"
	}

	return net.JoinHostPort(host, port), nil
}

func AddrToURL(hostport string) (*url.URL, error) {
	hostport, err := AddrToHostPort(hostport)
	if err != nil {
		return nil, err
	}

	urlText := fmt.Sprintf("http://%s", hostport)

	return url.Parse(urlText)
}

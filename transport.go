package main

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"

	"h12.io/socks"
)

func elktailTransport() http.RoundTripper {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	if socksProxy := os.Getenv("ELKTAILSOCKS"); socksProxy != "" {
		dialSocksProxy := socks.Dial(socksProxy)
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialSocksProxy(network, addr)
		}
	}

	return tr
}

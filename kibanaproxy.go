/* Copyright (C) 2026 Zoran ǅelajlija
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details. */

package main

import (
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"h12.io/socks"
)

type KibanaProxyTransport struct {
	KibanaURL string
}

func (t *KibanaProxyTransport) baseTransport() http.RoundTripper {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	if socksProxy := os.Getenv("ELKTAILSOCKS"); socksProxy != "" {
		dialSocksProxy := socks.Dial(socksProxy)
		tr.Dial = func(network, addr string) (net.Conn, error) {
			return dialSocksProxy(network, addr)
		}
	}

	return tr
}

func (t *KibanaProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil && req.Body != http.NoBody {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}

	proxyURL, err := url.Parse(t.KibanaURL + "/api/console/proxy")
	if err != nil {
		return nil, err
	}

	q := proxyURL.Query()
	if req.URL.RawQuery != "" {
		q.Set("path", req.URL.Path+"?"+req.URL.RawQuery)
	} else {
		q.Set("path", req.URL.Path)
	}
	q.Set("method", req.Method)
	proxyURL.RawQuery = q.Encode()

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	proxyReq, err := http.NewRequest(http.MethodPost, proxyURL.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	for name, values := range req.Header {
		lower := strings.ToLower(name)
		if lower == "host" || lower == "accept-encoding" || lower == "content-length" || lower == "transfer-encoding" {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}
	proxyReq.Header.Set("kbn-xsrf", "true")

	return t.baseTransport().RoundTrip(proxyReq)
}

func NewKibanaProxyTransport(kibanaURL string) *KibanaProxyTransport {
	if !strings.HasPrefix(kibanaURL, "http") {
		kibanaURL = "https://" + kibanaURL
	}
	return &KibanaProxyTransport{
		KibanaURL: kibanaURL,
	}
}

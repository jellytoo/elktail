/* Copyright (C) 2026 Zoran ǅelajlija
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details. */

package main

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
)

type KibanaProxyTransport struct {
	Base      http.RoundTripper
	KibanaURL string
}

func (t *KibanaProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
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

	proxyReq, err := http.NewRequest(http.MethodPost, proxyURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	proxyReq.Header = req.Header.Clone()
	proxyReq.Header.Set("kbn-xsrf", "true")

	return t.Base.RoundTrip(proxyReq)
}

func NewKibanaProxyTransport(kibanaURL string) *KibanaProxyTransport {
	return &KibanaProxyTransport{
		Base: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		KibanaURL: kibanaURL,
	}
}

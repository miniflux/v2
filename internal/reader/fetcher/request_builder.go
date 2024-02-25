// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"crypto/tls"
	"encoding/base64"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultHTTPClientTimeout     = 20
	defaultHTTPClientMaxBodySize = 15 * 1024 * 1024
	defaultAcceptHeader          = "application/xml, application/atom+xml, application/rss+xml, application/rdf+xml, application/feed+json, text/html, */*;q=0.9"
)

type RequestBuilder struct {
	headers          http.Header
	clientProxyURL   string
	useClientProxy   bool
	clientTimeout    int
	withoutRedirects bool
	ignoreTLSErrors  bool
	disableHTTP2     bool
}

func NewRequestBuilder() *RequestBuilder {
	return &RequestBuilder{
		headers:       make(http.Header),
		clientTimeout: defaultHTTPClientTimeout,
	}
}

func (r *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	r.headers.Set(key, value)
	return r
}

func (r *RequestBuilder) WithETag(etag string) *RequestBuilder {
	if etag != "" {
		r.headers.Set("If-None-Match", etag)
	}
	return r
}

func (r *RequestBuilder) WithLastModified(lastModified string) *RequestBuilder {
	if lastModified != "" {
		r.headers.Set("If-Modified-Since", lastModified)
	}
	return r
}

func (r *RequestBuilder) WithUserAgent(userAgent string, defaultUserAgent string) *RequestBuilder {
	if userAgent != "" {
		r.headers.Set("User-Agent", userAgent)
	} else {
		r.headers.Set("User-Agent", defaultUserAgent)
	}
	return r
}

func (r *RequestBuilder) WithCookie(cookie string) *RequestBuilder {
	if cookie != "" {
		r.headers.Set("Cookie", cookie)
	}
	return r
}

func (r *RequestBuilder) WithUsernameAndPassword(username, password string) *RequestBuilder {
	if username != "" && password != "" {
		r.headers.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	}
	return r
}

func (r *RequestBuilder) WithProxy(proxyURL string) *RequestBuilder {
	r.clientProxyURL = proxyURL
	return r
}

func (r *RequestBuilder) UseProxy(value bool) *RequestBuilder {
	r.useClientProxy = value
	return r
}

func (r *RequestBuilder) WithTimeout(timeout int) *RequestBuilder {
	r.clientTimeout = timeout
	return r
}

func (r *RequestBuilder) WithoutRedirects() *RequestBuilder {
	r.withoutRedirects = true
	return r
}

func (r *RequestBuilder) DisableHTTP2(value bool) *RequestBuilder {
	r.disableHTTP2 = value
	return r
}

func (r *RequestBuilder) IgnoreTLSErrors(value bool) *RequestBuilder {
	r.ignoreTLSErrors = value
	return r
}

func (r *RequestBuilder) ExecuteRequest(requestURL string) (*http.Response, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		// Setting `DialContext` disables HTTP/2, this option forces the transport to try HTTP/2 regardless.
		ForceAttemptHTTP2: true,
		DialContext: (&net.Dialer{
			// Default is 30s.
			Timeout: 10 * time.Second,

			// Default is 30s.
			KeepAlive: 15 * time.Second,
		}).DialContext,

		// Default is 100.
		MaxIdleConns: 50,

		// Default is 90s.
		IdleConnTimeout: 10 * time.Second,

		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: r.ignoreTLSErrors,
		},
	}

	if r.disableHTTP2 {
		transport.ForceAttemptHTTP2 = false

		// https://pkg.go.dev/net/http#hdr-HTTP_2
		// Programs that must disable HTTP/2 can do so by setting [Transport.TLSNextProto] (for clients) or [Server.TLSNextProto] (for servers) to a non-nil, empty map.
		transport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	}

	if r.useClientProxy && r.clientProxyURL != "" {
		if proxyURL, err := url.Parse(r.clientProxyURL); err != nil {
			slog.Warn("Unable to parse proxy URL",
				slog.String("proxy_url", r.clientProxyURL),
				slog.Any("error", err),
			)
		} else {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &http.Client{
		Timeout: time.Duration(r.clientTimeout) * time.Second,
	}

	if r.withoutRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	client.Transport = transport

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = r.headers
	req.Header.Set("Accept", defaultAcceptHeader)
	req.Header.Set("Connection", "close")

	slog.Debug("Making outgoing request", slog.Group("request",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Any("headers", req.Header),
		slog.Bool("without_redirects", r.withoutRedirects),
		slog.Bool("with_proxy", r.useClientProxy),
		slog.String("proxy_url", r.clientProxyURL),
		slog.Bool("ignore_tls_errors", r.ignoreTLSErrors),
		slog.Bool("disable_http2", r.disableHTTP2),
	))

	return client.Do(req)
}

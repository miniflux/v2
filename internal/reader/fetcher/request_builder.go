// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"slices"
	"time"

	"miniflux.app/v2/internal/proxyrotator"
)

const (
	defaultHTTPClientTimeout     = 20
	defaultHTTPClientMaxBodySize = 15 * 1024 * 1024
	defaultAcceptHeader          = "application/xml, application/atom+xml, application/rss+xml, application/rdf+xml, application/feed+json, text/html, */*;q=0.9"
)

type RequestBuilder struct {
	headers          http.Header
	clientProxyURL   *url.URL
	clientTimeout    int
	useClientProxy   bool
	withoutRedirects bool
	ignoreTLSErrors  bool
	disableHTTP2     bool
	proxyRotator     *proxyrotator.ProxyRotator
	feedProxyURL     string
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

func (r *RequestBuilder) WithProxyRotator(proxyRotator *proxyrotator.ProxyRotator) *RequestBuilder {
	r.proxyRotator = proxyRotator
	return r
}

func (r *RequestBuilder) WithCustomApplicationProxyURL(proxyURL *url.URL) *RequestBuilder {
	r.clientProxyURL = proxyURL
	return r
}

func (r *RequestBuilder) UseCustomApplicationProxyURL(value bool) *RequestBuilder {
	r.useClientProxy = value
	return r
}

func (r *RequestBuilder) WithCustomFeedProxyURL(proxyURL string) *RequestBuilder {
	r.feedProxyURL = proxyURL
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
			Timeout:   10 * time.Second, // Default is 30s.
			KeepAlive: 15 * time.Second, // Default is 30s.
		}).DialContext,
		MaxIdleConns:    50,               // Default is 100.
		IdleConnTimeout: 10 * time.Second, // Default is 90s.
	}

	if r.ignoreTLSErrors {
		//  Add insecure ciphers if we are ignoring TLS errors. This allows to connect to badly configured servers anyway
		ciphers := slices.Concat(tls.CipherSuites(), tls.InsecureCipherSuites())
		cipherSuites := make([]uint16, 0, len(ciphers))
		for _, cipher := range ciphers {
			cipherSuites = append(cipherSuites, cipher.ID)
		}
		transport.TLSClientConfig = &tls.Config{
			CipherSuites:       cipherSuites,
			InsecureSkipVerify: true,
		}
	}

	if r.disableHTTP2 {
		transport.ForceAttemptHTTP2 = false

		// https://pkg.go.dev/net/http#hdr-HTTP_2
		// Programs that must disable HTTP/2 can do so by setting [Transport.TLSNextProto] (for clients) or [Server.TLSNextProto] (for servers) to a non-nil, empty map.
		transport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	}

	var clientProxyURL *url.URL

	switch {
	case r.feedProxyURL != "":
		var err error
		clientProxyURL, err = url.Parse(r.feedProxyURL)
		if err != nil {
			return nil, fmt.Errorf(`fetcher: invalid feed proxy URL %q: %w`, r.feedProxyURL, err)
		}
	case r.useClientProxy && r.clientProxyURL != nil:
		clientProxyURL = r.clientProxyURL
	case r.proxyRotator != nil && r.proxyRotator.HasProxies():
		clientProxyURL = r.proxyRotator.GetNextProxy()
	}

	var clientProxyURLRedacted string
	if clientProxyURL != nil {
		transport.Proxy = http.ProxyURL(clientProxyURL)
		clientProxyURLRedacted = clientProxyURL.Redacted()
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
	req.Header.Set("Accept-Encoding", "br, gzip")
	req.Header.Set("Accept", defaultAcceptHeader)
	req.Header.Set("Connection", "close")

	slog.Debug("Making outgoing request", slog.Group("request",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Any("headers", req.Header),
		slog.Bool("without_redirects", r.withoutRedirects),
		slog.Bool("use_app_client_proxy", r.useClientProxy),
		slog.String("client_proxy_url", clientProxyURLRedacted),
		slog.Bool("ignore_tls_errors", r.ignoreTLSErrors),
		slog.Bool("disable_http2", r.disableHTTP2),
	))

	return client.Do(req)
}

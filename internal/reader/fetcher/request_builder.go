// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultHTTPClientMaxRequestDuration = 60
	defaultHTTPClientMaxBodySize        = 15 * 1024 * 1024
)

type RequestBuilder struct {
	headers            http.Header
	clientProxyURL     string
	useClientProxy     bool
	maxRequestDuration int
	withoutRedirects   bool
	ignoreTLSErrors    bool
}

func NewRequestBuilder() *RequestBuilder {
	return &RequestBuilder{
		headers:            make(http.Header),
		maxRequestDuration: defaultHTTPClientMaxRequestDuration,
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

func (r *RequestBuilder) WithUserAgent(userAgent string) *RequestBuilder {
	if userAgent != "" {
		r.headers.Set("User-Agent", userAgent)
	} else {
		r.headers.Del("User-Agent")
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

func (r *RequestBuilder) WithMaxRequestDuration(timeout int) *RequestBuilder {
	r.maxRequestDuration = timeout
	return r
}

func (r *RequestBuilder) WithoutRedirects() *RequestBuilder {
	r.withoutRedirects = true
	return r
}

func (r *RequestBuilder) IgnoreTLSErrors(value bool) *RequestBuilder {
	r.ignoreTLSErrors = value
	return r
}

func (r *RequestBuilder) ExecuteRequest(requestURL string) (*http.Response, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: r.ignoreTLSErrors,
		},
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
		Transport: transport,
	}

	if r.withoutRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.maxRequestDuration)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = r.headers
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "close")

	slog.Debug("Making outgoing request", slog.Group("request",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Any("headers", req.Header),
		slog.Int("max_request_duration", r.maxRequestDuration),
		slog.Bool("without_redirects", r.withoutRedirects),
		slog.Bool("with_proxy", r.useClientProxy),
		slog.String("proxy_url", r.clientProxyURL),
	))

	return client.Do(req)
}

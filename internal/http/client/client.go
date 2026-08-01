// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package client // import "miniflux.app/v2/internal/http/client"

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
)

const defaultRequestTimeout = 10 * time.Second

// ErrPrivateNetwork is returned when a connection to a private network is blocked.
var ErrPrivateNetwork = errors.New("client: connection to private network is blocked")

// Options holds configuration for creating an HTTP client.
type Options struct {
	Timeout              time.Duration
	BlockPrivateNetworks bool
}

// NewClientWithOptions creates a new HTTP client with the specified options.
func NewClientWithOptions(opts Options) *http.Client {
	if !opts.BlockPrivateNetworks {
		return &http.Client{Timeout: opts.Timeout}
	}

	dialer := &net.Dialer{
		Timeout: opts.Timeout,
	}

	transport := &http.Transport{
		// The check is performed at connect time on the actual resolved IP, which eliminates TOCTOU / DNS-rebinding vulnerabilities.
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("client: unable to parse address %q: %w", addr, err)
			}

			ips, err := net.LookupIP(host)
			if err != nil {
				return nil, fmt.Errorf("client: unable to resolve host %q: %w", host, err)
			}

			var safeIP net.IP
			for _, ip := range ips {
				if !urllib.IsNonPublicIP(ip) {
					safeIP = ip
					break
				}
			}

			if safeIP == nil {
				return nil, fmt.Errorf("%w: host %q resolves to a non-public IP address", ErrPrivateNetwork, host)
			}

			safeAddr := net.JoinHostPort(safeIP.String(), port)
			return dialer.DialContext(ctx, network, safeAddr)
		},
	}

	return &http.Client{
		Timeout:   opts.Timeout,
		Transport: transport,
	}
}

// requestBuilder builds and executes HTTP requests with the builder pattern.
type requestBuilder struct {
	err      error
	endpoint string
	method   string
	body     io.Reader
	headers  http.Header
}

// NewRequestBuilder creates a new request builder for the given endpoint.
func NewRequestBuilder(endpoint string) *requestBuilder {
	return &requestBuilder{
		endpoint: endpoint,
		method:   http.MethodGet,
		headers:  make(http.Header),
	}
}

// WithMethod sets the HTTP method.
func (r *requestBuilder) WithMethod(method string) *requestBuilder {
	r.method = method
	return r
}

// WithHeader sets a header value.
func (r *requestBuilder) WithHeader(key, value string) *requestBuilder {
	r.headers.Set(key, value)
	return r
}

// WithJSON marshals payload as JSON, sets the body and Content-Type.
func (r *requestBuilder) WithJSON(payload any) *requestBuilder {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		r.err = fmt.Errorf("unable to encode request body: %w", err)
		return r
	}

	return r.WithJSONBody(requestBody)
}

// WithJSONBody sets an already-marshaled JSON body and the Content-Type.
// It is useful when the caller needs the encoded payload for another
// purpose (e.g. computing a signature) to avoid marshaling it twice.
func (r *requestBuilder) WithJSONBody(body []byte) *requestBuilder {
	r.body = bytes.NewReader(body)
	r.headers.Set("Content-Type", "application/json")
	return r
}

// Do builds and executes the request.
//
// Private networks are blocked unless explicitly allowed through the
// INTEGRATION_ALLOW_PRIVATE_NETWORKS option.
func (r *requestBuilder) Do() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}

	// The request is assembled lazily here rather than being stored as a
	// prebuilt *http.Request in the builder: http.NewRequest inspects the
	// body's concrete type (e.g. *bytes.Reader) to populate ContentLength and
	// GetBody. Constructing it only once the body is known yields a correct
	// Content-Length header and lets the client replay the body on redirects.
	req, err := http.NewRequest(r.method, r.endpoint, r.body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	for key, values := range r.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	req.Header.Set("User-Agent", "Miniflux/"+version.Version)

	clientOptions := Options{
		Timeout:              defaultRequestTimeout,
		BlockPrivateNetworks: !config.Opts.IntegrationAllowPrivateNetworks(),
	}

	response, err := NewClientWithOptions(clientOptions).Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %w", err)
	}

	return response, nil
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package client // import "miniflux.app/v2/internal/http/client"

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"miniflux.app/v2/internal/urllib"
)

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

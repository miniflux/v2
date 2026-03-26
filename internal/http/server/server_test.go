// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"testing"
)

func TestDetermineListenTargets(t *testing.T) {
	tests := []struct {
		name       string
		addresses  []string
		certDomain string
		certFile   string
		keyFile    string
		expected   []listenTarget
	}{
		{
			name:      "single HTTP listener",
			addresses: []string{":8080"},
			expected: []listenTarget{
				{address: ":8080", mode: modeHTTP},
			},
		},
		{
			name:      "multiple HTTP listeners",
			addresses: []string{":8080", ":9090"},
			expected: []listenTarget{
				{address: ":8080", mode: modeHTTP},
				{address: ":9090", mode: modeHTTP},
			},
		},
		{
			name:      "TLS with cert files",
			addresses: []string{":443"},
			certFile:  "/path/to/cert.pem",
			keyFile:   "/path/to/key.pem",
			expected: []listenTarget{
				{address: ":443", mode: modeTLS, certFile: "/path/to/cert.pem", keyFile: "/path/to/key.pem"},
			},
		},
		{
			name:      "cert file without key file falls back to HTTP",
			addresses: []string{":8080"},
			certFile:  "/path/to/cert.pem",
			expected: []listenTarget{
				{address: ":8080", mode: modeHTTP},
			},
		},
		{
			name:      "key file without cert file falls back to HTTP",
			addresses: []string{":8080"},
			keyFile:   "/path/to/key.pem",
			expected: []listenTarget{
				{address: ":8080", mode: modeHTTP},
			},
		},
		{
			name:       "autocert with :https address",
			addresses:  []string{":https"},
			certDomain: "example.com",
			expected: []listenTarget{
				{address: ":https", mode: modeAutocertTLS},
			},
		},
		{
			name:       "autocert with first address containing colon",
			addresses:  []string{":443"},
			certDomain: "example.com",
			expected: []listenTarget{
				{address: ":443", mode: modeAutocertTLS},
			},
		},
		{
			name:       "autocert does not apply to second non-https address",
			addresses:  []string{":https", ":8080"},
			certDomain: "example.com",
			expected: []listenTarget{
				{address: ":https", mode: modeAutocertTLS},
				{address: ":8080", mode: modeHTTP},
			},
		},
		{
			name:      "unix socket",
			addresses: []string{"/var/run/miniflux.sock"},
			expected: []listenTarget{
				{address: "/var/run/miniflux.sock", mode: modeUnixSocket},
			},
		},
		{
			name:      "unix socket with TLS",
			addresses: []string{"/var/run/miniflux.sock"},
			certFile:  "/path/to/cert.pem",
			keyFile:   "/path/to/key.pem",
			expected: []listenTarget{
				{address: "/var/run/miniflux.sock", mode: modeUnixSocketTLS, certFile: "/path/to/cert.pem", keyFile: "/path/to/key.pem"},
			},
		},
		{
			name:      "mixed unix socket and TCP",
			addresses: []string{"/var/run/miniflux.sock", ":8080"},
			certFile:  "/path/to/cert.pem",
			keyFile:   "/path/to/key.pem",
			expected: []listenTarget{
				{address: "/var/run/miniflux.sock", mode: modeUnixSocketTLS, certFile: "/path/to/cert.pem", keyFile: "/path/to/key.pem"},
				{address: ":8080", mode: modeTLS, certFile: "/path/to/cert.pem", keyFile: "/path/to/key.pem"},
			},
		},
		{
			name:      "empty address list",
			addresses: []string{},
			expected:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := determineListenTargets(tc.addresses, tc.certDomain, tc.certFile, tc.keyFile)

			if len(got) != len(tc.expected) {
				t.Fatalf("got %d targets, want %d", len(got), len(tc.expected))
			}

			for i := range got {
				if got[i] != tc.expected[i] {
					t.Errorf("target[%d] = %+v, want %+v", i, got[i], tc.expected[i])
				}
			}
		})
	}
}

func TestAnyTLS(t *testing.T) {
	tests := []struct {
		name     string
		targets  []listenTarget
		expected bool
	}{
		{
			name:     "empty list",
			targets:  nil,
			expected: false,
		},
		{
			name:     "HTTP only",
			targets:  []listenTarget{{mode: modeHTTP}},
			expected: false,
		},
		{
			name:     "systemd only",
			targets:  []listenTarget{{mode: modeSystemd}},
			expected: false,
		},
		{
			name:     "unix socket without TLS",
			targets:  []listenTarget{{mode: modeUnixSocket}},
			expected: false,
		},
		{
			name:     "TLS mode",
			targets:  []listenTarget{{mode: modeTLS}},
			expected: true,
		},
		{
			name:     "autocert TLS mode",
			targets:  []listenTarget{{mode: modeAutocertTLS}},
			expected: true,
		},
		{
			name:     "unix socket TLS mode",
			targets:  []listenTarget{{mode: modeUnixSocketTLS}},
			expected: true,
		},
		{
			name:     "mixed with one TLS",
			targets:  []listenTarget{{mode: modeHTTP}, {mode: modeTLS}, {mode: modeUnixSocket}},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := anyTLS(tc.targets); got != tc.expected {
				t.Errorf("anyTLS() = %v, want %v", got, tc.expected)
			}
		})
	}
}

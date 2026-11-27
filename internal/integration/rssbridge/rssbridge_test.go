// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rssbridge // import "miniflux.app/v2/internal/integration/rssbridge"

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"miniflux.app/v2/internal/config"
)

const testToken = "secret-token"

func TestDetectBridgesSendsAndAppendsToken(t *testing.T) {
	configureOptions(t, true)
	logs := captureDebugLogs(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("token"); got != testToken {
			t.Errorf("Expected token %q, got %q", testToken, got)
		}
		if got := r.URL.Query().Get("action"); got != "findfeed" {
			t.Errorf("Expected action %q, got %q", "findfeed", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"url":"./?action=display&bridge=Example&format=Atom","bridgeMeta":{"name":"Example"}}]`))
	}))
	defer server.Close()

	bridges, err := DetectBridges(server.URL+"/", testToken, "https://example.org/")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(bridges) != 1 {
		t.Fatalf("Expected 1 bridge, got %d", len(bridges))
	}

	expectedURL := server.URL + "/?action=display&bridge=Example&format=Atom&token=" + testToken
	if bridges[0].URL != expectedURL {
		t.Errorf("Expected bridge URL %q, got %q", expectedURL, bridges[0].URL)
	}

	if logs.Len() == 0 {
		t.Fatal("Expected debug logs to be written")
	}

	if strings.Contains(logs.String(), testToken) {
		t.Errorf("Expected logs not to contain the token, got:\n%s", logs.String())
	}
}

func TestDetectBridgesRedactsTokenInRequestError(t *testing.T) {
	tests := []struct {
		name                 string
		allowPrivateNetworks bool
	}{
		{name: "connection refused", allowPrivateNetworks: true},
		{name: "private network blocked", allowPrivateNetworks: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configureOptions(t, tt.allowPrivateNetworks)
			logs := captureDebugLogs(t)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			server.Close()

			_, err := DetectBridges(server.URL+"/", testToken, "https://example.org/")
			if err == nil {
				t.Fatal("Expected an error")
			}

			if strings.Contains(err.Error(), testToken) {
				t.Errorf("Expected error not to contain the token, got %q", err)
			}

			if !strings.Contains(err.Error(), "token=xxxxx") {
				t.Errorf("Expected error to contain the redacted token, got %q", err)
			}

			if strings.Contains(logs.String(), testToken) {
				t.Errorf("Expected logs not to contain the token, got:\n%s", logs.String())
			}
		})
	}
}

func TestRedactedURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "token",
			input:    "https://bridge.example.org/?action=findfeed&token=" + testToken,
			expected: "https://bridge.example.org/?action=findfeed&token=xxxxx",
		},
		{
			name:     "no token",
			input:    "https://bridge.example.org/?action=findfeed&format=atom",
			expected: "https://bridge.example.org/?action=findfeed&format=atom",
		},
		{
			name:     "userinfo password",
			input:    "https://user:password@bridge.example.org/?token=" + testToken,
			expected: "https://user:xxxxx@bridge.example.org/?token=xxxxx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf("Unable to parse URL: %v", err)
			}

			if got := redactedURL(u); got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}

			if u.String() != tt.input {
				t.Errorf("Expected the original URL to be unchanged, got %q", u.String())
			}
		})
	}
}

func configureOptions(t *testing.T, allowPrivateNetworks bool) {
	t.Helper()

	if allowPrivateNetworks {
		t.Setenv("INTEGRATION_ALLOW_PRIVATE_NETWORKS", "1")
	} else {
		t.Setenv("INTEGRATION_ALLOW_PRIVATE_NETWORKS", "0")
	}

	configParser := config.NewConfigParser()
	parsedOptions, err := configParser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unable to configure test options: %v", err)
	}

	previousOptions := config.Opts
	config.Opts = parsedOptions
	t.Cleanup(func() {
		config.Opts = previousOptions
	})
}

func captureDebugLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buffer bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buffer, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() {
		slog.SetDefault(previousLogger)
	})

	return &buffer
}

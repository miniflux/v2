// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package discord

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

func TestSendDiscordMsgErrorsDoNotContainWebhookURL(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	const secret = "api/webhooks/123/0123456789abcdef"

	closeConnection := func(t *testing.T, w http.ResponseWriter) {
		connection, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Errorf("unable to hijack connection: %v", err)
			return
		}
		connection.Close()
	}

	tests := []struct {
		name            string
		handler         func(t *testing.T, w http.ResponseWriter)
		wantErrContains string
	}{
		{
			name: "connection fails",
			handler: func(t *testing.T, w http.ResponseWriter) {
				closeConnection(t, w)
			},
			wantErrContains: "discord: unable to send request",
		},
		{
			name: "error status",
			handler: func(t *testing.T, w http.ResponseWriter) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErrContains: "discord: unable to send a notification: status=500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.handler(t, w)
			}))
			defer server.Close()

			feed := &model.Feed{ID: 1, Title: "Feed", SiteURL: "https://example.org/"}
			entries := model.Entries{&model.Entry{ID: 1, Title: "Entry", URL: "https://example.org/entry"}}

			err := NewClient(server.URL+"/"+secret).SendDiscordMsg(feed, entries)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}

			if strings.Contains(err.Error(), secret) {
				t.Fatalf("expected the error to not contain the secret, got %q", err)
			}

			if !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Fatalf("expected the error to contain %q, got %q", tt.wantErrContains, err)
			}
		})
	}
}

func configureIntegrationAllowPrivateNetworksOption(t *testing.T) {
	t.Helper()

	t.Setenv("INTEGRATION_ALLOW_PRIVATE_NETWORKS", "1")

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

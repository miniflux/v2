// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

var webhookURLSecrets = []string{"secret-password", "secret-id", "secret-token"}

func secretWebhookURL(serverURL string) string {
	return strings.Replace(serverURL, "://", "://user:secret-password@", 1) + "/api/webhook/secret-id?token=secret-token"
}

func TestSendNewEntriesWebhookEventErrorsDoNotContainWebhookURL(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	closeConnection := func(t *testing.T, writer http.ResponseWriter) {
		connection, _, err := writer.(http.Hijacker).Hijack()
		if err != nil {
			t.Errorf("unable to hijack connection: %v", err)
			return
		}
		connection.Close()
	}

	scenarios := []struct {
		name            string
		handler         func(t *testing.T, writer http.ResponseWriter)
		wantErrContains string
	}{
		{
			name: "connection fails",
			handler: func(t *testing.T, writer http.ResponseWriter) {
				closeConnection(t, writer)
			},
			wantErrContains: "webhook: unable to send request",
		},
		{
			name: "error status",
			handler: func(t *testing.T, writer http.ResponseWriter) {
				writer.WriteHeader(http.StatusInternalServerError)
			},
			wantErrContains: "webhook: incorrect response status code 500",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				scenario.handler(t, writer)
			}))
			defer server.Close()

			feed := &model.Feed{ID: 1, Title: "Feed", Category: &model.Category{ID: 1, Title: "Category"}}
			entries := model.Entries{&model.Entry{ID: 1, Title: "Entry", URL: "https://example.org/entry"}}

			err := NewClient(secretWebhookURL(server.URL), "signing-secret").SendNewEntriesWebhookEvent(feed, entries)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}

			for _, secret := range webhookURLSecrets {
				if strings.Contains(err.Error(), secret) {
					t.Errorf("expected the error to not contain %q, got %q", secret, err)
				}
			}

			if !strings.Contains(err.Error(), scenario.wantErrContains) {
				t.Errorf("expected the error to contain %q, got %q", scenario.wantErrContains, err)
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

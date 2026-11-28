// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apprise

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

func TestSendNotificationDoesNotLogServicesURL(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	const servicesURL = "tgram://123456:bot-secret-token/chat-id, mailto://user:mail-secret@example.org"

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var payload struct {
			URLs string `json:"urls"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("unable to decode request body: %v", err)
		}
		if payload.URLs != servicesURL {
			t.Errorf("expected the services URL %q in the request body, got %q", servicesURL, payload.URLs)
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	feed := &model.Feed{ID: 1, Title: "Feed"}
	entries := model.Entries{&model.Entry{ID: 1, Title: "Entry", URL: "https://example.org/entry"}}

	if err := NewClient(servicesURL, server.URL).SendNotification(feed, entries); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(logs.String(), "Sending Apprise notification") {
		t.Fatalf("expected the notification to be logged, got:\n%s", logs.String())
	}

	for _, secret := range []string{"bot-secret-token", "mail-secret"} {
		if strings.Contains(logs.String(), secret) {
			t.Errorf("expected the logs to not contain %q, got:\n%s", secret, logs.String())
		}
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

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

func TestSendEntryLogsLinkwardenCollectionID(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	prev := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(prev)

	entry := &model.Entry{ID: 52, URL: "https://example.org/test.html", Title: "Test"}
	coll := int64(12345)
	userIntegrations := &model.Integration{
		UserID:                 1,
		LinkwardenEnabled:      true,
		LinkwardenCollectionID: &coll,
		LinkwardenURL:          "",
		LinkwardenAPIKey:       "",
	}

	SendEntry(entry, userIntegrations)

	out := buf.String()
	if !strings.Contains(out, `"collection_id":12345`) {
		t.Fatalf("expected collection_id in logs; got: %s", out)
	}
}

func TestSendEntryLogsLinkwardenWithoutCollectionID(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	prev := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(prev)

	entry := &model.Entry{ID: 52, URL: "https://example.org/test.html", Title: "Test"}
	userIntegrations := &model.Integration{
		UserID:            1,
		LinkwardenEnabled: true,
		LinkwardenURL:     "",
		LinkwardenAPIKey:  "",
	}

	SendEntry(entry, userIntegrations)

	out := buf.String()
	if strings.Contains(out, "collection_id") {
		t.Fatalf("did not expect collection_id in logs; got: %s", out)
	}
}

func TestSendEntryDoesNotLogWebhookURL(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)
	logs := captureDebugLogs(t)
	webhookURL := newFailingWebhookURL(t)

	entry := &model.Entry{
		ID:    52,
		URL:   "https://example.org/test.html",
		Title: "Test",
		Feed:  &model.Feed{ID: 1, Category: &model.Category{ID: 1}},
	}
	userIntegrations := &model.Integration{UserID: 1, WebhookEnabled: true, WebhookURL: webhookURL}

	SendEntry(entry, userIntegrations)

	assertLogsDoNotContainWebhookURL(t, logs, "Unable to send entry to Webhook")
}

func TestPushEntriesDoesNotLogWebhookURL(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)
	logs := captureDebugLogs(t)
	webhookURL := newFailingWebhookURL(t)

	feed := &model.Feed{ID: 1, Category: &model.Category{ID: 1}, WebhookURL: webhookURL}
	entries := model.Entries{&model.Entry{ID: 52, URL: "https://example.org/test.html", Title: "Test"}}
	userIntegrations := &model.Integration{UserID: 1, WebhookEnabled: true}

	PushEntries(feed, entries, userIntegrations)

	assertLogsDoNotContainWebhookURL(t, logs, "Unable to send new entries to Webhook")
}

// newFailingWebhookURL returns a webhook URL with secrets in its userinfo, path and query, served by a failing server.
func newFailingWebhookURL(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	return strings.Replace(server.URL, "://", "://user:secret-password@", 1) + "/api/webhook/secret-id?token=secret-token"
}

func assertLogsDoNotContainWebhookURL(t *testing.T, logs *bytes.Buffer, expectedMessage string) {
	t.Helper()

	if !strings.Contains(logs.String(), expectedMessage) {
		t.Fatalf("expected %q in logs; got: %s", expectedMessage, logs.String())
	}

	for _, secret := range []string{"secret-password", "secret-id", "secret-token"} {
		if strings.Contains(logs.String(), secret) {
			t.Errorf("expected logs to not contain %q; got: %s", secret, logs.String())
		}
	}
}

func captureDebugLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })
	return &logs
}

func configureIntegrationAllowPrivateNetworksOption(t *testing.T) {
	t.Helper()

	t.Setenv("INTEGRATION_ALLOW_PRIVATE_NETWORKS", "1")

	parsedOptions, err := config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unable to configure test options: %v", err)
	}

	previousOptions := config.Opts
	config.Opts = parsedOptions
	t.Cleanup(func() {
		config.Opts = previousOptions
	})
}

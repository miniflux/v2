// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

// configureIntegrationAllowPrivateNetworksOption sets the global config option
// required to allow the webhook HTTP client to reach the httptest server on
// localhost (a private address). It restores the previous config on test cleanup.
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

func testFeed() *model.Feed {
	return &model.Feed{
		ID:     1,
		UserID: 1,
		Category: &model.Category{
			ID:    1,
			Title: "Test",
		},
		FeedURL:   "https://example.org/feed.xml",
		SiteURL:   "https://example.org",
		Title:     "Test Feed",
		CheckedAt: time.Now(),
	}
}

func testEntries() model.Entries {
	return model.Entries{
		{ID: 10, UserID: 1, FeedID: 1, URL: "https://example.org/post-1", Title: "Post 1"},
	}
}

// TestSendNewEntriesWebhookEventType verifies that SendNewEntriesWebhookEvent
// sends a request whose JSON body has event_type = "new_entries".
func TestSendNewEntriesWebhookEventType(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	if err := client.SendNewEntriesWebhookEvent(testFeed(), testEntries()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("unable to unmarshal payload: %v", err)
	}

	if got := payload["event_type"]; got != NewEntriesEventType {
		t.Errorf("expected event_type %q, got %q", NewEntriesEventType, got)
	}
}

// TestSendUpdatedEntriesWebhookEventType verifies that SendUpdatedEntriesWebhookEvent
// sends a request whose JSON body has event_type = "updated_entries".
func TestSendUpdatedEntriesWebhookEventType(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	if err := client.SendUpdatedEntriesWebhookEvent(testFeed(), testEntries()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("unable to unmarshal payload: %v", err)
	}

	if got := payload["event_type"]; got != UpdatedEntriesEventType {
		t.Errorf("expected event_type %q, got %q", UpdatedEntriesEventType, got)
	}
}

// TestSendUpdatedEntriesWebhookEventTypeHeader verifies that the
// X-Miniflux-Event-Type header is set to "updated_entries".
func TestSendUpdatedEntriesWebhookEventTypeHeader(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Miniflux-Event-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	if err := client.SendUpdatedEntriesWebhookEvent(testFeed(), testEntries()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotHeader != UpdatedEntriesEventType {
		t.Errorf("expected X-Miniflux-Event-Type header %q, got %q", UpdatedEntriesEventType, gotHeader)
	}
}

// TestSendUpdatedEntriesWebhookNoEntriesIsNoop verifies that
// SendUpdatedEntriesWebhookEvent returns nil and makes no HTTP call when
// the entries slice is empty.
func TestSendUpdatedEntriesWebhookNoEntriesIsNoop(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	if err := client.SendUpdatedEntriesWebhookEvent(testFeed(), model.Entries{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if called {
		t.Error("expected no HTTP call for empty entries, but server was called")
	}
}

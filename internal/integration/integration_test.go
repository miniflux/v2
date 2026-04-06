// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

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

func testFeedAndEntries() (*model.Feed, model.Entries) {
	feed := &model.Feed{
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
	entries := model.Entries{
		{ID: 10, UserID: 1, FeedID: 1, URL: "https://example.org/post-1", Title: "Post 1"},
	}
	return feed, entries
}

// TestPushUpdatedEntriesLogsWebhookAttempt verifies that PushUpdatedEntries
// attempts to call the webhook integration when it is enabled.
// The webhook call will fail (empty URL) and produce a Warn log that we capture.
func TestPushUpdatedEntriesLogsWebhookAttempt(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	prev := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(prev)

	feed, entries := testFeedAndEntries()
	userIntegrations := &model.Integration{
		UserID:         1,
		WebhookEnabled: true,
		WebhookURL:     "", // empty → HTTP call fails → Warn log fires
	}

	PushUpdatedEntries(feed, entries, userIntegrations)

	out := buf.String()
	if !strings.Contains(out, "updated") {
		t.Fatalf("expected webhook warn log for updated entries; got: %s", out)
	}
}

// TestPushUpdatedEntriesSkipsNotificationIntegrations verifies that
// PushUpdatedEntries does not invoke notification channels (Telegram, Ntfy,
// Pushover, etc.) — only the webhook integration is triggered.
func TestPushUpdatedEntriesSkipsNotificationIntegrations(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	prev := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(prev)

	feed, entries := testFeedAndEntries()
	// All notification integrations enabled, webhook disabled.
	userIntegrations := &model.Integration{
		UserID:             1,
		WebhookEnabled:     false,
		TelegramBotEnabled: true,
		NtfyEnabled:        true,
		PushoverEnabled:    true,
		DiscordEnabled:     true,
		SlackEnabled:       true,
		MatrixBotEnabled:   true,
		AppriseEnabled:     true,
	}

	PushUpdatedEntries(feed, entries, userIntegrations)

	out := buf.String()
	if out != "" {
		t.Fatalf("expected no log output when webhook is disabled; got: %s", out)
	}
}

// TestPushUpdatedEntriesNoEntriesIsNoop verifies that PushUpdatedEntries
// exits immediately and calls no integrations when the entries slice is empty.
func TestPushUpdatedEntriesNoEntriesIsNoop(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	prev := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(prev)

	feed, _ := testFeedAndEntries()
	userIntegrations := &model.Integration{
		UserID:         1,
		WebhookEnabled: true,
		WebhookURL:     "https://example.org/hook",
	}

	PushUpdatedEntries(feed, model.Entries{}, userIntegrations)

	out := buf.String()
	if out != "" {
		t.Fatalf("expected no log output for empty entries; got: %s", out)
	}
}

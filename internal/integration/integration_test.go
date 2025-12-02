// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

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

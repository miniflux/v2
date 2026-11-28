// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"slices"
	"testing"
)

func TestJobListFeedURLsRedactsCredentials(t *testing.T) {
	jobs := JobList{
		{FeedURL: "https://example.org/feed.xml"},
		{FeedURL: "https://user:secret@example.org/?action=display&token=secret"},
	}

	expected := []string{
		"https://example.org/feed.xml",
		"https://user:xxxxx@example.org/?action=display&token=xxxxx",
	}
	if got := jobs.FeedURLs(); !slices.Equal(got, expected) {
		t.Errorf("Expected %q, got %q", expected, got)
	}
}

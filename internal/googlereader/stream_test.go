// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"slices"
	"testing"
)

func TestStreamTypeString(t *testing.T) {
	scenarios := map[StreamType]string{
		NoStream:       "NoStream",
		FeedStream:     "FeedStream",
		LikeStream:     "LikeStream",
		StreamType(42): "StreamType(42)",
		StreamType(-1): "StreamType(-1)",
	}

	for streamType, expected := range scenarios {
		if result := streamType.String(); result != expected {
			t.Errorf(`Unexpected result for stream type %d, got %q instead of %q`, int(streamType), result, expected)
		}
	}
}

func TestGetStream(t *testing.T) {
	const userID = int64(42)

	scenarios := []struct {
		streamID string
		expected Stream
	}{
		{"", Stream{NoStream, ""}},
		{"feed/123", Stream{FeedStream, "123"}},
		{"feed/https://example.org/feed.xml", Stream{FeedStream, "https://example.org/feed.xml"}},
		{"user/-/state/com.google/read", Stream{ReadStream, ""}},
		{"user/-/state/com.google/starred", Stream{StarredStream, ""}},
		{"user/-/state/com.google/reading-list", Stream{ReadingListStream, ""}},
		{"user/-/state/com.google/kept-unread", Stream{KeptUnreadStream, ""}},
		{"user/-/state/com.google/broadcast", Stream{BroadcastStream, ""}},
		{"user/-/state/com.google/broadcast-friends", Stream{BroadcastFriendsStream, ""}},
		{"user/-/state/com.google/like", Stream{LikeStream, ""}},
		{"user/42/state/com.google/read", Stream{ReadStream, ""}},
		{"user/42/state/com.google/reading-list", Stream{ReadingListStream, ""}},
		{"user/-/label/Tech", Stream{LabelStream, "Tech"}},
		{"user/42/label/Tech News", Stream{LabelStream, "Tech News"}},
		{"user/-/label/", Stream{LabelStream, ""}},
	}

	for _, scenario := range scenarios {
		result, err := getStream(scenario.streamID, userID)
		if err != nil {
			t.Errorf(`Unexpected error for stream %q: %v`, scenario.streamID, err)
			continue
		}
		if result != scenario.expected {
			t.Errorf(`Unexpected stream for %q, got %v instead of %v`, scenario.streamID, result, scenario.expected)
		}
	}
}

func TestGetStreamRejectsInvalidStreams(t *testing.T) {
	const userID = int64(42)

	for _, streamID := range []string{
		"invalid",
		"user/-/state/com.google/unknown",
		"user/42/state/com.google/unknown",
		// Another user's ID is not interchangeable with "-".
		"user/7/state/com.google/read",
		"user/7/label/Tech",
	} {
		if result, err := getStream(streamID, userID); err == nil {
			t.Errorf(`Expected an error for stream %q, got %v`, streamID, result)
		}
	}
}

func TestGetStreams(t *testing.T) {
	result, err := getStreams([]string{"feed/1", "user/-/label/Tech", "user/-/state/com.google/starred"}, 42)
	if err != nil {
		t.Fatalf(`Unexpected error: %v`, err)
	}

	expected := []Stream{{FeedStream, "1"}, {LabelStream, "Tech"}, {StarredStream, ""}}
	if !slices.Equal(result, expected) {
		t.Errorf(`Unexpected streams, got %v instead of %v`, result, expected)
	}

	result, err = getStreams(nil, 42)
	if err != nil {
		t.Fatalf(`Unexpected error: %v`, err)
	}
	if result == nil || len(result) != 0 {
		t.Errorf(`Expected an empty non-nil slice, got %#v`, result)
	}

	result, err = getStreams([]string{"feed/1", "invalid"}, 42)
	if err == nil {
		t.Fatal(`Expected an error when one of the streams is invalid`)
	}
	if len(result) != 0 {
		t.Errorf(`Expected no streams on error, got %v`, result)
	}
}

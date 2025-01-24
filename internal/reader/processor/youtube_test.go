// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"testing"
	"time"
)

func TestParseISO8601(t *testing.T) {
	var scenarios = []struct {
		duration string
		expected time.Duration
	}{
		// Live streams and radio.
		{"PT0M0S", 0},
		// https://www.youtube.com/watch?v=HLrqNhgdiC0
		{"PT6M20S", (6 * time.Minute) + (20 * time.Second)},
		// https://www.youtube.com/watch?v=LZa5KKfqHtA
		{"PT5M41S", (5 * time.Minute) + (41 * time.Second)},
		// https://www.youtube.com/watch?v=yIxEEgEuhT4
		{"PT51M52S", (51 * time.Minute) + (52 * time.Second)},
		// https://www.youtube.com/watch?v=bpHf1XcoiFs
		{"PT80M42S", (1 * time.Hour) + (20 * time.Minute) + (42 * time.Second)},
	}

	for _, tc := range scenarios {
		result, err := parseISO8601(tc.duration)
		if err != nil {
			t.Errorf("Got an error when parsing %q: %v", tc.duration, err)
		}

		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for duration %q`, result, tc.duration)
		}
	}
}

func TestGetYouTubeVideoIDFromURL(t *testing.T) {
	scenarios := []struct {
		url      string
		expected string
	}{
		{"https://www.youtube.com/watch?v=HLrqNhgdiC0", "HLrqNhgdiC0"},
		{"https://www.youtube.com/watch?v=HLrqNhgdiC0&feature=youtu.be", "HLrqNhgdiC0"},
		{"https://example.org/test", ""},
	}
	for _, tc := range scenarios {
		result := getVideoIDFromYouTubeURL(tc.url)
		if tc.expected != result {
			t.Errorf(`Unexpected result, got %q for url %q`, result, tc.url)
		}
	}
}

func TestIsYouTubeVideoURL(t *testing.T) {
	scenarios := []struct {
		url      string
		expected bool
	}{
		{"https://www.youtube.com/watch?v=HLrqNhgdiC0", true},
		{"https://www.youtube.com/watch?v=HLrqNhgdiC0&feature=youtu.be", true},
		{"https://example.org/test", false},
	}
	for _, tc := range scenarios {
		result := isYouTubeVideoURL(tc.url)
		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for url %q`, result, tc.url)
		}
	}
}

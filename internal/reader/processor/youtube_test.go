// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"testing"
)

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

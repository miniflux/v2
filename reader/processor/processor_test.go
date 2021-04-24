// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package processor // import "miniflux.app/reader/processor"

import (
	"testing"
	"time"

	"miniflux.app/model"
)

func TestBlockingEntries(t *testing.T) {
	var scenarios = []struct {
		feed     *model.Feed
		entry    *model.Entry
		expected bool
	}{
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Some Example"}, true},
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Something different but with matched content", Content: "Any Example"}, true},
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Something different"}, false},
		{&model.Feed{ID: 1}, &model.Entry{Title: "No rule defined"}, false},
	}

	for _, tc := range scenarios {
		result := isBlockedEntry(tc.feed, tc.entry)
		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for entry %q`, result, tc.entry.Title)
		}
	}
}

func TestAllowEntries(t *testing.T) {
	var scenarios = []struct {
		feed     *model.Feed
		entry    *model.Entry
		expected bool
	}{
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "Some Example"}, true},
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "Something different but with matched content", Content: "Any Example"}, true},
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "Something different"}, false},
		{&model.Feed{ID: 1}, &model.Entry{Title: "No rule defined"}, true},
	}

	for _, tc := range scenarios {
		result := isAllowedEntry(tc.feed, tc.entry)
		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for entry %q`, result, tc.entry.Title)
		}
	}
}

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

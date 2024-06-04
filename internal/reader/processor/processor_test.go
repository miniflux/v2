// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"testing"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

func TestBlockingEntries(t *testing.T) {
	var scenarios = []struct {
		feed     *model.Feed
		entry    *model.Entry
		user     *model.User
		expected bool
	}{
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{URL: "https://example.com"}, &model.User{}, true},
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{URL: "https://different.com"}, &model.User{}, false},
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Some Example"}, &model.User{}, true},
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Something different"}, &model.User{}, false},
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Something different", Tags: []string{"example", "something else"}}, &model.User{}, true},
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Example", Tags: []string{"example", "something else"}}, &model.User{}, true},
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Example", Tags: []string{"something different", "something else"}}, &model.User{}, true},
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Something different", Tags: []string{"something different", "something else"}}, &model.User{}, false},
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Something different", Author: "Example"}, &model.User{}, true},
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Something different", Author: "Something different"}, &model.User{}, false},
		{&model.Feed{ID: 1}, &model.Entry{Title: "No rule defined"}, &model.User{}, false},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://example.com", Title: "Some Example"}, &model.User{BlockFilterEntryRules: "URL((?i)example)~Title((?i)Test)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://different.com", Title: "Some Test"}, &model.User{BlockFilterEntryRules: "URL((?i)example)~Title((?i)Test)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://different.com", Title: "Some Example"}, &model.User{BlockFilterEntryRules: "URL((?i)example)~Title((?i)Test)"}, false},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://example.com", Content: "Some Example"}, &model.User{BlockFilterEntryRules: "CommentsURL((?i)example)~Content((?i)Test)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Test"}, &model.User{BlockFilterEntryRules: "CommentsURL((?i)example)~Content((?i)Test)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Example"}, &model.User{BlockFilterEntryRules: "CommentsURL((?i)example)~Content((?i)Test)"}, false},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Example", Tags: []string{"example", "something else"}}, &model.User{BlockFilterEntryRules: "Author((?i)example)~Tags((?i)Test)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{BlockFilterEntryRules: "Author((?i)example)~Tags((?i)example)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{BlockFilterEntryRules: "Author((?i)example)~Tags((?i)Test)"}, false},
	}

	for _, tc := range scenarios {
		result := isBlockedEntry(tc.feed, tc.entry, tc.user)
		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for entry %q`, result, tc.entry.Title)
		}
	}
}

func TestAllowEntries(t *testing.T) {
	var scenarios = []struct {
		feed     *model.Feed
		entry    *model.Entry
		user     *model.User
		expected bool
	}{
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "https://example.com"}, &model.User{}, true},
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "https://different.com"}, &model.User{}, false},
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "Some Example"}, &model.User{}, true},
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "Something different"}, &model.User{}, false},
		{&model.Feed{ID: 1}, &model.Entry{Title: "No rule defined"}, &model.User{}, true},
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "Something different", Tags: []string{"example", "something else"}}, &model.User{}, true},
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "Example", Tags: []string{"example", "something else"}}, &model.User{}, true},
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "Example", Tags: []string{"something different", "something else"}}, &model.User{}, true},
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "Something more", Tags: []string{"something different", "something else"}}, &model.User{}, false},
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "Something different", Author: "Example"}, &model.User{}, true},
		{&model.Feed{ID: 1, KeeplistRules: "(?i)example"}, &model.Entry{Title: "Something different", Author: "Something different"}, &model.User{}, false},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://example.com", Title: "Some Example"}, &model.User{KeepFilterEntryRules: "URL((?i)example)~Title((?i)Test)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://different.com", Title: "Some Test"}, &model.User{KeepFilterEntryRules: "URL((?i)example)~Title((?i)Test)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://different.com", Title: "Some Example"}, &model.User{KeepFilterEntryRules: "URL((?i)example)~Title((?i)Test)"}, false},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://example.com", Content: "Some Example"}, &model.User{KeepFilterEntryRules: "CommentsURL((?i)example)~Content((?i)Test)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Test"}, &model.User{KeepFilterEntryRules: "CommentsURL((?i)example)~Content((?i)Test)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Example"}, &model.User{KeepFilterEntryRules: "CommentsURL((?i)example)~Content((?i)Test)"}, false},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Example", Tags: []string{"example", "something else"}}, &model.User{KeepFilterEntryRules: "Author((?i)example)~Tags((?i)Test)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{KeepFilterEntryRules: "Author((?i)example)~Tags((?i)example)"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{KeepFilterEntryRules: "Author((?i)example)~Tags((?i)Test)"}, false},
	}

	for _, tc := range scenarios {
		result := isAllowedEntry(tc.feed, tc.entry, tc.user)
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

func TestIsRecentEntry(t *testing.T) {
	parser := config.NewParser()
	var err error
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}
	var scenarios = []struct {
		entry    *model.Entry
		expected bool
	}{
		{&model.Entry{Title: "Example1", Date: time.Date(2005, 5, 1, 05, 05, 05, 05, time.UTC)}, true},
		{&model.Entry{Title: "Example2", Date: time.Date(2010, 5, 1, 05, 05, 05, 05, time.UTC)}, true},
		{&model.Entry{Title: "Example3", Date: time.Date(2020, 5, 1, 05, 05, 05, 05, time.UTC)}, true},
		{&model.Entry{Title: "Example4", Date: time.Date(2024, 3, 15, 05, 05, 05, 05, time.UTC)}, true},
	}
	for _, tc := range scenarios {
		result := isRecentEntry(tc.entry)
		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for entry %q`, result, tc.entry.Title)
		}
	}
}

func TestMinifyEntryContent(t *testing.T) {
	input := `<p>    Some text with a <a href="http://example.org/"> link   </a>    </p>`
	expected := `<p>Some text with a <a href="http://example.org/">link</a></p>`
	result := minifyEntryContent(input)
	if expected != result {
		t.Errorf(`Unexpected result, got %q`, result)
	}
}

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
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://example.com", Title: "Some Example"}, &model.User{BlockFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://different.com", Title: "Some Test"}, &model.User{BlockFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://different.com", Title: "Some Example"}, &model.User{BlockFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, false},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://example.com", Content: "Some Example"}, &model.User{BlockFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Test"}, &model.User{BlockFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Example"}, &model.User{BlockFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, false},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Example", Tags: []string{"example", "something else"}}, &model.User{BlockFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)Test"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{BlockFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)example"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{BlockFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)Test"}, false},
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
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://example.com", Title: "Some Example"}, &model.User{KeepFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://different.com", Title: "Some Test"}, &model.User{KeepFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{URL: "https://different.com", Title: "Some Example"}, &model.User{KeepFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, false},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://example.com", Content: "Some Example"}, &model.User{KeepFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Test"}, &model.User{KeepFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Example"}, &model.User{KeepFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, false},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Example", Tags: []string{"example", "something else"}}, &model.User{KeepFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)Test"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{KeepFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)example"}, true},
		{&model.Feed{ID: 1, BlocklistRules: ""}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{KeepFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)Test"}, false},
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=between:2024-01-01,2024-12-31"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=between:2024-01-01,2024-12-31"}, false},
	}

	for _, tc := range scenarios {
		result := isAllowedEntry(tc.feed, tc.entry, tc.user)
		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for entry %q`, result, tc.entry.Title)
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

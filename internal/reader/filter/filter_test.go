// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package filter // import "miniflux.app/v2/internal/reader/filter"

import (
	"testing"
	"time"

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
		{&model.Feed{ID: 1, BlocklistRules: "[a-z"}, &model.Entry{URL: "https://example.com"}, &model.User{}, false}, // invalid regex
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
		{&model.Feed{ID: 1}, &model.Entry{URL: "https://example.com", Title: "Some Example"}, &model.User{BlockFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{URL: "https://different.com", Title: "Some Test"}, &model.User{BlockFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{URL: "https://different.com", Title: "Some Example"}, &model.User{BlockFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, false},
		{&model.Feed{ID: 1}, &model.Entry{CommentsURL: "https://example.com", Content: "Some Example"}, &model.User{BlockFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Test"}, &model.User{BlockFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Example"}, &model.User{BlockFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, false},
		{&model.Feed{ID: 1}, &model.Entry{Author: "Example", Tags: []string{"example", "something else"}}, &model.User{BlockFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{BlockFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)example"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{BlockFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)Test"}, false},
		{&model.Feed{ID: 1}, &model.Entry{Author: "Different", Tags: []string{"example", "test"}}, &model.User{BlockFilterEntryRules: "EntryAuthor\nEntryTag=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC)}, &model.User{BlockFilterEntryRules: "EntryDate=before:2024-03-15"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC)}, &model.User{BlockFilterEntryRules: "UnknownRuleType=test"}, false},
		{&model.Feed{ID: 1, BlockFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, &model.Entry{URL: "https://example.com", Title: "Some Example"}, &model.User{}, true},
		// Test cases for merged user and feed BlockFilterEntryRules
		{&model.Feed{ID: 1, BlockFilterEntryRules: "EntryURL=(?i)website"}, &model.Entry{URL: "https://example.com", Title: "Some Title"}, &model.User{BlockFilterEntryRules: "   EntryTitle=(?i)title   "}, true}, // User rule matches
		{&model.Feed{ID: 1, BlockFilterEntryRules: "EntryURL=(?i)example"}, &model.Entry{URL: "https://example.com", Title: "Some Other"}, &model.User{BlockFilterEntryRules: "EntryTitle=(?i)title"}, true},       // Feed rule matches
		{&model.Feed{ID: 1, BlockFilterEntryRules: "EntryURL=(?i)example"}, &model.Entry{URL: "https://different.com", Title: "Some Other"}, &model.User{BlockFilterEntryRules: "EntryTitle=(?i)title"}, false},    // Neither rule matches
		{&model.Feed{ID: 1, BlockFilterEntryRules: "EntryURL=(?i)example"}, &model.Entry{URL: "https://example.com", Title: "Some Title"}, &model.User{BlockFilterEntryRules: "EntryTitle=(?i)title"}, true},       // Both rules would match
	}

	for _, tc := range scenarios {
		result := IsBlockedEntry(tc.feed, tc.entry, tc.user)
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
		{&model.Feed{ID: 1, KeeplistRules: "[a-z"}, &model.Entry{Title: "https://example.com"}, &model.User{}, false}, // invalid regex
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
		{&model.Feed{ID: 1}, &model.Entry{URL: "https://example.com", Title: "Some Example"}, &model.User{KeepFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{URL: "https://different.com", Title: "Some Test"}, &model.User{KeepFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{URL: "https://different.com", Title: "Some Example"}, &model.User{KeepFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, false},
		{&model.Feed{ID: 1}, &model.Entry{CommentsURL: "https://example.com", Content: "Some Example"}, &model.User{KeepFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Test"}, &model.User{KeepFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{CommentsURL: "https://different.com", Content: "Some Example"}, &model.User{KeepFilterEntryRules: "EntryCommentsURL=(?i)example\nEntryContent=(?i)Test"}, false},
		{&model.Feed{ID: 1}, &model.Entry{Author: "Example", Tags: []string{"example", "something else"}}, &model.User{KeepFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{KeepFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)example"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Author: "Different", Tags: []string{"example", "something else"}}, &model.User{KeepFilterEntryRules: "EntryAuthor=(?i)example\nEntryTag=(?i)Test"}, false},
		{&model.Feed{ID: 1}, &model.Entry{Author: "Different", Tags: []string{"example", "some test"}}, &model.User{KeepFilterEntryRules: "EntryAuthor\nEntryTag=(?i)Test"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Now().Add(24 * time.Hour)}, &model.User{KeepFilterEntryRules: "EntryDate=future"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Now().Add(-24 * time.Hour)}, &model.User{KeepFilterEntryRules: "EntryDate=future"}, false},
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=before:2024-03-15"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=before:invalid-date"}, false}, // invalid date format
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 3, 16, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=after:2024-03-15"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 3, 16, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=after:invalid-date"}, false}, // invalid date format
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=between:2024-03-01,2024-03-15"}, true},
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=between:2024-03-01,2024-03-15"}, false},
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=between:invalid-date,2024-03-15"}, false}, // invalid date format
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=between:2024-03-15,invalid-date"}, false}, // invalid date format
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=between:2024-03-15"}, false},              // missing second date in range
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=abcd"}, false},                            // no colon in rule value
		{&model.Feed{ID: 1}, &model.Entry{Date: time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)}, &model.User{KeepFilterEntryRules: "EntryDate=unknown:2024-03-15"}, false},              // unknown rule type
		{&model.Feed{ID: 1, KeepFilterEntryRules: "EntryURL=(?i)example\nEntryTitle=(?i)Test"}, &model.Entry{URL: "https://example.com", Title: "Some Example"}, &model.User{}, true},
		// Test cases for merged user and feed KeepFilterEntryRules
		{&model.Feed{ID: 1, KeepFilterEntryRules: "EntryURL=(?i)website"}, &model.Entry{URL: "https://example.com", Title: "Some Title"}, &model.User{KeepFilterEntryRules: "EntryTitle=(?i)title"}, true},    // User rule matches
		{&model.Feed{ID: 1, KeepFilterEntryRules: "EntryURL=(?i)example"}, &model.Entry{URL: "https://example.com", Title: "Some Other"}, &model.User{KeepFilterEntryRules: "EntryTitle=(?i)title"}, true},    // Feed rule matches
		{&model.Feed{ID: 1, KeepFilterEntryRules: "EntryURL=(?i)example"}, &model.Entry{URL: "https://different.com", Title: "Some Other"}, &model.User{KeepFilterEntryRules: "EntryTitle=(?i)title"}, false}, // Neither rule matches
		{&model.Feed{ID: 1, KeepFilterEntryRules: "EntryURL=(?i)example"}, &model.Entry{URL: "https://example.com", Title: "Some Title"}, &model.User{KeepFilterEntryRules: "EntryTitle=(?i)title"}, true},    // Both rules would match
	}

	for _, tc := range scenarios {
		result := IsAllowedEntry(tc.feed, tc.entry, tc.user)
		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for entry %q`, result, tc.entry.Title)
		}
	}
}

// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package processor // import "miniflux.app/reader/processor"

import (
	"testing"

	"miniflux.app/model"
)

func TestBlockingEntries(t *testing.T) {
	var scenarios = []struct {
		feed     *model.Feed
		entry    *model.Entry
		expected bool
	}{
		{&model.Feed{ID: 1, BlocklistRules: "(?i)example"}, &model.Entry{Title: "Some Example"}, true},
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

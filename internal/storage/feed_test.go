// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"sort"
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestByNameSorting(t *testing.T) {
	feeds := model.Feeds{
		{ID: 1, Title: "zeta", UnreadCount: 42, ParsingErrorCount: 3},
		{ID: 2, Title: "Beta", UnreadCount: 0},
		{ID: 3, Title: "alpha", UnreadCount: 7},
	}

	sort.Sort(byName{feeds})

	titles := make([]string, len(feeds))
	for i, feed := range feeds {
		titles[i] = feed.Title
	}

	expected := []string{"alpha", "Beta", "zeta"}
	for i := range expected {
		if titles[i] != expected[i] {
			t.Fatalf("unexpected order: got %v, want %v", titles, expected)
		}
	}
}

func TestByNameSortingKeepsDisabledFeedsLast(t *testing.T) {
	feeds := model.Feeds{
		{ID: 1, Title: "a-disabled", Disabled: true},
		{ID: 2, Title: "zeta"},
		{ID: 3, Title: "beta"},
	}

	sort.Sort(byName{feeds})

	expected := []string{"beta", "zeta", "a-disabled"}
	for i := range expected {
		if feeds[i].Title != expected[i] {
			t.Fatalf("unexpected order: got %v, want %v", feedTitles(feeds), expected)
		}
	}
}

func TestByNameSortingIsDeterministicForEqualTitles(t *testing.T) {
	feeds := model.Feeds{
		{ID: 1, Title: "a"},
		{ID: 2, Title: "A"},
	}

	sort.Sort(byName{feeds})

	expected := []string{"A", "a"}
	for i := range expected {
		if feeds[i].Title != expected[i] {
			t.Fatalf("unexpected order: got %v, want %v", feedTitles(feeds), expected)
		}
	}
}

func feedTitles(feeds model.Feeds) []string {
	titles := make([]string, len(feeds))
	for i, feed := range feeds {
		titles[i] = feed.Title
	}
	return titles
}

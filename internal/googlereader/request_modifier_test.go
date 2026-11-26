// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"miniflux.app/v2/internal/http/request"
)

func newStreamFilterRequest(query string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/reader/api/0/stream/items/ids?"+query, nil)
	return r.WithContext(context.WithValue(r.Context(), request.UserIDContextKey, int64(42)))
}

func TestParseStreamFilterFromRequest(t *testing.T) {
	scenarios := []struct {
		name     string
		query    string
		expected requestModifiers
	}{
		{
			name:     "defaults",
			query:    "",
			expected: requestModifiers{UserID: 42, SortDirection: "desc", Streams: []Stream{}, ExcludeTargets: []Stream{}, FilterTargets: []Stream{}},
		},
		{
			name:     "oldest first",
			query:    "r=o",
			expected: requestModifiers{UserID: 42, SortDirection: "asc", Streams: []Stream{}, ExcludeTargets: []Stream{}, FilterTargets: []Stream{}},
		},
		{
			name:     "unknown sort order falls back to newest first",
			query:    "r=x",
			expected: requestModifiers{UserID: 42, SortDirection: "desc", Streams: []Stream{}, ExcludeTargets: []Stream{}, FilterTargets: []Stream{}},
		},
		{
			name:  "streams and targets",
			query: "s=user/-/state/com.google/reading-list&xt=user/-/state/com.google/read&xt=user/42/state/com.google/starred&it=user/-/label/Tech",
			expected: requestModifiers{
				UserID:         42,
				SortDirection:  "desc",
				Streams:        []Stream{{ReadingListStream, ""}},
				ExcludeTargets: []Stream{{ReadStream, ""}, {StarredStream, ""}},
				FilterTargets:  []Stream{{LabelStream, "Tech"}},
			},
		},
		{
			name:  "blank stream values are ignored",
			query: "s=+&s=feed/12&xt=",
			expected: requestModifiers{
				UserID:         42,
				SortDirection:  "desc",
				Streams:        []Stream{{FeedStream, "12"}},
				ExcludeTargets: []Stream{},
				FilterTargets:  []Stream{},
			},
		},
		{
			name:  "numeric parameters",
			query: "n=50&c=100&ot=1700000000&nt=1800000000",
			expected: requestModifiers{
				UserID:         42,
				SortDirection:  "desc",
				Count:          50,
				Offset:         100,
				StartTime:      1700000000,
				StopTime:       1800000000,
				Streams:        []Stream{},
				ExcludeTargets: []Stream{},
				FilterTargets:  []Stream{},
			},
		},
		{
			// The 10000 cap is applied by the entry query builder, not here.
			name:     "count above the maximum is kept as is",
			query:    "n=20000",
			expected: requestModifiers{UserID: 42, SortDirection: "desc", Count: 20000, Streams: []Stream{}, ExcludeTargets: []Stream{}, FilterTargets: []Stream{}},
		},
		{
			name:     "invalid and negative numbers fall back to zero",
			query:    "n=abc&c=-1&ot=yesterday&nt=-5",
			expected: requestModifiers{UserID: 42, SortDirection: "desc", Streams: []Stream{}, ExcludeTargets: []Stream{}, FilterTargets: []Stream{}},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result, err := parseStreamFilterFromRequest(newStreamFilterRequest(scenario.query))
			if err != nil {
				t.Fatalf(`Unexpected error: %v`, err)
			}
			if !reflect.DeepEqual(result, scenario.expected) {
				t.Errorf("Unexpected modifiers\ngot:  %#v\nwant: %#v", result, scenario.expected)
			}
		})
	}
}

func TestParseStreamFilterFromRequestRejectsInvalidStreams(t *testing.T) {
	for _, query := range []string{
		"s=invalid",
		"xt=invalid",
		"it=invalid",
		"s=user/-/state/com.google/unknown",
	} {
		if result, err := parseStreamFilterFromRequest(newStreamFilterRequest(query)); err == nil {
			t.Errorf(`Expected an error for query %q, got %v`, query, result)
		}
	}
}

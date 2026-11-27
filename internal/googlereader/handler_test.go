// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCheckAndSimplifyTags(t *testing.T) {
	read := Stream{Type: ReadStream}
	keptUnread := Stream{Type: KeptUnreadStream}
	starred := Stream{Type: StarredStream}
	like := Stream{Type: LikeStream}
	broadcast := Stream{Type: BroadcastStream}

	scenarios := []struct {
		name     string
		add      []Stream
		remove   []Stream
		expected map[StreamType]bool
	}{
		{"add read", []Stream{read}, nil, map[StreamType]bool{ReadStream: true}},
		{"remove read", nil, []Stream{read}, map[StreamType]bool{ReadStream: false}},
		{"add kept-unread", []Stream{keptUnread}, nil, map[StreamType]bool{ReadStream: false}},
		{"remove kept-unread", nil, []Stream{keptUnread}, map[StreamType]bool{ReadStream: true}},
		{"add starred", []Stream{starred}, nil, map[StreamType]bool{StarredStream: true}},
		{"remove starred", nil, []Stream{starred}, map[StreamType]bool{StarredStream: false}},
		{"add read twice", []Stream{read, read}, nil, map[StreamType]bool{ReadStream: true}},
		{"add kept-unread twice", []Stream{keptUnread, keptUnread}, nil, map[StreamType]bool{ReadStream: false}},
		{"remove read twice", nil, []Stream{read, read}, map[StreamType]bool{ReadStream: false}},
		{"remove starred twice", nil, []Stream{starred, starred}, map[StreamType]bool{StarredStream: false}},
		{"add read and remove kept-unread", []Stream{read}, []Stream{keptUnread}, map[StreamType]bool{ReadStream: true}},
		{"add kept-unread and remove read", []Stream{keptUnread}, []Stream{read}, map[StreamType]bool{ReadStream: false}},
		{"add read and starred", []Stream{read, starred}, nil, map[StreamType]bool{ReadStream: true, StarredStream: true}},
		{"add read and remove starred", []Stream{read}, []Stream{starred}, map[StreamType]bool{ReadStream: true, StarredStream: false}},
		{"like and broadcast are ignored", []Stream{like, broadcast}, []Stream{like, broadcast}, map[StreamType]bool{}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result, err := checkAndSimplifyTags(scenario.add, scenario.remove)
			if err != nil {
				t.Fatalf(`Unexpected error: %v`, err)
			}
			if !maps.Equal(result, scenario.expected) {
				t.Errorf(`Unexpected tags, got %v instead of %v`, result, scenario.expected)
			}
		})
	}
}

func TestCheckAndSimplifyTagsRejectsConflicts(t *testing.T) {
	read := Stream{Type: ReadStream}
	keptUnread := Stream{Type: KeptUnreadStream}
	starred := Stream{Type: StarredStream}

	scenarios := []struct {
		name   string
		add    []Stream
		remove []Stream
	}{
		{"add and remove read", []Stream{read}, []Stream{read}},
		{"add read then kept-unread", []Stream{read, keptUnread}, nil},
		{"add kept-unread then read", []Stream{keptUnread, read}, nil},
		{"add and remove kept-unread", []Stream{keptUnread}, []Stream{keptUnread}},
		{"remove read and kept-unread", nil, []Stream{read, keptUnread}},
		{"remove kept-unread and read", nil, []Stream{keptUnread, read}},
		{"add and remove starred", []Stream{starred}, []Stream{starred}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			if result, err := checkAndSimplifyTags(scenario.add, scenario.remove); err == nil {
				t.Errorf(`Expected an error, got %v`, result)
			}
		})
	}
}

func TestCheckAndSimplifyTagsRejectsUnsupportedTags(t *testing.T) {
	for _, stream := range []Stream{
		{Type: ReadingListStream},
		{Type: BroadcastFriendsStream},
		{Type: LabelStream, ID: "Tech"},
		{Type: FeedStream, ID: "1"},
		{Type: NoStream},
	} {
		if _, err := checkAndSimplifyTags([]Stream{stream}, nil); err == nil {
			t.Errorf(`Expected an error when adding %v`, stream)
		}
		if _, err := checkAndSimplifyTags(nil, []Stream{stream}); err == nil {
			t.Errorf(`Expected an error when removing %v`, stream)
		}
	}
}

func TestCheckAndSimplifyTagsReturnsSimultaneousError(t *testing.T) {
	_, err := checkAndSimplifyTags([]Stream{{Type: ReadStream}, {Type: KeptUnreadStream}}, nil)
	if !errors.Is(err, errSimultaneously) {
		t.Errorf(`Expected errSimultaneously, got %v`, err)
	}
}

func TestCheckOutputFormat(t *testing.T) {
	newPostRequest := func(target string, form url.Values) *http.Request {
		r := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return r
	}

	scenarios := []struct {
		name    string
		request *http.Request
		valid   bool
	}{
		{"GET with output=json", httptest.NewRequest(http.MethodGet, "/?output=json", nil), true},
		{"GET without output", httptest.NewRequest(http.MethodGet, "/", nil), false},
		{"GET with output=xml", httptest.NewRequest(http.MethodGet, "/?output=xml", nil), false},
		{"POST with output=json in the body", newPostRequest("/", url.Values{"output": {"json"}}), true},
		{"POST with output=json in the query string", newPostRequest("/?output=json", url.Values{}), true},
		{"POST without output", newPostRequest("/", url.Values{"T": {"token"}}), false},
		{"POST with output=atom", newPostRequest("/", url.Values{"output": {"atom"}}), false},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := checkOutputFormat(scenario.request)
			if scenario.valid && err != nil {
				t.Errorf(`Unexpected error: %v`, err)
			}
			if !scenario.valid && err == nil {
				t.Error(`Expected an error`)
			}
		})
	}
}

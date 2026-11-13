// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rssfeed // import "miniflux.app/v2/internal/rssfeed"

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
)

const testAPIKey = "secret-token"

type fakeAuthenticator struct {
	user           *model.User
	userErr        error
	lastLoginIDs   []int64
	updatedAPIKeys []string
}

func newFakeAuthenticator() *fakeAuthenticator {
	return &fakeAuthenticator{
		user: &model.User{
			ID:       42,
			Username: "jane",
			Timezone: "UTC",
		},
	}
}

func (f *fakeAuthenticator) UserByAPIKey(apiKey string) (*model.User, error) {
	if f.userErr != nil {
		return nil, f.userErr
	}
	if apiKey != testAPIKey {
		return nil, nil
	}
	return f.user, nil
}

func (f *fakeAuthenticator) SetLastLogin(userID int64) error {
	f.lastLoginIDs = append(f.lastLoginIDs, userID)
	return nil
}

func (f *fakeAuthenticator) SetAPIKeyUsedTimestamp(userID int64, apiKey string) error {
	f.updatedAPIKeys = append(f.updatedAPIKeys, apiKey)
	return nil
}

type fakeFeedStore struct {
	entries     model.Entries
	entriesErr  error
	feeds       map[int64]*model.Feed
	categories  map[int64]*model.Category
	lookupErr   error
	filters     []entryFilter
	queriedUser []int64
	marked      [][]int64
}

func newFakeFeedStore() *fakeFeedStore {
	return &fakeFeedStore{
		feeds: map[int64]*model.Feed{
			1: {ID: 1, Title: "Example Feed", SiteURL: "https://example.org/"},
		},
		categories: map[int64]*model.Category{
			3: {ID: 3, Title: "Tech"},
		},
	}
}

func (f *fakeFeedStore) Entries(userID int64, filter entryFilter) (model.Entries, error) {
	f.queriedUser = append(f.queriedUser, userID)
	f.filters = append(f.filters, filter)
	if f.entriesErr != nil {
		return nil, f.entriesErr
	}
	return f.entries, nil
}

func (f *fakeFeedStore) MarkAsRead(userID int64, entryIDs []int64) error {
	f.marked = append(f.marked, entryIDs)
	return nil
}

func (f *fakeFeedStore) Feed(userID, feedID int64) (*model.Feed, error) {
	if f.lookupErr != nil {
		return nil, f.lookupErr
	}
	return f.feeds[feedID], nil
}

func (f *fakeFeedStore) Category(userID, categoryID int64) (*model.Category, error) {
	if f.lookupErr != nil {
		return nil, f.lookupErr
	}
	return f.categories[categoryID], nil
}

func newTestHandler(store *fakeFeedStore, auth *fakeAuthenticator) http.Handler {
	return newHandler(store, newAuthMiddleware(auth))
}

func testEntries(ids ...int64) model.Entries {
	entries := make(model.Entries, 0, len(ids))
	for _, id := range ids {
		entries = append(entries, &model.Entry{
			ID:      id,
			Title:   "Entry title",
			URL:     "https://example.org/post",
			Content: "<p>Content</p>",
			Feed:    &model.Feed{ID: 1, Title: "Example Feed"},
		})
	}
	return entries
}

func get(t *testing.T, handler http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder
}

func TestUnreadFeedMarksReturnedEntries(t *testing.T) {
	store := newFakeFeedStore()
	store.entries = testEntries(10, 11, 12)

	recorder := get(t, newTestHandler(store, newFakeAuthenticator()), "/rss/unread?api_key="+testAPIKey)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/rss+xml; charset=utf-8" {
		t.Errorf("unexpected content type: %q", got)
	}
	if len(store.marked) != 1 {
		t.Fatalf("expected exactly one mark-as-read call, got %d", len(store.marked))
	}
	if len(store.marked[0]) != 3 || store.marked[0][0] != 10 || store.marked[0][2] != 12 {
		t.Errorf("expected returned entries to be marked as read, got %v", store.marked[0])
	}
	if len(store.queriedUser) != 1 || store.queriedUser[0] != 42 {
		t.Errorf("expected entries to be fetched for the authenticated user, got %v", store.queriedUser)
	}
	if len(store.filters) != 1 || store.filters[0].Kind != streamUnread || store.filters[0].Limit != defaultEntriesLimit {
		t.Errorf("unexpected entry filter: %+v", store.filters)
	}

	body := recorder.Body.String()
	for _, expected := range []string{"<rss version=\"2.0\"", "urn:miniflux:entry:10", "Entry title"} {
		if !strings.Contains(body, expected) {
			t.Errorf("response body does not contain %q:\n%s", expected, body)
		}
	}
}

func TestUnreadFeedMarkAsReadIsOptOut(t *testing.T) {
	for _, value := range []string{"0", "false"} {
		t.Run(value, func(t *testing.T) {
			store := newFakeFeedStore()
			store.entries = testEntries(10)

			target := "/rss/unread?api_key=" + testAPIKey + "&mark_read=" + value
			recorder := get(t, newTestHandler(store, newFakeAuthenticator()), target)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", recorder.Code)
			}
			if len(store.marked) != 0 {
				t.Errorf("entries must not be marked as read when mark_read=%s, got %v", value, store.marked)
			}
		})
	}
}

func TestUnreadFeedHeadRequestDoesNotMarkEntries(t *testing.T) {
	store := newFakeFeedStore()
	store.entries = testEntries(10)

	recorder := httptest.NewRecorder()
	newTestHandler(store, newFakeAuthenticator()).
		ServeHTTP(recorder, httptest.NewRequest(http.MethodHead, "/rss/unread?api_key="+testAPIKey, nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if len(store.marked) != 0 {
		t.Errorf("HEAD requests must not mark entries as read, got %v", store.marked)
	}
}

func TestUnreadFeedDoesNotMarkWhenNoEntriesAreReturned(t *testing.T) {
	store := newFakeFeedStore()
	store.entries = model.Entries{}

	recorder := get(t, newTestHandler(store, newFakeAuthenticator()), "/rss/unread?api_key="+testAPIKey)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if len(store.marked) != 0 {
		t.Errorf("empty responses must not issue a mark-as-read call, got %v", store.marked)
	}
}

func TestOnlyTheUnreadFeedMarksEntries(t *testing.T) {
	targets := []string{
		"/rss/starred",
		"/rss/read",
		"/rss/all",
		"/rss/feed/1",
		"/rss/category/3",
		"/rss/search?q=term",
	}

	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			store := newFakeFeedStore()
			store.entries = testEntries(10, 11)

			separator := "?"
			if strings.Contains(target, "?") {
				separator = "&"
			}

			recorder := get(t, newTestHandler(store, newFakeAuthenticator()), target+separator+"api_key="+testAPIKey)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
			}
			if len(store.marked) != 0 {
				t.Errorf("%s must not mark entries as read, got %v", target, store.marked)
			}
		})
	}
}

func TestFeedsUseExpectedEntryFilters(t *testing.T) {
	scenarios := []struct {
		target string
		want   entryFilter
	}{
		{"?api_key=" + testAPIKey, entryFilter{Kind: streamUnread, Limit: defaultEntriesLimit}},
		{"/starred?api_key=" + testAPIKey, entryFilter{Kind: streamStarred, Limit: defaultEntriesLimit}},
		{"/read?api_key=" + testAPIKey, entryFilter{Kind: streamRead, Limit: defaultEntriesLimit}},
		{"/all?api_key=" + testAPIKey, entryFilter{Kind: streamAll, Limit: defaultEntriesLimit}},
		{"/feed/1?api_key=" + testAPIKey, entryFilter{Kind: streamFeed, FeedID: 1, Limit: defaultEntriesLimit}},
		{"/category/3?api_key=" + testAPIKey, entryFilter{Kind: streamCategory, CategoryID: 3, Limit: defaultEntriesLimit}},
		{"/search?q=%20hello%20&api_key=" + testAPIKey, entryFilter{Kind: streamSearch, Search: "hello", Limit: defaultEntriesLimit}},
		{"?limit=10&api_key=" + testAPIKey, entryFilter{Kind: streamUnread, Limit: 10}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.target, func(t *testing.T) {
			store := newFakeFeedStore()

			target := "/rss/unread"
			if strings.HasPrefix(scenario.target, "/") {
				target = "/rss" + scenario.target
			} else {
				target += scenario.target
			}

			recorder := get(t, newTestHandler(store, newFakeAuthenticator()), target)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
			}
			if len(store.filters) != 1 {
				t.Fatalf("expected exactly one entry query, got %d", len(store.filters))
			}
			if store.filters[0] != scenario.want {
				t.Errorf("expected filter %+v, got %+v", scenario.want, store.filters[0])
			}
		})
	}
}

func TestLimitIsClamped(t *testing.T) {
	scenarios := []struct {
		value string
		want  int
	}{
		{"", defaultEntriesLimit},
		{"0", defaultEntriesLimit},
		{"-5", defaultEntriesLimit},
		{"abc", defaultEntriesLimit},
		{"10", 10},
		{"200", 200},
		{"500", 500},
		{"1000", maxEntriesLimit},
	}

	for _, scenario := range scenarios {
		t.Run("limit="+scenario.value, func(t *testing.T) {
			store := newFakeFeedStore()

			target := "/rss/unread?api_key=" + testAPIKey
			if scenario.value != "" {
				target += "&limit=" + scenario.value
			}

			recorder := get(t, newTestHandler(store, newFakeAuthenticator()), target)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", recorder.Code)
			}
			if len(store.filters) != 1 || store.filters[0].Limit != scenario.want {
				t.Errorf("expected limit %d, got %+v", scenario.want, store.filters)
			}
		})
	}
}

func TestSingleFeedUsesFeedMetadata(t *testing.T) {
	store := newFakeFeedStore()
	store.entries = testEntries(10)

	recorder := get(t, newTestHandler(store, newFakeAuthenticator()), "/rss/feed/1?api_key="+testAPIKey)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "<title>Example Feed</title>") {
		t.Errorf("expected the feed title to be used as the document title:\n%s", body)
	}
	if !strings.Contains(body, `<link>https://example.org/</link>`) {
		t.Errorf("expected the feed site URL to be used as the document link:\n%s", body)
	}
}

func TestUnknownFeedOrCategoryReturnsNotFound(t *testing.T) {
	for _, target := range []string{"/rss/feed/999", "/rss/category/999"} {
		t.Run(target, func(t *testing.T) {
			store := newFakeFeedStore()

			recorder := get(t, newTestHandler(store, newFakeAuthenticator()), target+"?api_key="+testAPIKey)
			if recorder.Code != http.StatusNotFound {
				t.Errorf("expected 404, got %d", recorder.Code)
			}
			if len(store.filters) != 0 {
				t.Errorf("entries must not be queried for a missing feed or category, got %+v", store.filters)
			}
		})
	}
}

func TestStorageFailuresReturnInternalServerError(t *testing.T) {
	t.Run("entry query", func(t *testing.T) {
		store := newFakeFeedStore()
		store.entriesErr = errors.New("boom")

		recorder := get(t, newTestHandler(store, newFakeAuthenticator()), "/rss/all?api_key="+testAPIKey)
		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", recorder.Code)
		}
	})

	t.Run("feed lookup", func(t *testing.T) {
		store := newFakeFeedStore()
		store.lookupErr = errors.New("boom")

		recorder := get(t, newTestHandler(store, newFakeAuthenticator()), "/rss/feed/1?api_key="+testAPIKey)
		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", recorder.Code)
		}
	})

	t.Run("category lookup", func(t *testing.T) {
		store := newFakeFeedStore()
		store.lookupErr = errors.New("boom")

		recorder := get(t, newTestHandler(store, newFakeAuthenticator()), "/rss/category/3?api_key="+testAPIKey)
		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", recorder.Code)
		}
	})
}

func TestSearchRequiresQueryParameter(t *testing.T) {
	store := newFakeFeedStore()

	recorder := get(t, newTestHandler(store, newFakeAuthenticator()), "/rss/search?api_key="+testAPIKey)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
	if len(store.filters) != 0 {
		t.Errorf("entries must not be queried without a search term, got %+v", store.filters)
	}
}

func TestAuthentication(t *testing.T) {
	scenarios := []struct {
		name       string
		prepare    func(*http.Request)
		wantStatus int
	}{
		{
			name:       "api key query parameter",
			prepare:    func(r *http.Request) {},
			wantStatus: http.StatusOK,
		},
		{
			name: "http basic authentication",
			prepare: func(r *http.Request) {
				r.SetBasicAuth("jane", testAPIKey)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "bearer token",
			prepare: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer "+testAPIKey)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "username mismatch",
			prepare: func(r *http.Request) {
				r.SetBasicAuth("john", testAPIKey)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid api key",
			prepare: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer wrong-key")
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing credentials",
			prepare: func(r *http.Request) {
				r.Header.Del("Authorization")
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			store := newFakeFeedStore()
			auth := newFakeAuthenticator()

			target := "/rss/unread?api_key=" + testAPIKey
			switch scenario.name {
			case "username mismatch", "invalid api key", "missing credentials":
				target = "/rss/unread"
			}

			request := httptest.NewRequest(http.MethodGet, target, nil)
			scenario.prepare(request)

			recorder := httptest.NewRecorder()
			newTestHandler(store, auth).ServeHTTP(recorder, request)

			if recorder.Code != scenario.wantStatus {
				t.Fatalf("expected %d, got %d: %s", scenario.wantStatus, recorder.Code, recorder.Body.String())
			}

			if scenario.wantStatus == http.StatusOK {
				if len(auth.lastLoginIDs) != 1 || auth.lastLoginIDs[0] != 42 {
					t.Errorf("expected the last login to be recorded, got %v", auth.lastLoginIDs)
				}
				if len(auth.updatedAPIKeys) != 1 || auth.updatedAPIKeys[0] != testAPIKey {
					t.Errorf("expected the API key usage to be recorded, got %v", auth.updatedAPIKeys)
				}
				return
			}

			if got := recorder.Header().Get("WWW-Authenticate"); !strings.HasPrefix(got, "Basic ") {
				t.Errorf("expected a Basic challenge, got %q", got)
			}
			if len(store.filters) != 0 {
				t.Errorf("rejected requests must not query entries, got %+v", store.filters)
			}
		})
	}
}

func TestAuthenticationBackendFailureReturnsUnauthorized(t *testing.T) {
	store := newFakeFeedStore()
	auth := newFakeAuthenticator()
	auth.userErr = errors.New("database is down")

	recorder := get(t, newTestHandler(store, auth), "/rss/unread?api_key="+testAPIKey)
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", recorder.Code)
	}
}

func TestAPIKeyIsNotLeakedInTheResponse(t *testing.T) {
	store := newFakeFeedStore()
	store.entries = testEntries(10)

	for _, target := range []string{
		"/rss/unread?api_key=" + testAPIKey + "&limit=10",
		"/atom/unread?api_key=" + testAPIKey + "&limit=10",
	} {
		t.Run(target, func(t *testing.T) {
			recorder := get(t, newTestHandler(store, newFakeAuthenticator()), target)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", recorder.Code)
			}
			if strings.Contains(recorder.Body.String(), testAPIKey) {
				t.Errorf("the API key must not be part of the generated document:\n%s", recorder.Body.String())
			}
		})
	}
}

func TestAtomFeedFormat(t *testing.T) {
	store := newFakeFeedStore()
	store.entries = testEntries(10)

	recorder := get(t, newTestHandler(store, newFakeAuthenticator()), "/atom/unread?mark_read=0&api_key="+testAPIKey)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/atom+xml; charset=utf-8" {
		t.Errorf("unexpected content type: %q", got)
	}

	body := recorder.Body.String()
	if !strings.HasPrefix(body, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Errorf("unexpected Atom prologue:\n%s", body)
	}
	if !strings.Contains(body, `<feed xmlns="http://www.w3.org/2005/Atom">`) {
		t.Errorf("expected an Atom feed:\n%s", body)
	}
}

func TestAtomFeedIDIsStableAcrossQueryParameters(t *testing.T) {
	store := newFakeFeedStore()
	store.entries = testEntries(10)
	handler := newTestHandler(store, newFakeAuthenticator())

	feedID := func(target string) string {
		t.Helper()

		recorder := get(t, handler, target)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", target, recorder.Code)
		}

		body := recorder.Body.String()
		start := strings.Index(body, "<id>")
		if start < 0 {
			t.Fatalf("the Atom feed has no id:\n%s", body)
		}
		rest := body[start+len("<id>"):]
		end := strings.Index(rest, "</id>")
		if end < 0 {
			t.Fatalf("the Atom feed id is not closed:\n%s", body)
		}

		return rest[:end]
	}

	withoutParameters := feedID("/atom/unread?api_key=" + testAPIKey)
	withParameters := feedID("/atom/unread?api_key=" + testAPIKey + "&limit=5&mark_read=0")

	if withParameters != withoutParameters {
		t.Errorf("the Atom feed id must not change with query parameters: %q != %q", withParameters, withoutParameters)
	}
	if strings.Contains(withParameters, "?") {
		t.Errorf("the Atom feed id must not contain a query string: %q", withParameters)
	}
}

func TestHelpPage(t *testing.T) {
	for _, prefix := range []string{"/rss/", "/atom/"} {
		t.Run(prefix, func(t *testing.T) {
			recorder := get(t, newTestHandler(newFakeFeedStore(), newFakeAuthenticator()), prefix+"?api_key="+testAPIKey)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", recorder.Code)
			}

			body := recorder.Body.String()
			for _, expected := range []string{"/unread", "/starred", "/feed/{feedID}", "limit=", "mark_read=0"} {
				if !strings.Contains(body, expected) {
					t.Errorf("help page does not mention %q:\n%s", expected, body)
				}
			}
		})
	}
}

func TestUnsupportedRoutesAndMethods(t *testing.T) {
	store := newFakeFeedStore()
	handler := newTestHandler(store, newFakeAuthenticator())

	recorder := get(t, handler, "/rss/unknown?api_key="+testAPIKey)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected 404 for an unknown route, got %d", recorder.Code)
	}

	recorder = get(t, handler, "/rss?api_key="+testAPIKey)
	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected a redirect to /rss/ for a path without a trailing slash, got %d", recorder.Code)
	}
	if location := recorder.Header().Get("Location"); !strings.HasPrefix(location, "/rss/") {
		t.Errorf("unexpected redirect location: %q", location)
	}

	postRecorder := httptest.NewRecorder()
	handler.ServeHTTP(postRecorder, httptest.NewRequest(http.MethodPost, "/rss/unread?api_key="+testAPIKey, nil))
	if postRecorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST, got %d", postRecorder.Code)
	}
	if len(store.marked) != 0 {
		t.Errorf("rejected methods must not mark entries as read, got %v", store.marked)
	}
}

func TestClampLimit(t *testing.T) {
	scenarios := map[int]int{
		-10:                 defaultEntriesLimit,
		-1:                  defaultEntriesLimit,
		0:                   defaultEntriesLimit,
		1:                   1,
		defaultEntriesLimit: defaultEntriesLimit,
		maxEntriesLimit:     maxEntriesLimit,
		maxEntriesLimit + 1: maxEntriesLimit,
		1_000_000:           maxEntriesLimit,
	}

	for input, expected := range scenarios {
		if got := clampLimit(input); got != expected {
			t.Errorf("clampLimit(%d) = %d, want %d", input, got, expected)
		}
	}
}

func TestShouldMarkAsRead(t *testing.T) {
	scenarios := []struct {
		name       string
		kind       streamKind
		method     string
		target     string
		entryCount int
		expected   bool
	}{
		{"unread GET", streamUnread, http.MethodGet, "/rss/unread", 3, true},
		{"unread GET without entries", streamUnread, http.MethodGet, "/rss/unread", 0, false},
		{"unread HEAD", streamUnread, http.MethodHead, "/rss/unread", 3, false},
		{"unread GET opt-out", streamUnread, http.MethodGet, "/rss/unread?mark_read=0", 3, false},
		{"unread GET invalid opt-out value", streamUnread, http.MethodGet, "/rss/unread?mark_read=maybe", 3, true},
		{"starred GET", streamStarred, http.MethodGet, "/rss/starred", 3, false},
		{"search GET", streamSearch, http.MethodGet, "/rss/search?q=x", 3, false},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			r := httptest.NewRequest(scenario.method, scenario.target, nil)
			if got := shouldMarkAsRead(scenario.kind, r, scenario.entryCount); got != scenario.expected {
				t.Errorf("shouldMarkAsRead() = %v, want %v", got, scenario.expected)
			}
		})
	}
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"testing"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/version"
)

func TestNewHandlerHandlesOptionsRequests(t *testing.T) {
	handler := NewHandler(nil, nil)

	r := httptest.NewRequest(http.MethodOptions, "/v1/users", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if got := w.Code; got != http.StatusNoContent {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, got, http.StatusNoContent)
	}

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf(`Unexpected Access-Control-Allow-Origin header, got %q`, got)
	}

	if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Fatalf(`Unexpected Access-Control-Allow-Methods header, got %q`, got)
	}

	if got := w.Header().Get("Access-Control-Allow-Headers"); got != "X-Auth-Token, Authorization, Content-Type, Accept" {
		t.Fatalf(`Unexpected Access-Control-Allow-Headers header, got %q`, got)
	}

	if got := w.Header().Get("Access-Control-Max-Age"); got != "3600" {
		t.Fatalf(`Unexpected Access-Control-Max-Age header, got %q`, got)
	}
}

func TestVersionHandler(t *testing.T) {
	h := &handler{}
	r := httptest.NewRequest(http.MethodGet, "/v1/version", nil)
	w := httptest.NewRecorder()

	h.versionHandler(w, r)

	if got := w.Code; got != http.StatusOK {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, got, http.StatusOK)
	}

	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf(`Unexpected Content-Type header, got %q`, got)
	}

	var responseBody versionResponse
	if err := json.NewDecoder(w.Body).Decode(&responseBody); err != nil {
		t.Fatalf("Unexpected JSON decoding error: %v", err)
	}

	if responseBody.Version != version.Version {
		t.Fatalf(`Unexpected version, got %q instead of %q`, responseBody.Version, version.Version)
	}

	if responseBody.Commit != version.Commit {
		t.Fatalf(`Unexpected commit, got %q instead of %q`, responseBody.Commit, version.Commit)
	}

	if responseBody.BuildDate != version.BuildDate {
		t.Fatalf(`Unexpected build date, got %q instead of %q`, responseBody.BuildDate, version.BuildDate)
	}

	if responseBody.GoVersion != runtime.Version() {
		t.Fatalf(`Unexpected Go version, got %q instead of %q`, responseBody.GoVersion, runtime.Version())
	}

	if responseBody.Compiler != runtime.Compiler {
		t.Fatalf(`Unexpected compiler, got %q instead of %q`, responseBody.Compiler, runtime.Compiler)
	}

	if responseBody.Arch != runtime.GOARCH {
		t.Fatalf(`Unexpected architecture, got %q instead of %q`, responseBody.Arch, runtime.GOARCH)
	}

	if responseBody.OS != runtime.GOOS {
		t.Fatalf(`Unexpected OS, got %q instead of %q`, responseBody.OS, runtime.GOOS)
	}
}

func TestGetEntryIDsHandlerRequiresAuthentication(t *testing.T) {
	handler := NewHandler(nil, nil)

	for _, query := range []string{"", "?starred=maybe", "?status=invalid"} {
		t.Run(query, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/v1/entries/ids"+query, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)

			if got := w.Code; got != http.StatusUnauthorized {
				t.Fatalf(`Unexpected status code, got %d instead of %d`, got, http.StatusUnauthorized)
			}
		})
	}
}

func TestValidateEntryIDsFilters(t *testing.T) {
	for _, query := range []string{
		"", "status=read&status=unread&status=read", "tags=one&tags=two",
		"starred=false&globally_visible=true&order=category_title&direction=desc",
		"feed_id=1&category_id=1&before_entry_id=1&after_entry_id=1",
		"before=0&after=0&published_before=0&published_after=0&changed_before=0&changed_after=0",
		"before_entry_id=9223372036854775807&after_entry_id=9223372036854775807",
		"limit=invalid&offset=-1", // Legacy pagination parsing remains independent.
	} {
		t.Run("valid/"+query, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/v1/entries/ids?"+query, nil)
			if err := validateEntryIDsFilters(r); err != nil {
				t.Fatal(err)
			}
			if _, err := parseEntryFilters(r, 0, 0); err != nil {
				t.Fatal(err)
			}
		})
	}

	for _, name := range []string{
		"feed_id", "category_id", "before_entry_id", "after_entry_id",
		"before", "after", "published_before", "published_after", "changed_before", "changed_after",
	} {
		for _, value := range []string{"", "invalid", "-1", "1.5", "9223372036854775808"} {
			t.Run("invalid/"+name+"/"+value, func(t *testing.T) {
				r := httptest.NewRequest(http.MethodGet, "/v1/entries/ids?"+url.Values{name: {value}}.Encode(), nil)
				if err := validateEntryIDsFilters(r); err == nil {
					t.Fatal("Expected invalid numeric parameter to be rejected")
				}
			})
		}
	}
}

func TestValidateEntryIDsFiltersRejectsZeroIDs(t *testing.T) {
	for _, name := range []string{"feed_id", "category_id", "before_entry_id", "after_entry_id"} {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/v1/entries/ids?"+name+"=0", nil)
			if err := validateEntryIDsFilters(r); err == nil {
				t.Fatal("Expected an explicit zero ID to be rejected")
			}
		})
	}
}

func TestValidateEntryIDsFiltersRejectsMalformedQueryStrings(t *testing.T) {
	for _, query := range []string{
		"feed_id=%zz", "status=unread;status=read",
	} {
		t.Run(query, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/v1/entries/ids?"+query, nil)
			if err := validateEntryIDsFilters(r); err == nil {
				t.Fatal("Expected malformed query parameters to be rejected")
			}
		})
	}
}

func TestParseEntryFiltersRejectsValuesPostgreSQLCannotAccept(t *testing.T) {
	queries := []string{"search=%00", "search=%ff", "tags=%00", "tags=%ff", "tags=one&tags=%ff"}
	for _, name := range []string{"before", "after", "published_before", "published_after", "changed_before", "changed_after"} {
		queries = append(queries, name+"=9223372036854775807", name+"="+strconv.FormatInt(maxEntryFilterTimestamp, 10))
	}
	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/v1/entries?"+query, nil)
			if _, err := parseEntryFilters(r, 0, 0); err == nil {
				t.Fatal("Expected the value to be rejected")
			}
		})
	}

	r := httptest.NewRequest(http.MethodGet, "/v1/entries?before="+strconv.FormatInt(maxEntryFilterTimestamp-1, 10)+"&search=caf%C3%A9", nil)
	if _, err := parseEntryFilters(r, 0, 0); err != nil {
		t.Fatalf("Expected the last valid timestamp and UTF-8 search text to be accepted: %v", err)
	}
}

func TestParseEntryFiltersPreservesEndpointCompatibility(t *testing.T) {
	for _, test := range []struct {
		routeFeedID     int64
		routeCategoryID int64
		query           string
		feedID          int64
		categoryID      int64
		starred         *bool
		globallyVisible bool
	}{
		{query: "feed_id=12&category_id=20", feedID: 12, categoryID: 20},
		{query: "feed_id=invalid&category_id=-1"},
		{query: "feed_id=0&category_id=0"},
		// Route IDs scope the endpoint: the query string cannot replace or clear them.
		{routeFeedID: 11, query: "", feedID: 11},
		{routeFeedID: 11, query: "feed_id=12", feedID: 11},
		{routeFeedID: 11, query: "feed_id=0", feedID: 11},
		{routeFeedID: 11, query: "category_id=20", feedID: 11, categoryID: 20},
		{routeCategoryID: 10, query: "category_id=20", categoryID: 10},
		{routeCategoryID: 10, query: "category_id=0", categoryID: 10},
		{routeCategoryID: 10, query: "feed_id=12", feedID: 12, categoryID: 10},
		{routeFeedID: 11, query: "feed_id=&category_id=", feedID: 11},
		{query: "starred=1", starred: new(true)},
		{query: "starred=0", starred: new(false)},
		{query: "starred=invalid"},
		{query: "globally_visible=invalid", globallyVisible: true},
		{query: "globally_visible=false"},
	} {
		t.Run(fmt.Sprintf("feed=%d/category=%d/%s", test.routeFeedID, test.routeCategoryID, test.query), func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/v1/entries?"+test.query, nil)
			filters, err := parseEntryFilters(r, test.routeFeedID, test.routeCategoryID)
			if err != nil {
				t.Fatal(err)
			}
			if filters.feedID != test.feedID || filters.categoryID != test.categoryID || filters.globallyVisible != test.globallyVisible {
				t.Fatalf("Unexpected filters: %+v", filters)
			}
			if !reflect.DeepEqual(filters.starred, test.starred) {
				t.Fatalf("Unexpected starred filter: %v", filters.starred)
			}
		})
	}

	r := httptest.NewRequest(http.MethodGet, "/v1/entries?status=read&status=unread&status=read&tags=one&tags=two", nil)
	filters, err := parseEntryFilters(r, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(filters.statuses, []string{"read", "unread", "read"}) || !slices.Equal(filters.tags, []string{"one", "two"}) {
		t.Fatalf("Unexpected list filters: %+v", filters)
	}
}

func TestParseEntryFiltersRejectsInvalidRouteScopedQueryIDs(t *testing.T) {
	for _, route := range []struct {
		name               string
		feedID, categoryID int64
	}{
		{"feed", 11, 0},
		{"category", 0, 10},
	} {
		for _, query := range []string{
			"feed_id=-1",
			"category_id=-1",
			"feed_id=invalid",
			"category_id=invalid",
			"feed_id=9223372036854775808",
			"category_id=9223372036854775808",
		} {
			t.Run(route.name+"/"+query, func(t *testing.T) {
				r := httptest.NewRequest(http.MethodGet, "/v1/entries?"+query, nil)
				if _, err := parseEntryFilters(r, route.feedID, route.categoryID); err == nil {
					t.Fatal("Expected invalid scope query ID to be rejected")
				}
			})
		}
	}
}

func TestRouteScopedEntryHandlersRejectInvalidScopeQueryIDs(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		pathValues map[string]string
		handle     func(*handler, http.ResponseWriter, *http.Request)
	}{
		{
			name:       "feed route",
			path:       "/v1/feeds/11/entries?category_id=-1",
			pathValues: map[string]string{"feedID": "11"},
			handle:     (*handler).getFeedEntriesHandler,
		},
		{
			name:       "category route",
			path:       "/v1/categories/10/entries?feed_id=invalid",
			pathValues: map[string]string{"categoryID": "10"},
			handle:     (*handler).getCategoryEntriesHandler,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, test.path, nil)
			for name, value := range test.pathValues {
				r.SetPathValue(name, value)
			}
			w := httptest.NewRecorder()

			test.handle(&handler{}, w, r)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("Unexpected status code: got %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestParseEntryIDsParamsDefaults(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/entries/ids", nil)
	limit, offset := parseEntryIDsParams(r)

	if limit != 10000 {
		t.Fatalf(`Expected default limit 10000, got %d`, limit)
	}

	if offset != 0 {
		t.Fatalf(`Expected default offset 0, got %d`, offset)
	}
}

func TestParseEntryIDsParamsCustomValues(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/entries/ids?limit=500&offset=100", nil)
	limit, offset := parseEntryIDsParams(r)

	if limit != 500 {
		t.Fatalf(`Expected limit 500, got %d`, limit)
	}

	if offset != 100 {
		t.Fatalf(`Expected offset 100, got %d`, offset)
	}
}

func TestParseEntryIDsParamsLimitCappedAtMaximum(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/entries/ids?limit=99999", nil)
	limit, _ := parseEntryIDsParams(r)

	if limit != 10000 {
		t.Fatalf(`Expected limit capped at 10000, got %d`, limit)
	}
}

func TestParseEntryIDsParamsZeroLimitUsesDefault(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/entries/ids?limit=0", nil)
	limit, _ := parseEntryIDsParams(r)

	if limit != 10000 {
		t.Fatalf(`Expected zero limit to use default 10000, got %d`, limit)
	}
}

func TestNewHandlerSupportsBasePathStripping(t *testing.T) {
	scenarios := []struct {
		name   string
		prefix string
		path   string
	}{
		{name: "empty base path", prefix: "", path: "/v1/users"},
		{name: "non empty base path", prefix: "/base", path: "/base/v1/users"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			handler := http.StripPrefix(scenario.prefix, NewHandler(nil, nil))

			r := httptest.NewRequest(http.MethodOptions, scenario.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			if got := w.Code; got != http.StatusNoContent {
				t.Fatalf(`Unexpected status code, got %d instead of %d`, got, http.StatusNoContent)
			}
		})
	}
}

func TestBasicAuthIsRejectedWhenLocalAuthDisabled(t *testing.T) {
	t.Setenv("DISABLE_LOCAL_AUTH", "1")
	t.Setenv("OAUTH2_PROVIDER", "oidc")
	t.Setenv("OAUTH2_CLIENT_ID", "client")
	t.Setenv("OAUTH2_CLIENT_SECRET", "secret")
	t.Setenv("OAUTH2_REDIRECT_URL", "https://example.org/oauth2/oidc/callback")
	t.Setenv("OAUTH2_OIDC_DISCOVERY_ENDPOINT", "https://example.org")

	parsedOptions, err := config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unable to configure test options: %v", err)
	}

	previousOptions := config.Opts
	config.Opts = parsedOptions
	t.Cleanup(func() {
		config.Opts = previousOptions
	})

	handler := NewHandler(nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	r.SetBasicAuth("admin", "password")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if got := w.Code; got != http.StatusUnauthorized {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, got, http.StatusUnauthorized)
	}
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"runtime"
	"slices"
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
		"feed_id=%zz", "status=unread;status=read", "search=%00", "tags=%00", "search=%ff", "tags=%ff",
	} {
		t.Run(query, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/v1/entries/ids?"+query, nil)
			if err := validateEntryIDsFilters(r); err == nil {
				t.Fatal("Expected malformed query parameters to be rejected")
			}
		})
	}
}

func TestParseEntryFiltersPreservesEndpointCompatibility(t *testing.T) {
	for _, test := range []struct {
		query           string
		feedID          int64
		categoryID      int64
		starred         *bool
		globallyVisible bool
	}{
		{query: "", feedID: 11, categoryID: 10},
		{query: "feed_id=12&category_id=20", feedID: 12, categoryID: 20},
		{query: "feed_id=invalid&category_id=-1", feedID: 11, categoryID: 10},
		{query: "feed_id=0&category_id=0"},
		{query: "starred=1", feedID: 11, categoryID: 10, starred: new(true)},
		{query: "starred=0", feedID: 11, categoryID: 10, starred: new(false)},
		{query: "starred=invalid", feedID: 11, categoryID: 10},
		{query: "globally_visible=invalid", feedID: 11, categoryID: 10, globallyVisible: true},
		{query: "globally_visible=false", feedID: 11, categoryID: 10},
	} {
		t.Run(test.query, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/v1/entries?"+test.query, nil)
			filters, err := parseEntryFilters(r, 11, 10)
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

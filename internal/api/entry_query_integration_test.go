// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"cmp"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/storage"
)

func newEntryQueryAPI(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("Set TEST_DATABASE_URL to a migrated PostgreSQL database to run API query integration tests")
	}

	previousOptions := config.Opts
	config.Opts = config.NewConfigOptions()
	t.Cleanup(func() { config.Opts = previousOptions })

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	// Temporary tables stay on one connection and never modify the shared fixture database.
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`
		SET TIME ZONE 'UTC';
		CREATE TEMP TABLE users (LIKE users INCLUDING DEFAULTS);
		CREATE TEMP TABLE categories (LIKE categories INCLUDING DEFAULTS);
		CREATE TEMP TABLE feeds (LIKE feeds INCLUDING DEFAULTS);
		CREATE TEMP TABLE entries (LIKE entries INCLUDING DEFAULTS);
		CREATE TEMP TABLE enclosures (LIKE enclosures INCLUDING DEFAULTS);
		CREATE TEMP TABLE feed_icons (LIKE feed_icons INCLUDING DEFAULTS);
		CREATE TEMP TABLE icons (LIKE icons INCLUDING DEFAULTS);

		INSERT INTO users (id, username) VALUES (1, 'entry-query-user'), (2, 'entry-query-other');
		INSERT INTO categories (id, user_id, title, hide_globally) VALUES
			(10, 1, 'Zulu category', false), (20, 1, 'Alpha category', false),
			(30, 1, 'Hidden category', true), (40, 2, 'Other category', false);
		INSERT INTO feeds (id, user_id, category_id, title, feed_url, site_url, hide_globally) VALUES
			(11, 1, 10, 'Visible feed', 'https://example.org/11', 'https://example.org', false),
			(12, 1, 20, 'Other visible feed', 'https://example.org/12', 'https://example.org', false),
			(13, 1, 10, 'Hidden feed', 'https://example.org/13', 'https://example.org', true),
			(14, 1, 30, 'Visible feed in hidden category', 'https://example.org/14', 'https://example.org', false),
			(15, 1, 30, 'Hidden feed in hidden category', 'https://example.org/15', 'https://example.org', true),
			(16, 2, 40, 'Other user feed', 'https://example.org/16', 'https://example.org', false);
		INSERT INTO entries (id, user_id, feed_id, hash, title, url, author, content, status, starred,
			published_at, created_at, changed_at, document_vectors, tags) VALUES
			(1, 1, 11, '1', 'Zulu entry', 'https://example.org/1', 'Bravo', 'shared content', 'unread', false,
				'2025-01-03', '2025-01-02', '2025-01-04', to_tsvector('shared'), ARRAY['one', 'two']),
			(2, 1, 12, '2', 'Alpha entry', 'https://example.org/2', 'Charlie', 'shared content', 'read', true,
				'2025-01-01', '2025-01-03', '2025-01-02', to_tsvector('shared'), ARRAY['one']),
			(3, 1, 11, '3', 'Alpha entry', 'https://example.org/3', 'Alpha', 'shared content', 'unread', false,
				'2025-01-02', '2025-01-01', '2025-01-03', to_tsvector('shared'), ARRAY['one', 'two']),
			(4, 1, 13, '4', 'Hidden feed entry', 'https://example.org/4', '', '', 'unread', true,
				'2025-01-01', '2025-01-01', '2025-01-01', to_tsvector('hidden'), ARRAY[]::text[]),
			(5, 1, 14, '5', 'Hidden category entry', 'https://example.org/5', '', '', 'read', false,
				'2025-01-01', '2025-01-01', '2025-01-01', to_tsvector('hidden'), ARRAY[]::text[]),
			(6, 1, 15, '6', 'Both hidden entry', 'https://example.org/6', '', '', 'unread', false,
				'2025-01-01', '2025-01-01', '2025-01-01', to_tsvector('hidden'), ARRAY[]::text[]),
			(7, 2, 16, '7', 'Other user entry', 'https://example.org/7', '', '', 'unread', true,
				'2025-01-01', '2025-01-01', '2025-01-01', to_tsvector('shared'), ARRAY['one', 'two']),
			(8, 1, 11, '8', 'Alpha entry', 'https://example.org/8', 'Alpha', 'shared content', 'unread', false,
				'2025-01-02', '2025-01-01', '2025-01-03', to_tsvector('shared'), ARRAY['one', 'two']);
	`)
	if err != nil {
		t.Fatal(err)
	}

	password, err := bcrypt.GenerateFromPassword([]byte("entry-query-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE users SET password=$1`, string(password)); err != nil {
		t.Fatal(err)
	}

	return NewHandler(storage.NewStorage(db), nil), db
}

func entryQueryAPIRequest(api http.Handler, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.SetBasicAuth("entry-query-user", "entry-query-password")
	w := httptest.NewRecorder()
	api.ServeHTTP(w, r)
	return w
}

func checkEntryQueryAPIs(t *testing.T, api http.Handler, query string, want []int64, total int) {
	t.Helper()

	values, err := url.ParseQuery(query)
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{"order": "id", "direction": "asc", "limit": "100", "offset": "0"} {
		if !values.Has(name) {
			values.Set(name, value)
		}
	}
	query = values.Encode()

	w := entryQueryAPIRequest(api, http.MethodGet, "/v1/entries?"+query, "")
	if w.Code != http.StatusOK {
		t.Fatalf("GetEntries returned %d: %s", w.Code, w.Body.String())
	}
	var entries entriesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &entries); err != nil {
		t.Fatal(err)
	}
	ids := make([]int64, len(entries.Entries))
	for i, entry := range entries.Entries {
		ids[i] = entry.ID
	}

	w = entryQueryAPIRequest(api, http.MethodGet, "/v1/entries/ids?"+query, "")
	if w.Code != http.StatusOK {
		t.Fatalf("GetEntryIDs returned %d: %s", w.Code, w.Body.String())
	}
	var entryIDs entryIDsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &entryIDs); err != nil {
		t.Fatal(err)
	}
	if entryIDs.EntryIDs == nil {
		t.Fatal("Expected entry_ids to be a JSON array, not null")
	}
	if !slices.Equal(ids, entryIDs.EntryIDs) || entries.Total != entryIDs.Total {
		t.Fatalf("Queries differ: entries=%v total=%d, IDs=%v total=%d", ids, entries.Total, entryIDs.EntryIDs, entryIDs.Total)
	}
	if !slices.Equal(ids, want) || entries.Total != total {
		t.Fatalf("Got IDs %v total=%d, want %v total=%d", ids, entries.Total, want, total)
	}
}

func TestEntryQueryAPIFilters(t *testing.T) {
	api, _ := newEntryQueryAPI(t)
	all := []int64{1, 2, 3, 4, 5, 6, 8}

	for _, test := range []struct {
		name  string
		query string
		ids   []int64
	}{
		{"all including hidden", "", all},
		{"feed", "feed_id=11", []int64{1, 3, 8}},
		{"category", "category_id=10", []int64{1, 3, 4, 8}},
		{"unread", "status=unread", []int64{1, 3, 4, 6, 8}},
		{"read", "status=read", []int64{2, 5}},
		{"multiple statuses", "status=unread&status=read", all},
		{"duplicate statuses", "status=unread&status=unread", []int64{1, 3, 4, 6, 8}},
		{"mixed duplicate statuses", "status=read&status=unread&status=read", all},
		{"trimmed statuses", "status=%20unread%20", []int64{1, 3, 4, 6, 8}},
		{"starred", "starred=true", []int64{2, 4}},
		{"not starred", "starred=false", []int64{1, 3, 5, 6, 8}},
		{"tag", "tags=one", []int64{1, 2, 3, 8}},
		{"multiple tags", "tags=ONE&tags=two", []int64{1, 3, 8}},
		{"duplicate tags", "tags=one&tags=one", []int64{1, 2, 3, 8}},
		{"trimmed tags", "tags=%20one%20&tags=two&tags=", []int64{1, 3, 8}},
		{"globally visible", "globally_visible=true", []int64{1, 2, 3, 8}},
		{"include hidden explicitly", "globally_visible=false", all},
		{"hidden feed", "feed_id=13", []int64{4}},
		{"hidden category", "category_id=30", []int64{5, 6}},
		{"both hidden", "feed_id=15&category_id=30", []int64{6}},
		{"exclude hidden feed", "feed_id=13&globally_visible=true", nil},
		{"exclude hidden category", "category_id=30&globally_visible=true", nil},
		{"exclude both hidden", "feed_id=15&category_id=30&globally_visible=true", nil},
		{"matching scope", "feed_id=11&category_id=10", []int64{1, 3, 8}},
		{"mismatched owned scope", "feed_id=11&category_id=20", nil},
		{"before ID is exclusive", "before_entry_id=3", []int64{1, 2}},
		{"after ID is exclusive", "after_entry_id=3", []int64{4, 5, 6, 8}},
		{"combined ID bounds", "after_entry_id=1&before_entry_id=3", []int64{2}},
		{"before alias is exclusive", "before=1735776000", []int64{2, 4, 5, 6}},
		{"after alias is exclusive", "after=1735776000", []int64{1}},
		{"published before is exclusive", "published_before=1735776000", []int64{2, 4, 5, 6}},
		{"published after is exclusive", "published_after=1735776000", []int64{1}},
		{"changed before is exclusive", "changed_before=1735862400", []int64{2, 4, 5, 6}},
		{"changed after is exclusive", "changed_after=1735862400", []int64{1}},
		{"both publication aliases apply", "before=1735862400&published_before=1735776000", []int64{2, 4, 5, 6}},
		{"both lower aliases apply", "after=1735689600&published_after=1735776000", []int64{1}},
		{"publication interval", "published_after=1735689600&published_before=1735862400", []int64{3, 8}},
		{"changed interval", "changed_after=1735776000&changed_before=1735948800", []int64{3, 8}},
		{"zero time filters stay inactive", "before=0&after=0&published_before=0&published_after=0&changed_before=0&changed_after=0", all},
		{"search", "search=shared", []int64{1, 2, 3, 8}},
		{"search hidden", "search=hidden", []int64{4, 5, 6}},
		{"empty search", "search=", all},
		{"empty result", "search=missingterm", nil},
		{"no matching tags", "tags=missing", nil},
		{"empty ID interval", "after_entry_id=8", nil},
		{"conflicting bounds", "before_entry_id=3&after_entry_id=3", nil},
		{"combined filters", "feed_id=11&category_id=10&status=unread&starred=false&tags=one&tags=two&globally_visible=true&search=shared&after_entry_id=1&before_entry_id=8&after=1735689600&before=1735862400&changed_after=1735776000&changed_before=1735948800", []int64{3}},
		{"read starred scope", "feed_id=12&category_id=20&status=read&starred=true&search=shared", []int64{2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			checkEntryQueryAPIs(t, api, test.query, test.ids, len(test.ids))
		})
	}

	for _, test := range []struct {
		query string
		ids   []int64
		total int
	}{
		{"limit=2", []int64{1, 2}, 7},
		{"offset=2", []int64{3, 4, 5, 6, 8}, 7},
		{"limit=2&offset=1", []int64{2, 3}, 7},
		{"offset=999", nil, 7},
		{"after_entry_id=1&before_entry_id=8&limit=2&offset=1", []int64{3, 4}, 5},
		{"after_entry_id=1&before_entry_id=3&offset=999", nil, 1},
		{"published_after=1735689600&published_before=1735862400&limit=1", []int64{3}, 2},
		{"changed_after=1735776000&changed_before=1735948800&offset=999", nil, 2},
	} {
		t.Run(test.query, func(t *testing.T) {
			checkEntryQueryAPIs(t, api, test.query, test.ids, test.total)
		})
	}
}

func TestEntryQueryAPISorting(t *testing.T) {
	api, _ := newEntryQueryAPI(t)
	visible := []int64{1, 2, 3, 8}
	// Ordinal values describe the fixed fixture, independently of either SQL query.
	keys := map[string][]int{
		"id":             {1, 2, 3, 8},
		"status":         {0, 1, 0, 0},
		"changed_at":     {4, 2, 3, 3},
		"published_at":   {3, 1, 2, 2},
		"created_at":     {2, 3, 1, 1},
		"category_title": {2, 1, 2, 2},
		"category_id":    {10, 20, 10, 10},
		"title":          {2, 1, 1, 1},
		"author":         {2, 3, 1, 1},
	}
	for order, key := range keys {
		for _, direction := range []string{"asc", "desc"} {
			for _, search := range []string{"", "shared"} {
				t.Run(order+"/"+direction+"/search="+search, func(t *testing.T) {
					want := slices.Clone(visible)
					slices.SortFunc(want, func(a, b int64) int {
						i, j := slices.Index(visible, a), slices.Index(visible, b)
						result := cmp.Compare(key[i], key[j])
						if result == 0 && search != "" {
							// Equal vectors leave the search expression's publication recency first.
							if rank := cmp.Compare(keys["published_at"][j], keys["published_at"][i]); rank != 0 {
								return rank
							}
						}
						if result == 0 {
							result = cmp.Compare(a, b)
						}
						if direction == "desc" {
							result = -result
						}
						return result
					})
					query := url.Values{"order": {order}, "direction": {direction}, "search": {search}, "globally_visible": {"true"}}.Encode()
					checkEntryQueryAPIs(t, api, query, want, len(want))
				})
			}
		}
	}
}

func TestEntryQueryAPISearchRankPrecedesIDTiebreak(t *testing.T) {
	api, db := newEntryQueryAPI(t)
	if _, err := db.Exec(`DELETE FROM entries WHERE id NOT IN (1, 8);
		UPDATE entries SET title='Shared title', published_at='2025-01-02', status='unread';`); err != nil {
		t.Fatal(err)
	}

	for _, direction := range []string{"asc", "desc"} {
		highRankID, lowRankID := int64(8), int64(1)
		if direction == "desc" {
			highRankID, lowRankID = lowRankID, highRankID
		}
		if _, err := db.Exec(`UPDATE entries SET document_vectors = CASE WHEN id=$1
			THEN setweight(to_tsvector('shared'), 'A') ELSE setweight(to_tsvector('shared'), 'D') END`, highRankID); err != nil {
			t.Fatal(err)
		}
		for _, order := range []string{"published_at", "title", "status"} {
			t.Run(order+"/"+direction, func(t *testing.T) {
				query := "search=shared&order=" + order + "&direction=" + direction
				checkEntryQueryAPIs(t, api, query, []int64{highRankID, lowRankID}, 2)
				checkEntryQueryAPIs(t, api, query+"&limit=1", []int64{highRankID}, 2)
				checkEntryQueryAPIs(t, api, query+"&limit=1&offset=1", []int64{lowRankID}, 2)
			})
		}
	}
}

func TestEntryQueryAPIStablePagination(t *testing.T) {
	api, db := newEntryQueryAPI(t)
	const count, pageSize = 120, 25
	if _, err := db.Exec(`DELETE FROM entries;
		INSERT INTO entries (id, user_id, feed_id, hash, title, url, author, content, status, published_at, created_at, changed_at, document_vectors)
		SELECT id, 1, 11, id::text, 'Shared pagination title', 'https://example.org/' || id, '', '',
			'unread', '2025-01-02'::timestamptz, '2025-01-02'::timestamptz, '2025-01-02'::timestamptz,
			to_tsvector('pagination') FROM generate_series(1, 120) id;`); err != nil {
		t.Fatal(err)
	}

	w := entryQueryAPIRequest(api, http.MethodGet, "/v1/entries", "")
	var entries entriesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &entries); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || len(entries.Entries) != 100 || entries.Total != count {
		t.Fatalf("GetEntries lost its default page size: status=%d length=%d total=%d", w.Code, len(entries.Entries), entries.Total)
	}
	w = entryQueryAPIRequest(api, http.MethodGet, "/v1/entries/ids", "")
	var ids entryIDsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &ids); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || len(ids.EntryIDs) != count || ids.Total != count {
		t.Fatalf("GetEntryIDs unexpectedly inherited the entries page size: status=%d length=%d total=%d", w.Code, len(ids.EntryIDs), ids.Total)
	}

	for _, order := range []string{"published_at", "title", "status"} {
		for _, direction := range []string{"asc", "desc"} {
			for _, search := range []string{"", "pagination"} {
				t.Run(order+"/"+direction+"/search="+search, func(t *testing.T) {
					want := make([]int64, count)
					for i := range count {
						want[i] = int64(i + 1)
					}
					if direction == "desc" {
						slices.Reverse(want)
					}
					for offset := 0; offset <= count; offset += pageSize {
						query := url.Values{
							"order": {order}, "direction": {direction}, "search": {search},
							"offset": {strconv.Itoa(offset)}, "limit": {strconv.Itoa(pageSize)},
						}.Encode()
						checkEntryQueryAPIs(t, api, query, want[offset:min(offset+pageSize, count)], count)
					}
				})
			}
		}
	}
}

func TestEntryQueryAPIRejectsInvalidParametersAfterAuthentication(t *testing.T) {
	api, _ := newEntryQueryAPI(t)
	// Prove these credentials reach the handler before testing its validation errors.
	if w := entryQueryAPIRequest(api, http.MethodGet, "/v1/me", ""); w.Code != http.StatusOK {
		t.Fatalf("Authentication failed: %d %s", w.Code, w.Body.String())
	}

	queries := []string{
		"status=unread&feed_id=0", "status=unread&category_id=0",
		"status=unread&before_entry_id=0", "status=unread&after_entry_id=0",
		"feed_id=%zz", "status=unread;status=read", "search=%00", "tags=%00", "search=%ff", "tags=%ff",
		"starred=", "starred=maybe", "starred=1", "starred=TRUE",
		"globally_visible=", "globally_visible=maybe", "globally_visible=1", "globally_visible=TRUE",
		"status=", "status=%20", "status=removed", "status=unread&status=invalid",
		"status=invalid&status=unread", "status=unread&status=",
		"order=", "order=unknown", "order=id%3BDROP%20TABLE%20entries",
		"direction=", "direction=up", "direction=ASC",
	}
	for _, name := range []string{
		"feed_id", "category_id", "before_entry_id", "after_entry_id",
		"before", "after", "published_before", "published_after", "changed_before", "changed_after",
	} {
		for _, value := range []string{"", "invalid", "-1", "1.5", "9223372036854775808"} {
			queries = append(queries, url.Values{name: {value}}.Encode())
		}
		queries = append(queries, url.Values{name: {"1", "2"}}.Encode())
	}
	for _, name := range []string{"before", "after", "published_before", "published_after", "changed_before", "changed_after"} {
		queries = append(queries, name+"=9223372036854775807")
	}
	for _, name := range []string{"order", "direction", "starred", "globally_visible", "search"} {
		queries = append(queries, url.Values{name: {"first", "second"}}.Encode())
	}
	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			w := entryQueryAPIRequest(api, http.MethodGet, "/v1/entries/ids?"+query, "")
			if w.Code != http.StatusBadRequest {
				t.Fatalf("Got status %d, want 400: %s", w.Code, w.Body.String())
			}
		})
	}

	for _, query := range []string{"status=invalid", "status=unread&status=invalid", "order=invalid", "direction=invalid", "limit=1001"} {
		t.Run("entries/"+query, func(t *testing.T) {
			w := entryQueryAPIRequest(api, http.MethodGet, "/v1/entries?"+query, "")
			if w.Code != http.StatusBadRequest {
				t.Fatalf("Got status %d, want 400: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestEntryQueryAPIRejectsForeignScopes(t *testing.T) {
	api, _ := newEntryQueryAPI(t)
	for _, endpoint := range []string{"/v1/entries", "/v1/entries/ids"} {
		for _, query := range []string{
			"feed_id=16", "category_id=40", "feed_id=16&category_id=40",
			"feed_id=16&category_id=10", "feed_id=11&category_id=40",
			"feed_id=999", "category_id=999",
		} {
			t.Run(endpoint+"/"+query, func(t *testing.T) {
				w := entryQueryAPIRequest(api, http.MethodGet, endpoint+"?"+query, "")
				if w.Code != http.StatusBadRequest {
					t.Fatalf("Got status %d, want 400: %s", w.Code, w.Body.String())
				}
			})
		}
	}
	for _, path := range []string{"/v1/feeds/16/entries", "/v1/categories/40/entries"} {
		w := entryQueryAPIRequest(api, http.MethodGet, path, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("Got status %d, want 400: %s", w.Code, w.Body.String())
		}
	}
}

func TestEntryQueryAPICompatibility(t *testing.T) {
	api, _ := newEntryQueryAPI(t)
	for _, test := range []struct {
		path string
		ids  []int64
	}{
		{"/v1/entries/ids", []int64{8, 6, 5, 4, 3, 2, 1}},
		{"/v1/entries/ids?status=read", []int64{5, 2}},
		{"/v1/entries/ids?starred=true", []int64{4, 2}},
		{"/v1/entries/ids?starred=false&status=unread", []int64{8, 6, 3, 1}},
		{"/v1/entries/ids?limit=0", []int64{8, 6, 5, 4, 3, 2, 1}},
		{"/v1/entries/ids?limit=99999", []int64{8, 6, 5, 4, 3, 2, 1}},
		{"/v1/entries/ids?limit=invalid&offset=-1", []int64{8, 6, 5, 4, 3, 2, 1}},
	} {
		t.Run(test.path, func(t *testing.T) {
			w := entryQueryAPIRequest(api, http.MethodGet, test.path, "")
			if w.Code != http.StatusOK {
				t.Fatalf("Got status %d: %s", w.Code, w.Body.String())
			}
			var result entryIDsResponse
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(result.EntryIDs, test.ids) || result.Total != len(test.ids) {
				t.Fatalf("Got IDs %v total=%d, want %v total=%d", result.EntryIDs, result.Total, test.ids, len(test.ids))
			}
		})
	}

	for _, test := range []struct {
		path string
		ids  []int64
	}{
		{"/v1/entries", []int64{2, 4, 5, 6, 3, 8, 1}},
		{"/v1/feeds/11/entries", []int64{3, 8, 1}},
		{"/v1/categories/10/entries", []int64{4, 3, 8, 1}},
		{"/v1/feeds/12/entries?feed_id=11", []int64{3, 8, 1}},
		{"/v1/categories/20/entries?category_id=10", []int64{4, 3, 8, 1}},
		{"/v1/entries?starred=1", []int64{2, 4}},
		{"/v1/entries?starred=invalid", []int64{2, 4, 5, 6, 3, 8, 1}},
		{"/v1/entries?feed_id=invalid&before_entry_id=-1", []int64{2, 4, 5, 6, 3, 8, 1}},
		{"/v1/entries?feed_id=0&category_id=0&before_entry_id=0&after_entry_id=0", []int64{2, 4, 5, 6, 3, 8, 1}},
		{"/v1/entries?globally_visible=invalid", []int64{2, 3, 8, 1}},
	} {
		t.Run(test.path, func(t *testing.T) {
			w := entryQueryAPIRequest(api, http.MethodGet, test.path, "")
			if w.Code != http.StatusOK {
				t.Fatalf("Got status %d: %s", w.Code, w.Body.String())
			}
			var result entriesResponse
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			ids := make([]int64, len(result.Entries))
			for i, entry := range result.Entries {
				ids[i] = entry.ID
			}
			if !slices.Equal(ids, test.ids) || result.Total != len(test.ids) {
				t.Fatalf("Got IDs %v total=%d, want %v total=%d", ids, result.Total, test.ids, len(test.ids))
			}
		})
	}
}

func TestEntryQueryAPIUnreadIDBatchesSurviveStatusUpdates(t *testing.T) {
	api, _ := newEntryQueryAPI(t)
	const query = "/v1/entries/ids?status=unread&order=id&direction=desc"
	readPage := func(path string) entryIDsResponse {
		t.Helper()
		w := entryQueryAPIRequest(api, http.MethodGet, path, "")
		if w.Code != http.StatusOK {
			t.Fatalf("Got status %d: %s", w.Code, w.Body.String())
		}
		var page entryIDsResponse
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
			t.Fatal(err)
		}
		return page
	}

	original := readPage(query)
	if !slices.Equal(original.EntryIDs, []int64{8, 6, 4, 3, 1}) {
		t.Fatalf("Unexpected target set: %v", original.EntryIDs)
	}
	var got []int64
	var before int64
	for {
		path := query + "&limit=2"
		if before > 0 {
			path += "&before_entry_id=" + strconv.FormatInt(before, 10)
		}
		page := readPage(path)
		if remaining := len(original.EntryIDs) - len(got); page.Total != remaining {
			t.Fatalf("Got remaining total %d, want %d", page.Total, remaining)
		}
		if len(page.EntryIDs) == 0 {
			break
		}
		got = append(got, page.EntryIDs...)
		if len(got) > len(original.EntryIDs) {
			t.Fatal("Pagination repeated entries instead of making progress")
		}
		body, err := json.Marshal(map[string]any{"entry_ids": page.EntryIDs, "status": "read"})
		if err != nil {
			t.Fatal(err)
		}
		w := entryQueryAPIRequest(api, http.MethodPut, "/v1/entries", string(body))
		if w.Code != http.StatusNoContent {
			t.Fatalf("Unable to mark batch read: %d %s", w.Code, w.Body.String())
		}
		before = page.EntryIDs[len(page.EntryIDs)-1]
	}
	if !slices.Equal(got, original.EntryIDs) {
		t.Fatalf("Missed or repeated entries after status updates: got %v, want %v", got, original.EntryIDs)
	}
	if unread := readPage(query); unread.Total != 0 || len(unread.EntryIDs) != 0 {
		t.Fatalf("Unread entries remain after consuming all batches: %+v", unread)
	}
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"database/sql"
	"os"
	"slices"
	"testing"
	"time"

	"miniflux.app/v2/internal/model"
)

func newEntryQueryTestStorage(t *testing.T) *Storage {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("Set TEST_DATABASE_URL to a migrated PostgreSQL database to run storage integration tests")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	// Keep temporary tables on the same connection and leave the database unchanged.
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`
		SET TIME ZONE 'UTC';
		CREATE TEMP TABLE users (LIKE users INCLUDING DEFAULTS);
		CREATE TEMP TABLE categories (LIKE categories INCLUDING DEFAULTS);
		CREATE TEMP TABLE feeds (LIKE feeds INCLUDING DEFAULTS);
		CREATE TEMP TABLE entries (LIKE entries INCLUDING DEFAULTS);
		CREATE TEMP TABLE feed_icons (LIKE feed_icons INCLUDING DEFAULTS);
		CREATE TEMP TABLE icons (LIKE icons INCLUDING DEFAULTS);

		INSERT INTO users (id, username) VALUES (1, 'entry-query-user'), (2, 'entry-query-other');
		INSERT INTO categories (id, user_id, title, hide_globally) VALUES
			(10, 1, 'Zulu category', false),
			(20, 1, 'Alpha category', false),
			(30, 1, 'Hidden category', true),
			(40, 2, 'Other category', false);
		INSERT INTO feeds (id, user_id, category_id, title, feed_url, site_url, hide_globally) VALUES
			(11, 1, 10, 'Alpha feed', 'https://example.org/11', 'https://example.org', false),
			(12, 1, 20, 'Zulu feed', 'https://example.org/12', 'https://example.org', false),
			(13, 1, 10, 'Hidden feed', 'https://example.org/13', 'https://example.org', true),
			(14, 1, 30, 'Visible feed in hidden category', 'https://example.org/14', 'https://example.org', false),
			(15, 1, 30, 'Hidden feed in hidden category', 'https://example.org/15', 'https://example.org', true),
			(16, 2, 40, 'Other feed', 'https://example.org/16', 'https://example.org', false);
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
				'2025-01-01', '2025-01-01', '2025-01-01', to_tsvector('shared'), ARRAY['one', 'two']);
	`)
	if err != nil {
		t.Fatal(err)
	}

	return NewStorage(db)
}

func TestGetEntryIDsSorting(t *testing.T) {
	store := newEntryQueryTestStorage(t)

	for _, test := range []struct {
		order string
		ids   []int64
	}{
		{"id", []int64{1, 2, 3}},
		{"status", []int64{1, 3, 2}},
		{"changed_at", []int64{2, 3, 1}},
		{"published_at", []int64{2, 3, 1}},
		{"created_at", []int64{3, 1, 2}},
		{"category_title", []int64{2, 1, 3}},
		{"category_id", []int64{1, 3, 2}},
		{"title", []int64{2, 3, 1}},
		{"author", []int64{3, 1, 2}},
	} {
		for _, direction := range []string{"asc", "desc"} {
			t.Run(test.order+"/"+direction, func(t *testing.T) {
				want := slices.Clone(test.ids)
				if direction == "desc" {
					slices.Reverse(want)
				}

				builder := store.NewEntryQueryBuilder(1).
					WithEntryIDs(1, 2, 3).
					WithSorting(test.order, direction).
					WithSorting("id", direction)
				ids, err := builder.GetEntryIDs()
				if err != nil {
					t.Fatal(err)
				}
				if !slices.Equal(ids, want) {
					t.Fatalf("Got IDs %v, want %v", ids, want)
				}

				entries, err := builder.GetEntries()
				if err != nil {
					t.Fatal(err)
				}
				if len(entries) != len(ids) {
					t.Fatalf("Got %d entries, want %d", len(entries), len(ids))
				}
				for i, entry := range entries {
					if entry.ID != ids[i] {
						t.Fatalf("Entry %d has ID %d, want %d", i, entry.ID, ids[i])
					}
				}
			})
		}
	}
}

func TestGetEntryIDsFiltersAndCount(t *testing.T) {
	store := newEntryQueryTestStorage(t)

	for _, test := range []struct {
		name      string
		configure func(*EntryQueryBuilder) *EntryQueryBuilder
		ids       []int64
	}{
		{"all including hidden", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b }, []int64{1, 2, 3, 4, 5, 6}},
		{"globally visible", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithGloballyVisible() }, []int64{1, 2, 3}},
		{"hidden feed", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithFeedID(13).WithGloballyVisible() }, nil},
		{"hidden category", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithCategoryID(30).WithGloballyVisible() }, nil},
		{"matching scope", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithFeedID(11).WithCategoryID(10) }, []int64{1, 3}},
		{"mismatched scope", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithFeedID(11).WithCategoryID(20) }, nil},
		{"unread", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithStatuses(model.EntryStatusUnread) }, []int64{1, 3, 4, 6}},
		{"read", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithStatuses(model.EntryStatusRead) }, []int64{2, 5}},
		{"both statuses", func(b *EntryQueryBuilder) *EntryQueryBuilder {
			return b.WithStatuses(model.EntryStatusRead, model.EntryStatusUnread)
		}, []int64{1, 2, 3, 4, 5, 6}},
		{"starred", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithStarred(true) }, []int64{2, 4}},
		{"not starred", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithStarred(false) }, []int64{1, 3, 5, 6}},
		{"combined", func(b *EntryQueryBuilder) *EntryQueryBuilder {
			return b.WithGloballyVisible().WithStatuses(model.EntryStatusUnread).WithTags("ONE", "two").WithSearchQuery("shared")
		}, []int64{1, 3}},
		{"strict ID boundaries", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.AfterEntryID(1).BeforeEntryID(3) }, []int64{2}},
		{"strict time boundaries", func(b *EntryQueryBuilder) *EntryQueryBuilder {
			return b.AfterPublishedDate(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)).
				BeforePublishedDate(time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC))
		}, []int64{3}},
	} {
		t.Run(test.name, func(t *testing.T) {
			builder := test.configure(store.NewEntryQueryBuilder(1).WithSorting("id", "asc"))
			ids, total, err := builder.GetEntryIDsWithCount()
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(ids, test.ids) || total != len(test.ids) {
				t.Fatalf("Got IDs %v and total %d, want %v and %d", ids, total, test.ids, len(test.ids))
			}

			for offset := 0; offset <= len(test.ids); offset++ {
				builder.WithLimitAndMaximum(1, model.MaxEntryIDsLimit).WithOffset(offset)
				ids, total, err := builder.GetEntryIDsWithCount()
				if err != nil {
					t.Fatal(err)
				}
				want := test.ids[offset:min(offset+1, len(test.ids))]
				if !slices.Equal(ids, want) || total != len(test.ids) {
					t.Fatalf("Offset %d: got IDs %v and total %d, want %v and %d", offset, ids, total, want, len(test.ids))
				}
			}
		})
	}
}

func TestGetEntryIDsStablePagination(t *testing.T) {
	const pageSize = 25

	store := newEntryQueryTestStorage(t)
	_, err := store.db.Exec(`
		DELETE FROM entries;
		INSERT INTO entries (id, user_id, feed_id, hash, title, url, author, content,
			published_at, created_at, changed_at, document_vectors)
		SELECT n, 1, 11, n::text, 'Shared title', 'https://example.org/' || n, 'Shared author', '',
			'2025-01-01', '2025-01-01', '2025-01-01', to_tsvector('shared')
		FROM generate_series(1, 120) AS n;
	`)
	if err != nil {
		t.Fatal(err)
	}

	for _, order := range []string{"id", "status", "changed_at", "published_at", "created_at", "category_title", "category_id", "title", "author"} {
		for _, direction := range []string{"asc", "desc"} {
			want := make([]int64, 120)
			for i := range want {
				want[i] = int64(i + 1)
			}
			if direction == "desc" {
				slices.Reverse(want)
			}

			for _, search := range []string{"", "shared"} {
				t.Run(order+"/"+direction+"/search="+search, func(t *testing.T) {
					var got []int64
					for offset := 0; offset < len(want); offset += pageSize {
						ids, total, err := store.NewEntryQueryBuilder(1).
							WithSorting(order, direction).
							WithSearchQuery(search).
							WithSorting("id", direction).
							WithLimit(pageSize).
							WithOffset(offset).
							GetEntryIDsWithCount()
						if err != nil {
							t.Fatal(err)
						}
						if total != len(want) {
							t.Fatalf("Got total %d, want %d", total, len(want))
						}
						got = append(got, ids...)
					}
					if !slices.Equal(got, want) {
						t.Fatalf("Got IDs %v, want %v", got, want)
					}
				})
			}
		}
	}
}

func TestGetEntryIDsSearchSorting(t *testing.T) {
	store := newEntryQueryTestStorage(t)

	for _, test := range []struct {
		direction string
		ids       []int64
	}{
		{"asc", []int64{3, 2, 1}},
		{"desc", []int64{1, 3, 2}},
	} {
		t.Run(test.direction, func(t *testing.T) {
			ids, err := store.NewEntryQueryBuilder(1).
				WithSorting("title", test.direction).
				WithSearchQuery("shared").
				WithSorting("id", test.direction).
				GetEntryIDs()
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(ids, test.ids) {
				t.Fatalf("Got IDs %v, want %v", ids, test.ids)
			}
		})
	}
}

func TestGetEntryIDsWithoutEntryData(t *testing.T) {
	store := newEntryQueryTestStorage(t)
	if _, err := store.db.Exec(`ALTER TABLE entries DROP COLUMN content`); err != nil {
		t.Fatal(err)
	}

	ids, err := store.NewEntryQueryBuilder(1).
		WithEntryIDs(1, 2, 3).
		WithSorting("category_title", "asc").
		WithSorting("id", "asc").
		WithEnclosures().
		GetEntryIDs()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(ids, []int64{2, 1, 3}) {
		t.Fatalf("Got IDs %v, want [2 1 3]", ids)
	}
}

func TestGetEntryIDsCompatibilityQueries(t *testing.T) {
	store := newEntryQueryTestStorage(t)

	// Fever fetches unread and starred IDs without sorting or pagination.
	for _, test := range []struct {
		builder *EntryQueryBuilder
		ids     []int64
	}{
		{store.NewEntryQueryBuilder(1).WithStatuses(model.EntryStatusUnread), []int64{1, 3, 4, 6}},
		{store.NewEntryQueryBuilder(1).WithStarred(true), []int64{2, 4}},
	} {
		ids, err := test.builder.GetEntryIDs()
		if err != nil {
			t.Fatal(err)
		}
		slices.Sort(ids)
		if !slices.Equal(ids, test.ids) {
			t.Fatalf("Got IDs %v, want %v", ids, test.ids)
		}
	}

	// Google Reader sorts IDs by publication date and uses limit/offset.
	ids, err := store.NewEntryQueryBuilder(1).
		WithFeedID(11).
		WithSorting(model.DefaultSortingOrder, "desc").
		WithLimitAndMaximum(1, model.MaxEntryIDsLimit).
		WithOffset(1).
		GetEntryIDs()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(ids, []int64{3}) {
		t.Fatalf("Got IDs %v, want [3]", ids)
	}
}

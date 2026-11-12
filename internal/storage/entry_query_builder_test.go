// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestEntryQuerySortingColumns(t *testing.T) {
	for _, test := range []struct {
		column          string
		qualifiedColumn string
	}{
		{"id", `"e"."id"`},
		{"status", `"e"."status"`},
		{"changed_at", `"e"."changed_at"`},
		{"published_at", `"e"."published_at"`},
		{"created_at", `"e"."created_at"`},
		{"category_title", `"c"."title"`},
		{"category_id", `"f"."category_id"`},
		{"title", `"e"."title"`},
		{"author", `"e"."author"`},
	} {
		for _, direction := range []string{"asc", "DESC"} {
			t.Run(test.column+"/"+direction, func(t *testing.T) {
				builder := new(Storage).NewEntryQueryBuilder(1).WithSorting(test.column, direction)
				for _, idsOnly := range []bool{true, false} {
					got, err := builder.buildSorting(idsOnly)
					if err != nil {
						t.Fatal(err)
					}
					column := test.qualifiedColumn
					if !idsOnly {
						column = `"` + test.column + `"`
					}
					want := " ORDER BY " + column + " " + strings.ToUpper(direction)
					if got != want {
						t.Fatalf("IDs only=%t: got %q, want %q", idsOnly, got, want)
					}
				}
			})
		}
	}
}

func TestEntryQuerySortingDirections(t *testing.T) {
	for _, test := range []struct {
		name          string
		direction     string
		wantDirection string
	}{
		{"ascending", "ASC", "ASC"},
		{"lowercase ascending", "asc", "ASC"},
		{"mixed-case ascending", "aSc", "ASC"},
		{"descending", "DESC", "DESC"},
		{"lowercase descending", "desc", "DESC"},
		{"mixed-case descending", "dEsC", "DESC"},
		{"Unicode ascending", "aſc", "ASC"},
		{"Unicode descending", "deſc", "DESC"},
		{"empty", "", ""},
		{"unsupported", "invalid", ""},
		{"leading whitespace", " asc", ""},
		{"trailing whitespace", "desc ", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			builder := new(Storage).NewEntryQueryBuilder(1).WithSorting("id", "asc")
			if builder.WithSorting("title", test.direction) != builder {
				t.Fatal("Expected WithSorting to return the same builder")
			}

			got, err := builder.buildSorting(false)
			if err != nil {
				t.Fatal(err)
			}
			want := ` ORDER BY "id" ASC`
			if test.wantDirection != "" {
				want += `, "title" ` + test.wantDirection
			}
			if got != want {
				t.Fatalf("Got %q, want %q", got, want)
			}
		})
	}
}

func TestEntryQueryIDSortingRejectsUnknownColumns(t *testing.T) {
	for _, column := range []string{"", "unknown", "feed_title", "e.title", `title" DESC; DROP TABLE entries; --`, "title\x00"} {
		t.Run(column, func(t *testing.T) {
			builder := new(Storage).NewEntryQueryBuilder(1).WithSorting(column, "asc")
			if _, err := builder.buildSorting(true); err == nil {
				t.Fatal("Expected an error for an unsupported ID sorting column")
			}
		})
	}
}

func TestEntryQuerySortingPreservesSearchPriority(t *testing.T) {
	for _, direction := range []string{"asc", "desc"} {
		t.Run(direction, func(t *testing.T) {
			builder := new(Storage).NewEntryQueryBuilder(1).
				WithSorting("title", direction).
				WithSearchQuery("shared").
				WithSorting("id", direction).
				WithLimit(25).
				WithOffset(50)
			rank := "ts_rank(document_vectors, websearch_to_tsquery($2)) - extract (epoch from now() - published_at)::float * 0.0000001 DESC"
			sqlDirection := strings.ToUpper(direction)
			for _, idsOnly := range []bool{false, true, false} {
				got, err := builder.buildSorting(idsOnly)
				if err != nil {
					t.Fatal(err)
				}
				prefix := ""
				if idsOnly {
					prefix = `"e".`
				}
				want := " ORDER BY " + strings.Join([]string{
					prefix + `"title" ` + sqlDirection,
					rank,
					prefix + `"id" ` + sqlDirection,
				}, ", ") + " LIMIT 25 OFFSET 50"
				if got != want {
					t.Fatalf("IDs only=%t: got %q, want %q", idsOnly, got, want)
				}
			}
			if len(builder.args) != 2 || builder.args[1] != "shared" {
				t.Fatalf("Unexpected search arguments: %v", builder.args)
			}
		})
	}
}

func TestEntryQuerySortingCompatibility(t *testing.T) {
	for _, test := range []struct {
		name      string
		configure func(*EntryQueryBuilder) *EntryQueryBuilder
		want      string
	}{
		{"no implicit sorting", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b }, ""},
		{"empty search", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithSearchQuery("") }, ""},
		{"invalid direction", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithSorting("id", "invalid") }, ""},
		{"negative offset", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithOffset(-1) }, ""},
		{"IDs pagination", func(b *EntryQueryBuilder) *EntryQueryBuilder {
			return b.WithLimitAndMaximum(0, model.MaxEntryIDsLimit).WithOffset(1)
		}, " LIMIT 10000 OFFSET 1"},
		{"IDs maximum", func(b *EntryQueryBuilder) *EntryQueryBuilder {
			return b.WithLimitAndMaximum(10001, model.MaxEntryIDsLimit)
		}, " LIMIT 10000"},
		{"entries maximum", func(b *EntryQueryBuilder) *EntryQueryBuilder { return b.WithLimit(10001) }, " LIMIT 1000"},
	} {
		t.Run(test.name, func(t *testing.T) {
			builder := test.configure(new(Storage).NewEntryQueryBuilder(1))
			for _, idsOnly := range []bool{false, true} {
				got, err := builder.buildSorting(idsOnly)
				if err != nil {
					t.Fatal(err)
				}
				if got != test.want {
					t.Fatalf("IDs only=%t: got %q, want %q", idsOnly, got, test.want)
				}
			}
		})
	}
}

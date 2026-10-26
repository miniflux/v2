// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestGetEntriesStablePaginationWithIdenticalPublishedDates(t *testing.T) {
	t.Parallel()

	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	const (
		entryCount = 120
		pageSize   = 25
	)

	var feed strings.Builder
	feed.WriteString(`<?xml version="1.0"?><rss version="2.0"><channel><title>Pagination</title><link>https://example.org</link><description>Pagination test</description>`)
	for i := range entryCount {
		fmt.Fprintf(&feed, `<item><guid>pagination-%d</guid><title>Shared pagination title</title><link>https://example.org/%d</link><pubDate>Wed, 01 Jan 2025 00:00:00 GMT</pubDate><description>Shared pagination content</description></item>`, i, i)
	}
	feed.WriteString(`</channel></rss>`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, feed.String())
	}))
	defer server.Close()

	adminClient := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	user, err := adminClient.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	defer adminClient.DeleteUser(user.ID)

	client := miniflux.NewClient(testConfig.testBaseURL, user.Username, testConfig.testRegularPassword)
	if _, err := client.CreateFeed(&miniflux.FeedCreationRequest{FeedURL: server.URL}); err != nil {
		t.Fatal(err)
	}

	all, err := client.EntryIDs(nil)
	if err != nil {
		t.Fatal(err)
	}
	if all.Total != entryCount || len(all.EntryIDs) != entryCount {
		t.Fatalf("Got %d IDs and total %d, want %d", len(all.EntryIDs), all.Total, entryCount)
	}

	for _, direction := range []string{"asc", "desc"} {
		want := slices.Clone(all.EntryIDs)
		slices.Sort(want)
		if direction == "desc" {
			slices.Reverse(want)
		}

		for _, search := range []string{"", "pagination"} {
			t.Run(direction+"/search="+search, func(t *testing.T) {
				var got []int64
				for offset := 0; offset < len(want); offset += pageSize {
					page, err := client.Entries(&miniflux.Filter{
						Order:     "published_at",
						Direction: direction,
						Search:    search,
						Limit:     pageSize,
						Offset:    offset,
					})
					if err != nil {
						t.Fatal(err)
					}
					if page.Total != len(want) {
						t.Fatalf("Got total %d, want %d", page.Total, len(want))
					}
					for _, entry := range page.Entries {
						got = append(got, entry.ID)
					}
				}
				if !slices.Equal(got, want) {
					t.Fatalf("Got IDs %v, want %v", got, want)
				}
			})
		}
	}
}

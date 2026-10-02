// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestFeedLastSuccessfulRefresh(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}
	admin := miniflux.NewClient(testConfig.testBaseURL, testConfig.testAdminUsername, testConfig.testAdminPassword)
	user, err := admin.CreateUser(testConfig.genRandomUsername(), testConfig.testRegularPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.DeleteUser(user.ID) })
	client := miniflux.NewClient(testConfig.testBaseURL, user.Username, testConfig.testRegularPassword)
	category, err := client.CreateCategory("Refresh tracking")
	if err != nil {
		t.Fatal(err)
	}
	var status atomic.Int32
	status.Store(http.StatusOK)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/feed" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("ETag", `"feed-v1"`)
		w.WriteHeader(int(status.Load()))
		if status.Load() == http.StatusOK {
			fmt.Fprint(w, `<rss version="2.0"><channel><title>Refresh test</title><link>http://example.com</link><description>Test</description></channel></rss>`)
		}
	}))
	defer server.Close()
	id, err := client.CreateFeed(&miniflux.FeedCreationRequest{FeedURL: server.URL + "/feed", CategoryID: category.ID})
	if err != nil {
		t.Fatal(err)
	}
	read := func() *miniflux.Feed {
		t.Helper()
		feed, err := client.Feed(id)
		if err != nil {
			t.Fatal(err)
		}
		if feed.LastSuccessfulRefreshAt == nil {
			t.Fatal("successful refresh time is missing")
		}
		return feed
	}
	previous := *read().LastSuccessfulRefreshAt
	for _, code := range []int{http.StatusServiceUnavailable, http.StatusNotModified, http.StatusOK} {
		status.Store(int32(code))
		err := client.RefreshFeed(id)
		if (err != nil) != (code == http.StatusServiceUnavailable) {
			t.Fatalf("refresh status %d: unexpected error %v", code, err)
		}
		feed := read()
		if code == http.StatusServiceUnavailable {
			if !feed.LastSuccessfulRefreshAt.Equal(previous) {
				t.Fatal("failed refresh changed the last successful time")
			}
		} else if !feed.LastSuccessfulRefreshAt.After(previous) || !feed.LastSuccessfulRefreshAt.Equal(feed.CheckedAt) {
			t.Fatal("successful refresh did not advance the last successful time")
		}
		previous = *feed.LastSuccessfulRefreshAt
	}
	newTitle := "Renamed feed"
	if _, err := client.UpdateFeed(id, &miniflux.FeedModificationRequest{Title: &newTitle}); err != nil {
		t.Fatal(err)
	}
	if !read().LastSuccessfulRefreshAt.Equal(previous) {
		t.Fatal("editing feed settings changed the last successful time")
	}
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader_test

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

func TestTagListEndpoint(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	createCategory(t, user, "News")

	var result tagListResult
	user.reader.getJSON(t, "/reader/api/0/tag/list", nil, &result)

	labelPrefix := fmt.Sprintf("user/%d/label/", user.id)
	expected := []tagResult{
		{ID: fmt.Sprintf("user/%d/state/com.google/starred", user.id)},
		{ID: labelPrefix + "All", Label: "All", Type: "folder"},
		{ID: labelPrefix + "News", Label: "News", Type: "folder"},
	}
	if !slices.Equal(result.Tags, expected) {
		t.Errorf(`Unexpected tags, got %+v instead of %+v`, result.Tags, expected)
	}
}

func TestRenameTagEndpoint(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name   string
		source string
	}{
		{"exact title", labelStream("News")},
		{"title in another case", labelStream("news")},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			user := newTestUser(t)
			news := createCategory(t, user, "News")

			user.reader.postOK(t, "/reader/api/0/rename-tag", url.Values{
				"s":    {scenario.source},
				"dest": {labelStream("World")},
			})

			categories, err := user.rest.Categories()
			if err != nil {
				t.Fatal(err)
			}
			index := slices.IndexFunc(categories, func(category *miniflux.Category) bool { return category.ID == news.ID })
			if index == -1 || categories[index].Title != "World" {
				t.Errorf(`Expected category %d to be renamed "World", got %+v`, news.ID, categories)
			}
		})
	}
}

func TestRenameTagRejectsInvalidRequests(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	createCategory(t, user, "News")

	scenarios := []struct {
		name           string
		form           url.Values
		expectedStatus int
	}{
		{"unknown label", url.Values{"s": {labelStream("Unknown")}, "dest": {labelStream("World")}}, http.StatusNotFound},
		{"invalid source", url.Values{"s": {"invalid"}, "dest": {labelStream("World")}}, http.StatusBadRequest},
		{"source is not a label", url.Values{"s": {feedStream(1)}, "dest": {labelStream("World")}}, http.StatusBadRequest},
		{"missing destination", url.Values{"s": {labelStream("News")}}, http.StatusBadRequest},
		{"destination is not a label", url.Values{"s": {labelStream("News")}, "dest": {starredStream}}, http.StatusBadRequest},
		{"empty destination", url.Values{"s": {labelStream("News")}, "dest": {labelStream("")}}, http.StatusBadRequest},
		{"existing destination", url.Values{"s": {labelStream("News")}, "dest": {labelStream("All")}}, http.StatusBadRequest},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			status, body := user.reader.post(t, "/reader/api/0/rename-tag", scenario.form)
			requireStatus(t, status, body, scenario.expectedStatus)
		})
	}

	if titles := categoryTitles(t, user); !slices.Equal(titles, []string{"All", "News"}) {
		t.Errorf(`Expected the categories to be unchanged, got %v`, titles)
	}
}

func TestDisableTagEndpoint(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)
	news := createCategory(t, user, "News")
	createCategory(t, user, "Tech")
	createCategory(t, user, "World")

	feedID, err := user.rest.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL:    servertest.NewFeedFixture(t).FeedURL,
		CategoryID: news.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	user.reader.postOK(t, "/reader/api/0/disable-tag", url.Values{"s": {labelStream("News"), labelStream("Tech")}})

	if titles := categoryTitles(t, user); !slices.Equal(titles, []string{"All", "World"}) {
		t.Errorf(`Expected the remaining categories to be All and World, got %v`, titles)
	}

	// The feeds of the removed categories move to the first remaining category by title.
	feed, err := user.rest.Feed(feedID)
	if err != nil {
		t.Fatal(err)
	}
	if feed.Category.Title != "All" {
		t.Errorf(`Expected the feed to move to "All", got %q`, feed.Category.Title)
	}
}

func TestDisableTagRejectsInvalidRequests(t *testing.T) {
	t.Parallel()

	user := newTestUser(t)

	scenarios := []struct {
		name           string
		streams        []string
		expectedStatus int
	}{
		{"missing stream", nil, http.StatusBadRequest},
		{"invalid stream", []string{"invalid"}, http.StatusBadRequest},
		{"not a label", []string{feedStream(1)}, http.StatusBadRequest},
		{"a label and a state", []string{labelStream("All"), starredStream}, http.StatusBadRequest},
		{"last category", []string{labelStream("All")}, http.StatusBadRequest},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			status, body := user.reader.post(t, "/reader/api/0/disable-tag", url.Values{"s": scenario.streams})
			requireStatus(t, status, body, scenario.expectedStatus)
		})
	}

	if titles := categoryTitles(t, user); !slices.Equal(titles, []string{"All"}) {
		t.Errorf(`Expected the categories to be unchanged, got %v`, titles)
	}
}

func createCategory(t *testing.T, user *testUser, title string) *miniflux.Category {
	t.Helper()

	category, err := user.rest.CreateCategory(title)
	if err != nil {
		t.Fatal(err)
	}
	return category
}

// categoryTitles returns the category titles of the user in alphabetical order, as seen by the REST API.
func categoryTitles(t *testing.T, user *testUser) []string {
	t.Helper()

	categories, err := user.rest.Categories()
	if err != nil {
		t.Fatal(err)
	}

	titles := make([]string, 0, len(categories))
	for _, category := range categories {
		titles = append(titles, category.Title)
	}
	return titles
}

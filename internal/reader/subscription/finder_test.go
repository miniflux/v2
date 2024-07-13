// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package subscription

import (
	"strings"
	"testing"
)

func TestFindYoutubePlaylistFeed(t *testing.T) {
	type testResult struct {
		websiteURL     string
		feedURL        string
		discoveryError bool
	}

	scenarios := []testResult{
		// Video URL
		{
			websiteURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			feedURL:    "",
		},
		// Video URL with position argument
		{
			websiteURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=1",
			feedURL:    "",
		},
		// Video URL with position argument
		{
			websiteURL: "https://www.youtube.com/watch?t=1&v=dQw4w9WgXcQ",
			feedURL:    "",
		},
		// Channel URL
		{
			websiteURL: "https://www.youtube.com/channel/UC-Qj80avWItNRjkZ41rzHyw",
			feedURL:    "",
		},
		// Channel URL with name
		{
			websiteURL: "https://www.youtube.com/@ABCDEFG",
			feedURL:    "",
		},
		// Playlist URL
		{
			websiteURL: "https://www.youtube.com/playlist?list=PLOOwEPgFWm_NHcQd9aCi5JXWASHO_n5uR",
			feedURL:    "https://www.youtube.com/feeds/videos.xml?playlist_id=PLOOwEPgFWm_NHcQd9aCi5JXWASHO_n5uR",
		},
		// Playlist URL with video ID
		{
			websiteURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PLOOwEPgFWm_N42HlCLhqyJ0ZBWr5K1QDM",
			feedURL:    "https://www.youtube.com/feeds/videos.xml?playlist_id=PLOOwEPgFWm_N42HlCLhqyJ0ZBWr5K1QDM",
		},
		// Playlist URL with video ID and index argument
		{
			websiteURL: "https://www.youtube.com/watch?v=6IutBmRJNLk&list=PLOOwEPgFWm_NHcQd9aCi5JXWASHO_n5uR&index=4",
			feedURL:    "https://www.youtube.com/feeds/videos.xml?playlist_id=PLOOwEPgFWm_NHcQd9aCi5JXWASHO_n5uR",
		},
		// Non-Youtube URL
		{
			websiteURL: "https://www.example.com/channel/UC-Qj80avWItNRjkZ41rzHyw",
			feedURL:    "",
		},
		// Invalid URL
		{
			websiteURL:     "https://example|org/",
			feedURL:        "",
			discoveryError: true,
		},
	}

	for _, scenario := range scenarios {
		subscriptions, localizedError := NewSubscriptionFinder(nil).FindSubscriptionsFromYouTubePlaylistPage(scenario.websiteURL)
		if scenario.discoveryError {
			if localizedError == nil {
				t.Fatalf(`Parsing an invalid URL should return an error`)
			}
		}

		if scenario.feedURL == "" {
			if len(subscriptions) > 0 {
				t.Fatalf(`Parsing a non-playlist URL should not return any subscription: %q`, scenario.websiteURL)
			}
		} else {
			if localizedError != nil {
				t.Fatalf(`Parsing a correctly formatted YouTube playlist page should not return any error: %v`, localizedError)
			}

			if len(subscriptions) != 1 {
				t.Fatalf(`Incorrect number of subscriptions returned`)
			}

			if subscriptions[0].URL != scenario.feedURL {
				t.Errorf(`Unexpected Feed, got %s, instead of %s`, subscriptions[0].URL, scenario.feedURL)
			}
		}
	}
}

func TestFindYoutubeChannelFeed(t *testing.T) {
	type testResult struct {
		websiteURL     string
		feedURL        string
		discoveryError bool
	}

	scenarios := []testResult{
		// Video URL
		{
			websiteURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			feedURL:    "",
		},
		// Video URL with position argument
		{
			websiteURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=1",
			feedURL:    "",
		},
		// Video URL with position argument
		{
			websiteURL: "https://www.youtube.com/watch?t=1&v=dQw4w9WgXcQ",
			feedURL:    "",
		},
		// Channel URL
		{
			websiteURL: "https://www.youtube.com/channel/UC-Qj80avWItNRjkZ41rzHyw",
			feedURL:    "https://www.youtube.com/feeds/videos.xml?channel_id=UC-Qj80avWItNRjkZ41rzHyw",
		},
		// Channel URL with name
		{
			websiteURL: "https://www.youtube.com/@ABCDEFG",
			feedURL:    "",
		},
		// Playlist URL
		{
			websiteURL: "https://www.youtube.com/playlist?list=PLOOwEPgFWm_NHcQd9aCi5JXWASHO_n5uR",
			feedURL:    "",
		},
		// Playlist URL with video ID
		{
			websiteURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PLOOwEPgFWm_N42HlCLhqyJ0ZBWr5K1QDM",
			feedURL:    "",
		},
		// Playlist URL with video ID and index argument
		{
			websiteURL: "https://www.youtube.com/watch?v=6IutBmRJNLk&list=PLOOwEPgFWm_NHcQd9aCi5JXWASHO_n5uR&index=4",
			feedURL:    "",
		},
		// Non-Youtube URL
		{
			websiteURL: "https://www.example.com/channel/UC-Qj80avWItNRjkZ41rzHyw",
			feedURL:    "",
		},
		// Invalid URL
		{
			websiteURL:     "https://example|org/",
			feedURL:        "",
			discoveryError: true,
		},
	}

	for _, scenario := range scenarios {
		subscriptions, localizedError := NewSubscriptionFinder(nil).FindSubscriptionsFromYouTubeChannelPage(scenario.websiteURL)
		if scenario.discoveryError {
			if localizedError == nil {
				t.Fatalf(`Parsing an invalid URL should return an error`)
			}
		}

		if scenario.feedURL == "" {
			if len(subscriptions) > 0 {
				t.Fatalf(`Parsing a non-channel URL should not return any subscription: %q`, scenario.websiteURL)
			}
		} else {
			if localizedError != nil {
				t.Fatalf(`Parsing a correctly formatted YouTube channel page should not return any error: %v`, localizedError)
			}

			if len(subscriptions) != 1 {
				t.Fatalf(`Incorrect number of subscriptions returned`)
			}

			if subscriptions[0].URL != scenario.feedURL {
				t.Errorf(`Unexpected Feed, got %s, instead of %s`, subscriptions[0].URL, scenario.feedURL)
			}
		}
	}
}

func TestParseWebPageWithRssFeed(t *testing.T) {
	htmlPage := `
	<!doctype html>
	<html>
		<head>
			<link href="http://example.org/rss" rel="alternate" type="application/rss+xml" title="Some Title">
		</head>
		<body>
		</body>
	</html>`

	subscriptions, err := NewSubscriptionFinder(nil).FindSubscriptionsFromWebPage("http://example.org/", "text/html", strings.NewReader(htmlPage))
	if err != nil {
		t.Fatalf(`Parsing a correctly formatted HTML page should not return any error: %v`, err)
	}

	if len(subscriptions) != 1 {
		t.Fatal(`Incorrect number of subscriptions returned`)
	}

	if subscriptions[0].Title != "Some Title" {
		t.Errorf(`Incorrect subscription title: %q`, subscriptions[0].Title)
	}

	if subscriptions[0].URL != "http://example.org/rss" {
		t.Errorf(`Incorrect subscription URL: %q`, subscriptions[0].URL)
	}

	if subscriptions[0].Type != "rss" {
		t.Errorf(`Incorrect subscription type: %q`, subscriptions[0].Type)
	}
}

func TestParseWebPageWithAtomFeed(t *testing.T) {
	htmlPage := `
	<!doctype html>
	<html>
		<head>
			<link href="http://example.org/atom.xml" rel="alternate" type="application/atom+xml" title="Some Title">
		</head>
		<body>
		</body>
	</html>`

	subscriptions, err := NewSubscriptionFinder(nil).FindSubscriptionsFromWebPage("http://example.org/", "text/html", strings.NewReader(htmlPage))
	if err != nil {
		t.Fatalf(`Parsing a correctly formatted HTML page should not return any error: %v`, err)
	}

	if len(subscriptions) != 1 {
		t.Fatal(`Incorrect number of subscriptions returned`)
	}

	if subscriptions[0].Title != "Some Title" {
		t.Errorf(`Incorrect subscription title: %q`, subscriptions[0].Title)
	}

	if subscriptions[0].URL != "http://example.org/atom.xml" {
		t.Errorf(`Incorrect subscription URL: %q`, subscriptions[0].URL)
	}

	if subscriptions[0].Type != "atom" {
		t.Errorf(`Incorrect subscription type: %q`, subscriptions[0].Type)
	}
}

func TestParseWebPageWithJSONFeed(t *testing.T) {
	htmlPage := `
	<!doctype html>
	<html>
		<head>
			<link href="http://example.org/feed.json" rel="alternate" type="application/feed+json" title="Some Title">
		</head>
		<body>
		</body>
	</html>`

	subscriptions, err := NewSubscriptionFinder(nil).FindSubscriptionsFromWebPage("http://example.org/", "text/html", strings.NewReader(htmlPage))
	if err != nil {
		t.Fatalf(`Parsing a correctly formatted HTML page should not return any error: %v`, err)
	}

	if len(subscriptions) != 1 {
		t.Fatal(`Incorrect number of subscriptions returned`)
	}

	if subscriptions[0].Title != "Some Title" {
		t.Errorf(`Incorrect subscription title: %q`, subscriptions[0].Title)
	}

	if subscriptions[0].URL != "http://example.org/feed.json" {
		t.Errorf(`Incorrect subscription URL: %q`, subscriptions[0].URL)
	}

	if subscriptions[0].Type != "json" {
		t.Errorf(`Incorrect subscription type: %q`, subscriptions[0].Type)
	}
}

func TestParseWebPageWithOldJSONFeedMimeType(t *testing.T) {
	htmlPage := `
	<!doctype html>
	<html>
		<head>
			<link href="http://example.org/feed.json" rel="alternate" type="application/json" title="Some Title">
		</head>
		<body>
		</body>
	</html>`

	subscriptions, err := NewSubscriptionFinder(nil).FindSubscriptionsFromWebPage("http://example.org/", "text/html", strings.NewReader(htmlPage))
	if err != nil {
		t.Fatalf(`Parsing a correctly formatted HTML page should not return any error: %v`, err)
	}

	if len(subscriptions) != 1 {
		t.Fatal(`Incorrect number of subscriptions returned`)
	}

	if subscriptions[0].Title != "Some Title" {
		t.Errorf(`Incorrect subscription title: %q`, subscriptions[0].Title)
	}

	if subscriptions[0].URL != "http://example.org/feed.json" {
		t.Errorf(`Incorrect subscription URL: %q`, subscriptions[0].URL)
	}

	if subscriptions[0].Type != "json" {
		t.Errorf(`Incorrect subscription type: %q`, subscriptions[0].Type)
	}
}

func TestParseWebPageWithRelativeFeedURL(t *testing.T) {
	htmlPage := `
	<!doctype html>
	<html>
		<head>
			<link href="/feed.json" rel="alternate" type="application/feed+json" title="Some Title">
		</head>
		<body>
		</body>
	</html>`

	subscriptions, err := NewSubscriptionFinder(nil).FindSubscriptionsFromWebPage("http://example.org/", "text/html", strings.NewReader(htmlPage))
	if err != nil {
		t.Fatalf(`Parsing a correctly formatted HTML page should not return any error: %v`, err)
	}

	if len(subscriptions) != 1 {
		t.Fatal(`Incorrect number of subscriptions returned`)
	}

	if subscriptions[0].Title != "Some Title" {
		t.Errorf(`Incorrect subscription title: %q`, subscriptions[0].Title)
	}

	if subscriptions[0].URL != "http://example.org/feed.json" {
		t.Errorf(`Incorrect subscription URL: %q`, subscriptions[0].URL)
	}

	if subscriptions[0].Type != "json" {
		t.Errorf(`Incorrect subscription type: %q`, subscriptions[0].Type)
	}
}

func TestParseWebPageWithEmptyTitle(t *testing.T) {
	htmlPage := `
	<!doctype html>
	<html>
		<head>
			<link href="/feed.json" rel="alternate" type="application/feed+json">
		</head>
		<body>
		</body>
	</html>`

	subscriptions, err := NewSubscriptionFinder(nil).FindSubscriptionsFromWebPage("http://example.org/", "text/html", strings.NewReader(htmlPage))
	if err != nil {
		t.Fatalf(`Parsing a correctly formatted HTML page should not return any error: %v`, err)
	}

	if len(subscriptions) != 1 {
		t.Fatal(`Incorrect number of subscriptions returned`)
	}

	if subscriptions[0].Title != "http://example.org/feed.json" {
		t.Errorf(`Incorrect subscription title: %q`, subscriptions[0].Title)
	}

	if subscriptions[0].URL != "http://example.org/feed.json" {
		t.Errorf(`Incorrect subscription URL: %q`, subscriptions[0].URL)
	}

	if subscriptions[0].Type != "json" {
		t.Errorf(`Incorrect subscription type: %q`, subscriptions[0].Type)
	}
}

func TestParseWebPageWithMultipleFeeds(t *testing.T) {
	htmlPage := `
	<!doctype html>
	<html>
		<head>
			<link href="http://example.org/atom.xml" rel="alternate" type="application/atom+xml" title="Atom Feed">
			<link href="http://example.org/feed.json" rel="alternate" type="application/feed+json" title="JSON Feed">
		</head>
		<body>
		</body>
	</html>`

	subscriptions, err := NewSubscriptionFinder(nil).FindSubscriptionsFromWebPage("http://example.org/", "text/html", strings.NewReader(htmlPage))
	if err != nil {
		t.Fatalf(`Parsing a correctly formatted HTML page should not return any error: %v`, err)
	}

	if len(subscriptions) != 2 {
		t.Fatal(`Incorrect number of subscriptions returned`)
	}
}

func TestParseWebPageWithDuplicatedFeeds(t *testing.T) {
	htmlPage := `
	<!doctype html>
	<html>
		<head>
			<link href="http://example.org/feed.xml" rel="alternate" type="application/rss+xml" title="Feed A">
			<link href="http://example.org/feed.xml" rel="alternate" type="application/rss+xml" title="Feed B">
		</head>
		<body>
		</body>
	</html>`

	subscriptions, err := NewSubscriptionFinder(nil).FindSubscriptionsFromWebPage("http://example.org/", "text/html", strings.NewReader(htmlPage))
	if err != nil {
		t.Fatalf(`Parsing a correctly formatted HTML page should not return any error: %v`, err)
	}

	if len(subscriptions) != 1 {
		t.Fatal(`Incorrect number of subscriptions returned`)
	}

	if subscriptions[0].Title != "Feed A" {
		t.Errorf(`Incorrect subscription title: %q`, subscriptions[0].Title)
	}

	if subscriptions[0].URL != "http://example.org/feed.xml" {
		t.Errorf(`Incorrect subscription URL: %q`, subscriptions[0].URL)
	}

	if subscriptions[0].Type != "rss" {
		t.Errorf(`Incorrect subscription type: %q`, subscriptions[0].Type)
	}
}

func TestParseWebPageWithEmptyFeedURL(t *testing.T) {
	htmlPage := `
	<!doctype html>
	<html>
		<head>
			<link href rel="alternate" type="application/feed+json" title="Some Title">
		</head>
		<body>
		</body>
	</html>`

	subscriptions, err := NewSubscriptionFinder(nil).FindSubscriptionsFromWebPage("http://example.org/", "text/html", strings.NewReader(htmlPage))
	if err != nil {
		t.Fatalf(`Parsing a correctly formatted HTML page should not return any error: %v`, err)
	}

	if len(subscriptions) != 0 {
		t.Fatal(`Incorrect number of subscriptions returned`)
	}
}

func TestParseWebPageWithNoHref(t *testing.T) {
	htmlPage := `
	<!doctype html>
	<html>
		<head>
			<link rel="alternate" type="application/feed+json" title="Some Title">
		</head>
		<body>
		</body>
	</html>`

	subscriptions, err := NewSubscriptionFinder(nil).FindSubscriptionsFromWebPage("http://example.org/", "text/html", strings.NewReader(htmlPage))
	if err != nil {
		t.Fatalf(`Parsing a correctly formatted HTML page should not return any error: %v`, err)
	}

	if len(subscriptions) != 0 {
		t.Fatal(`Incorrect number of subscriptions returned`)
	}
}

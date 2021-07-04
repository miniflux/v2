// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package json // import "miniflux.app/reader/json"

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestParseJsonFeed(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"title": "My Example Feed",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"items": [
			{
				"id": "2",
				"content_text": "This is a second item.",
				"url": "https://example.org/second-item"
			},
			{
				"id": "1",
				"content_html": "<p>Hello, world!</p>",
				"url": "https://example.org/initial-post"
			}
		]
	}`

	feed, err := Parse("https://example.org/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "My Example Feed" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "https://example.org/feed.json" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "https://example.org/" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if len(feed.Entries) != 2 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Hash != "d4735e3a265e16eee03f59718b9b5d03019c07d8b6c51f90da3a666eec13ab35" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[0].Hash)
	}

	if feed.Entries[0].URL != "https://example.org/second-item" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].Title != "This is a second item." {
		t.Errorf(`Incorrect entry title, got: "%s"`, feed.Entries[0].Title)
	}

	if feed.Entries[0].Content != "This is a second item." {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[0].Content)
	}

	if feed.Entries[1].Hash != "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[1].Hash)
	}

	if feed.Entries[1].URL != "https://example.org/initial-post" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[1].URL)
	}

	if feed.Entries[1].Title != "https://example.org/initial-post" {
		t.Errorf(`Incorrect entry title, got: "%s"`, feed.Entries[1].Title)
	}

	if feed.Entries[1].Content != "<p>Hello, world!</p>" {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[1].Content)
	}
}

func TestParsePodcast(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"user_comment": "This is a podcast feed. You can add this feed to your podcast client using the following URL: http://therecord.co/feed.json",
		"title": "The Record",
		"home_page_url": "http://therecord.co/",
		"feed_url": "http://therecord.co/feed.json",
		"items": [
			{
				"id": "http://therecord.co/chris-parrish",
				"title": "Special #1 - Chris Parrish",
				"url": "http://therecord.co/chris-parrish",
				"content_text": "Chris has worked at Adobe and as a founder of Rogue Sheep, which won an Apple Design Award for Postage. Chris’s new company is Aged & Distilled with Guy English — which shipped Napkin, a Mac app for visual collaboration. Chris is also the co-host of The Record. He lives on Bainbridge Island, a quick ferry ride from Seattle.",
				"content_html": "Chris has worked at <a href=\"http://adobe.com/\">Adobe</a> and as a founder of Rogue Sheep, which won an Apple Design Award for Postage. Chris’s new company is Aged & Distilled with Guy English — which shipped <a href=\"http://aged-and-distilled.com/napkin/\">Napkin</a>, a Mac app for visual collaboration. Chris is also the co-host of The Record. He lives on <a href=\"http://www.ci.bainbridge-isl.wa.us/\">Bainbridge Island</a>, a quick ferry ride from Seattle.",
				"summary": "Brent interviews Chris Parrish, co-host of The Record and one-half of Aged & Distilled.",
				"date_published": "2014-05-09T14:04:00-07:00",
				"attachments": [
					{
						"url": "http://therecord.co/downloads/The-Record-sp1e1-ChrisParrish.m4a",
						"mime_type": "audio/x-m4a",
						"size_in_bytes": 89970236,
						"duration_in_seconds": 6629
					}
				]
			}
		]
	}`

	feed, err := Parse("http://therecord.co/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "The Record" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "http://therecord.co/feed.json" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "http://therecord.co/" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if len(feed.Entries) != 1 {
		t.Fatalf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Hash != "6b678e57962a1b001e4e873756563cdc08bbd06ca561e764e0baa9a382485797" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[0].Hash)
	}

	if feed.Entries[0].URL != "http://therecord.co/chris-parrish" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].Title != "Special #1 - Chris Parrish" {
		t.Errorf(`Incorrect entry title, got: "%s"`, feed.Entries[0].Title)
	}

	if feed.Entries[0].Content != `Chris has worked at <a href="http://adobe.com/">Adobe</a> and as a founder of Rogue Sheep, which won an Apple Design Award for Postage. Chris’s new company is Aged & Distilled with Guy English — which shipped <a href="http://aged-and-distilled.com/napkin/">Napkin</a>, a Mac app for visual collaboration. Chris is also the co-host of The Record. He lives on <a href="http://www.ci.bainbridge-isl.wa.us/">Bainbridge Island</a>, a quick ferry ride from Seattle.` {
		t.Errorf(`Incorrect entry content, got: "%s"`, feed.Entries[0].Content)
	}

	location, _ := time.LoadLocation("America/Vancouver")
	if !feed.Entries[0].Date.Equal(time.Date(2014, time.May, 9, 14, 4, 0, 0, location)) {
		t.Errorf("Incorrect entry date, got: %v", feed.Entries[0].Date)
	}

	if len(feed.Entries[0].Enclosures) != 1 {
		t.Fatalf("Incorrect number of enclosures, got: %d", len(feed.Entries[0].Enclosures))
	}

	if feed.Entries[0].Enclosures[0].URL != "http://therecord.co/downloads/The-Record-sp1e1-ChrisParrish.m4a" {
		t.Errorf("Incorrect enclosure URL, got: %s", feed.Entries[0].Enclosures[0].URL)
	}

	if feed.Entries[0].Enclosures[0].MimeType != "audio/x-m4a" {
		t.Errorf("Incorrect enclosure type, got: %s", feed.Entries[0].Enclosures[0].MimeType)
	}

	if feed.Entries[0].Enclosures[0].Size != 89970236 {
		t.Errorf("Incorrect enclosure length, got: %d", feed.Entries[0].Enclosures[0].Size)
	}
}

func TestParseEntryWithoutAttachmentURL(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"user_comment": "This is a podcast feed. You can add this feed to your podcast client using the following URL: http://therecord.co/feed.json",
		"title": "The Record",
		"home_page_url": "http://therecord.co/",
		"feed_url": "http://therecord.co/feed.json",
		"items": [
			{
				"id": "http://therecord.co/chris-parrish",
				"title": "Special #1 - Chris Parrish",
				"url": "http://therecord.co/chris-parrish",
				"content_text": "Chris has worked at Adobe and as a founder of Rogue Sheep, which won an Apple Design Award for Postage. Chris’s new company is Aged & Distilled with Guy English — which shipped Napkin, a Mac app for visual collaboration. Chris is also the co-host of The Record. He lives on Bainbridge Island, a quick ferry ride from Seattle.",
				"date_published": "2014-05-09T14:04:00-07:00",
				"attachments": [
					{
						"url": "",
						"mime_type": "audio/x-m4a",
						"size_in_bytes": 0
					}
				]
			}
		]
	}`

	feed, err := Parse("http://therecord.co/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Fatalf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if len(feed.Entries[0].Enclosures) != 0 {
		t.Errorf("Incorrect number of enclosures, got: %d", len(feed.Entries[0].Enclosures))
	}
}

func TestParseFeedWithRelativeURL(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"title": "Example",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"items": [
			{
				"id": "2347259",
				"url": "something.html",
				"date_published": "2016-02-09T14:22:00-07:00"
			}
		]
	}`

	feed, err := Parse("https://example.org/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].URL != "https://example.org/something.html" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}
}

func TestParseAuthor(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"user_comment": "This is a microblog feed. You can add this to your feed reader using the following URL: https://example.org/feed.json",
		"title": "Brent Simmons’s Microblog",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"author": {
			"name": "Brent Simmons",
			"url": "http://example.org/",
			"avatar": "https://example.org/avatar.png"
		},
		"items": [
			{
				"id": "2347259",
				"url": "https://example.org/2347259",
				"content_text": "Cats are neat. \n\nhttps://example.org/cats",
				"date_published": "2016-02-09T14:22:00-07:00"
			}
		]
	}`

	feed, err := Parse("https://example.org/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Author != "Brent Simmons" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseFeedWithoutTitle(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"items": [
			{
				"id": "2347259",
				"url": "https://example.org/2347259",
				"content_text": "Cats are neat. \n\nhttps://example.org/cats",
				"date_published": "2016-02-09T14:22:00-07:00"
			}
		]
	}`

	feed, err := Parse("https://example.org/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "https://example.org/" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}
}

func TestParseFeedItemWithInvalidDate(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"title": "My Example Feed",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"items": [
			{
				"id": "2347259",
				"url": "https://example.org/2347259",
				"content_text": "Cats are neat. \n\nhttps://example.org/cats",
				"date_published": "Tomorrow"
			}
		]
	}`

	feed, err := Parse("https://example.org/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	duration := time.Since(feed.Entries[0].Date)
	if duration.Seconds() > 1 {
		t.Errorf("Incorrect entry date, got: %v", feed.Entries[0].Date)
	}
}

func TestParseFeedItemWithoutID(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"title": "My Example Feed",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"items": [
			{
				"content_text": "Some text."
			}
		]
	}`

	feed, err := Parse("https://example.org/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Hash != "13b4c5aecd1b6d749afcee968fbf9c80f1ed1bbdbe1aaf25cb34ebd01144bbe9" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[0].Hash)
	}
}

func TestParseFeedItemWithoutTitle(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"title": "My Example Feed",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"items": [
			{
				"url": "https://example.org/item"
			}
		]
	}`

	feed, err := Parse("https://example.org/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Title != "https://example.org/item" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseTruncateItemTitle(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"title": "My Example Feed",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"items": [
			{
				"title": "` + strings.Repeat("a", 200) + `"
			}
		]
	}`

	feed, err := Parse("https://example.org/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if len(feed.Entries[0].Title) != 103 {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}

	if len([]rune(feed.Entries[0].Title)) != 101 {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseTruncateItemTitleUnicode(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"title": "My Example Feed",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"items": [
			{
				"title": "I’m riding my electric bike and came across this castle. It’s called “Schloss Richmond”. 🚴‍♂️"
			}
		]
	}`

	feed, err := Parse("https://example.org/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if len(feed.Entries[0].Title) != 110 {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}

	if len([]rune(feed.Entries[0].Title)) != 93 {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseItemTitleWithXMLTags(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"title": "My Example Feed",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"items": [
			{
				"title": "</example>"
			}
		]
	}`

	feed, err := Parse("https://example.org/feed.json", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Title != "</example>" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseInvalidJSON(t *testing.T) {
	data := `garbage`
	_, err := Parse("https://example.org/feed.json", bytes.NewBufferString(data))
	if err == nil {
		t.Error("Parse should returns an error")
	}
}

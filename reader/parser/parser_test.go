// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package parser // import "miniflux.app/reader/parser"

import (
	"bytes"
	"os"
	"testing"

	"miniflux.app/http/client"
)

func TestParseAtom(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">

	  <title>Example Feed</title>
	  <link href="http://example.org/"/>
	  <updated>2003-12-13T18:30:02Z</updated>
	  <author>
		<name>John Doe</name>
	  </author>
	  <id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>

	  <entry>
		<title>Atom-Powered Robots Run Amok</title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := ParseFeed("https://example.org/", data)
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "Example Feed" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}
}

func TestParseAtomFeedWithRelativeURL(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="/blog/atom.xml" rel="self" type="application/atom+xml"/>
	  <link href="/blog"/>

	  <entry>
		<title>Test</title>
		<link href="/blog/article.html"/>
		<link href="/blog/article.html" rel="alternate" type="text/html"/>
		<id>/blog/article.html</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := ParseFeed("https://example.org/blog/atom.xml", data)
	if err != nil {
		t.Fatal(err)
	}

	if feed.FeedURL != "https://example.org/blog/atom.xml" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "https://example.org/blog" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if feed.Entries[0].URL != "https://example.org/blog/article.html" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}
}

func TestParseRSS(t *testing.T) {
	data := `<?xml version="1.0"?>
	<rss version="2.0">
	<channel>
		<title>Liftoff News</title>
		<link>http://liftoff.msfc.nasa.gov/</link>
		<item>
			<title>Star City</title>
			<link>http://liftoff.msfc.nasa.gov/news/2003/news-starcity.asp</link>
			<description>How do Americans get ready to work with Russians aboard the International Space Station? They take a crash course in culture, language and protocol at Russia's &lt;a href="http://howe.iki.rssi.ru/GCTC/gctc_e.htm"&gt;Star City&lt;/a&gt;.</description>
			<pubDate>Tue, 03 Jun 2003 09:39:21 GMT</pubDate>
			<guid>http://liftoff.msfc.nasa.gov/2003/06/03.html#item573</guid>
		</item>
	</channel>
	</rss>`

	feed, err := ParseFeed("http://liftoff.msfc.nasa.gov/", data)
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "Liftoff News" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}
}

func TestParseRSSFeedWithRelativeURL(t *testing.T) {
	data := `<?xml version="1.0"?>
	<rss version="2.0">
	<channel>
		<title>Example Feed</title>
		<link>/blog</link>
		<item>
			<title>Example Entry</title>
			<link>/blog/article.html</link>
			<description>Something</description>
			<pubDate>Tue, 03 Jun 2003 09:39:21 GMT</pubDate>
			<guid>1234</guid>
		</item>
	</channel>
	</rss>`

	feed, err := ParseFeed("http://example.org/rss.xml", data)
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "Example Feed" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "http://example.org/rss.xml" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "http://example.org/blog" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if feed.Entries[0].URL != "http://example.org/blog/article.html" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}
}

func TestParseRDF(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rdf:RDF
		  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		  xmlns="http://purl.org/rss/1.0/"
		>

		  <channel>
			<title>RDF Example</title>
			<link>http://example.org/</link>
		  </channel>

		  <item>
			<title>Title</title>
			<link>http://example.org/item</link>
			<description>Test</description>
		  </item>
		</rdf:RDF>`

	feed, err := ParseFeed("http://example.org/", data)
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "RDF Example" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}
}

func TestParseRDFWithRelativeURL(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rdf:RDF
		  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		  xmlns="http://purl.org/rss/1.0/"
		>

		  <channel>
			<title>RDF Example</title>
			<link>/blog</link>
		  </channel>

		  <item>
			<title>Title</title>
			<link>/blog/article.html</link>
			<description>Test</description>
		  </item>
		</rdf:RDF>`

	feed, err := ParseFeed("http://example.org/rdf.xml", data)
	if err != nil {
		t.Error(err)
	}

	if feed.FeedURL != "http://example.org/rdf.xml" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "http://example.org/blog" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if feed.Entries[0].URL != "http://example.org/blog/article.html" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}
}

func TestParseJson(t *testing.T) {
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

	feed, err := ParseFeed("https://example.org/feed.json", data)
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "My Example Feed" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}
}

func TestParseJsonFeedWithRelativeURL(t *testing.T) {
	data := `{
		"version": "https://jsonfeed.org/version/1",
		"title": "My Example Feed",
		"home_page_url": "/blog",
		"feed_url": "/blog/feed.json",
		"items": [
			{
				"id": "2",
				"content_text": "This is a second item.",
				"url": "/blog/article.html"
			}
		]
	}`

	feed, err := ParseFeed("https://example.org/blog/feed.json", data)
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "My Example Feed" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "https://example.org/blog/feed.json" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "https://example.org/blog" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if feed.Entries[0].URL != "https://example.org/blog/article.html" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}
}

func TestParseUnknownFeed(t *testing.T) {
	data := `
		<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
		<html xmlns="http://www.w3.org/1999/xhtml">
			<head>
				<title>Title of document</title>
			</head>
			<body>
				some content
			</body>
		</html>
	`

	_, err := ParseFeed("https://example.org/", data)
	if err == nil {
		t.Error("ParseFeed must returns an error")
	}
}

func TestParseEmptyFeed(t *testing.T) {
	_, err := ParseFeed("", "")
	if err == nil {
		t.Error("ParseFeed must returns an error")
	}
}

func TestDifferentEncodingWithResponse(t *testing.T) {
	var unicodeTestCases = []struct {
		filename, contentType string
		index                 int
		title                 string
	}{
		// Arabic language encoded in UTF-8.
		{"urdu_UTF8.xml", "text/xml; charset=utf-8", 0, "امریکی عسکری امداد کی بندش کی وجوہات: انڈیا سے جنگ، جوہری پروگرام اور اب دہشت گردوں کی پشت پناہی"},

		// Windows-1251 encoding and not charset in HTTP header.
		{"encoding_WINDOWS-1251.xml", "text/xml", 0, "Цитата #17703"},

		// No encoding in XML, but defined in HTTP Content-Type header.
		{"no_encoding_ISO-8859-1.xml", "application/xml; charset=ISO-8859-1", 2, "La criminalité liée surtout à... l'ennui ?"},

		// ISO-8859-1 encoding defined in XML and HTTP header.
		{"encoding_ISO-8859-1.xml", "application/rss+xml; charset=ISO-8859-1", 5, "Projekt Jedi: Microsoft will weiter mit US-Militär zusammenarbeiten"},

		// UTF-8 encoding defined in RDF document and HTTP header.
		{"rdf_UTF8.xml", "application/rss+xml; charset=utf-8", 1, "Mega-Deal: IBM übernimmt Red Hat"},

		// UTF-8 encoding defined only in RDF document.
		{"rdf_UTF8.xml", "application/rss+xml", 1, "Mega-Deal: IBM übernimmt Red Hat"},
	}

	for _, tc := range unicodeTestCases {
		content, err := os.ReadFile("testdata/" + tc.filename)
		if err != nil {
			t.Fatalf(`Unable to read file %q: %v`, tc.filename, err)
		}

		r := &client.Response{Body: bytes.NewReader(content), ContentType: tc.contentType}
		if encodingErr := r.EnsureUnicodeBody(); encodingErr != nil {
			t.Fatalf(`Encoding error for %q: %v`, tc.filename, encodingErr)
		}

		feed, parseErr := ParseFeed("https://example.org/", r.BodyAsString())
		if parseErr != nil {
			t.Fatalf(`Parsing error for %q - %q: %v`, tc.filename, tc.contentType, parseErr)
		}

		if feed.Entries[tc.index].Title != tc.title {
			t.Errorf(`Unexpected title, got %q instead of %q`, feed.Entries[tc.index].Title, tc.title)
		}
	}
}

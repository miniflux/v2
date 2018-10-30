// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package parser // import "miniflux.app/reader/parser"

import (
	"bytes"
	"io/ioutil"
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

	feed, err := ParseFeed(data)
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "Example Feed" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
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

	feed, err := ParseFeed(data)
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "Liftoff News" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
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

	feed, err := ParseFeed(data)
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "RDF Example" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
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

	feed, err := ParseFeed(data)
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "My Example Feed" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
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

	_, err := ParseFeed(data)
	if err == nil {
		t.Error("ParseFeed must returns an error")
	}
}

func TestParseEmptyFeed(t *testing.T) {
	_, err := ParseFeed("")
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
		content, err := ioutil.ReadFile("testdata/" + tc.filename)
		if err != nil {
			t.Fatalf(`Unable to read file %q: %v`, tc.filename, err)
		}

		r := &client.Response{Body: bytes.NewReader(content), ContentType: tc.contentType}
		r.EnsureUnicodeBody()
		feed, parseErr := ParseFeed(r.String())
		if parseErr != nil {
			t.Errorf(`Parsing error for %q - %q: %v`, tc.filename, tc.contentType, parseErr)
		}

		if feed.Entries[tc.index].Title != tc.title {
			t.Errorf(`Unexpected title, got %q instead of %q`, feed.Entries[tc.index].Title, tc.title)
		}
	}
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package parser // import "miniflux.app/v2/internal/reader/parser"

import (
	"os"
	"strings"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	var testCases = map[string][]string{
		"large_atom.xml": {"https://dustri.org/b", ""},
		"large_rss.xml":  {"https://dustri.org/b", ""},
		"small_atom.xml": {"https://github.com/miniflux/v2/commits/main", ""},
	}
	for filename := range testCases {
		data, err := os.ReadFile("./testdata/" + filename)
		if err != nil {
			b.Fatalf(`Unable to read file %q: %v`, filename, err)
		}
		testCases[filename][1] = string(data)
	}
	for range b.N {
		for _, v := range testCases {
			ParseFeed(v[0], strings.NewReader(v[1]))
		}
	}
}

func FuzzParse(f *testing.F) {
	f.Add("https://z.org", `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
<title>Example Feed</title>
<link href="http://z.org/"/>
<link href="/k"/>
<updated>2003-12-13T18:30:02Z</updated>
<author><name>John Doe</name></author>
<id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>
<entry>
<title>a</title>
<link href="http://example.org/b"/>
<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
<updated>2003-12-13T18:30:02Z</updated>
<summary>c</summary>
</entry>
</feed>`)
	f.Add("https://z.org", `<?xml version="1.0"?>
<rss version="2.0">
<channel>
<title>a</title>
<link>http://z.org</link>
<item>
<title>a</title>
<link>http://z.org</link>
<description>d</description>
<pubDate>Tue, 03 Jun 2003 09:39:21 GMT</pubDate>
<guid>l</guid>
</item>
</channel>
</rss>`)
	f.Add("https://z.org", `<?xml version="1.0" encoding="utf-8"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/">
<channel>
<title>a</title>
<link>http://z.org/</link>
</channel>
<item>
<title>a</title>
<link>/</link>
<description>c</description>
</item>
</rdf:RDF>`)
	f.Add("http://z.org", `{
"version": "http://jsonfeed.org/version/1",
"title": "a",
"home_page_url": "http://z.org/",
"feed_url": "http://z.org/a.json",
"items": [
{"id": "2","content_text": "a","url": "https://z.org/2"},
{"id": "1","content_html": "<a","url":"http://z.org/1"}]}`)
	f.Fuzz(func(t *testing.T, url string, data string) {
		ParseFeed(url, strings.NewReader(data))
	})
}

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

	feed, err := ParseFeed("https://example.org/", strings.NewReader(data))
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

	feed, err := ParseFeed("https://example.org/blog/atom.xml", strings.NewReader(data))
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

	feed, err := ParseFeed("http://liftoff.msfc.nasa.gov/", strings.NewReader(data))
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

	feed, err := ParseFeed("http://example.org/rss.xml", strings.NewReader(data))
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

	feed, err := ParseFeed("http://example.org/", strings.NewReader(data))
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

	feed, err := ParseFeed("http://example.org/rdf.xml", strings.NewReader(data))
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

	feed, err := ParseFeed("https://example.org/feed.json", strings.NewReader(data))
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

	feed, err := ParseFeed("https://example.org/blog/feed.json", strings.NewReader(data))
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

	_, err := ParseFeed("https://example.org/", strings.NewReader(data))
	if err == nil {
		t.Error("ParseFeed must returns an error")
	}
}

func TestParseEmptyFeed(t *testing.T) {
	_, err := ParseFeed("", strings.NewReader(""))
	if err == nil {
		t.Error("ParseFeed must returns an error")
	}
}

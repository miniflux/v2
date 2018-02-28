// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rss

import (
	"bytes"
	"testing"
	"time"
)

func TestParseRss2Sample(t *testing.T) {
	data := `
		<?xml version="1.0"?>
		<rss version="2.0">
		<channel>
			<title>Liftoff News</title>
			<link>http://liftoff.msfc.nasa.gov/</link>
			<description>Liftoff to Space Exploration.</description>
			<language>en-us</language>
			<pubDate>Tue, 10 Jun 2003 04:00:00 GMT</pubDate>
			<lastBuildDate>Tue, 10 Jun 2003 09:41:01 GMT</lastBuildDate>
			<docs>http://blogs.law.harvard.edu/tech/rss</docs>
			<generator>Weblog Editor 2.0</generator>
			<managingEditor>editor@example.com</managingEditor>
			<webMaster>webmaster@example.com</webMaster>
			<item>
				<title>Star City</title>
				<link>http://liftoff.msfc.nasa.gov/news/2003/news-starcity.asp</link>
				<description>How do Americans get ready to work with Russians aboard the International Space Station? They take a crash course in culture, language and protocol at Russia's &lt;a href="http://howe.iki.rssi.ru/GCTC/gctc_e.htm"&gt;Star City&lt;/a&gt;.</description>
				<pubDate>Tue, 03 Jun 2003 09:39:21 GMT</pubDate>
				<guid>http://liftoff.msfc.nasa.gov/2003/06/03.html#item573</guid>
			</item>
			<item>
				<description>Sky watchers in Europe, Asia, and parts of Alaska and Canada will experience a &lt;a href="http://science.nasa.gov/headlines/y2003/30may_solareclipse.htm"&gt;partial eclipse of the Sun&lt;/a&gt; on Saturday, May 31st.</description>
				<pubDate>Fri, 30 May 2003 11:06:42 GMT</pubDate>
				<guid>http://liftoff.msfc.nasa.gov/2003/05/30.html#item572</guid>
			</item>
			<item>
				<title>The Engine That Does More</title>
				<link>http://liftoff.msfc.nasa.gov/news/2003/news-VASIMR.asp</link>
				<description>Before man travels to Mars, NASA hopes to design new engines that will let us fly through the Solar System more quickly.  The proposed VASIMR engine would do that.</description>
				<pubDate>Tue, 27 May 2003 08:37:32 GMT</pubDate>
				<guid>http://liftoff.msfc.nasa.gov/2003/05/27.html#item571</guid>
			</item>
			<item>
				<title>Astronauts' Dirty Laundry</title>
				<link>http://liftoff.msfc.nasa.gov/news/2003/news-laundry.asp</link>
				<description>Compared to earlier spacecraft, the International Space Station has many luxuries, but laundry facilities are not one of them.  Instead, astronauts have other options.</description>
				<pubDate>Tue, 20 May 2003 08:56:02 GMT</pubDate>
				<guid>http://liftoff.msfc.nasa.gov/2003/05/20.html#item570</guid>
			</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "Liftoff News" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "http://liftoff.msfc.nasa.gov/" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if len(feed.Entries) != 4 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	expectedDate := time.Date(2003, time.June, 3, 9, 39, 21, 0, time.UTC)
	if !feed.Entries[0].Date.Equal(expectedDate) {
		t.Errorf("Incorrect entry date, got: %v, want: %v", feed.Entries[0].Date, expectedDate)
	}

	if feed.Entries[0].Hash != "5b2b4ac2fe1786ddf0fd2da2f1b07f64e691264f41f2db3ea360f31bb6d9152b" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[0].Hash)
	}

	if feed.Entries[0].URL != "http://liftoff.msfc.nasa.gov/news/2003/news-starcity.asp" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].Title != "Star City" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}

	if feed.Entries[0].Content != `How do Americans get ready to work with Russians aboard the International Space Station? They take a crash course in culture, language and protocol at Russia's <a href="http://howe.iki.rssi.ru/GCTC/gctc_e.htm">Star City</a>.` {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[0].Content)
	}
}

func TestParseFeedWithoutTitle(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0">
		<channel>
			<link>https://example.org/</link>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "https://example.org/" {
		t.Errorf("Incorrect feed title, got: %s", feed.Title)
	}
}

func TestParseEntryWithoutTitle(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0">
		<channel>
			<link>https://example.org/</link>
			<item>
				<link>https://example.org/item</link>
			</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Title != "https://example.org/item" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseEntryWithoutLink(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0">
		<channel>
			<link>https://example.org/</link>
			<item>
				<guid isPermaLink="false">1234</guid>
			</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].URL != "https://example.org/" {
		t.Errorf("Incorrect entry link, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].Hash != "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[0].Hash)
	}
}

func TestParseEntryWithAtomLink(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
		<channel>
			<link>https://example.org/</link>
			<item>
				<title>Test</title>
				<atom:link href="https://example.org/item" />
			</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].URL != "https://example.org/item" {
		t.Errorf("Incorrect entry link, got: %s", feed.Entries[0].URL)
	}
}

func TestParseEntryWithMultipleAtomLinks(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
		<channel>
			<link>https://example.org/</link>
			<item>
				<title>Test</title>
				<atom:link rel="payment" href="https://example.org/a" />
				<atom:link rel="http://foobar.tld" href="https://example.org/b" />
			</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].URL != "https://example.org/b" {
		t.Errorf("Incorrect entry link, got: %s", feed.Entries[0].URL)
	}
}

func TestParseFeedURLWithAtomLink(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss xmlns:atom="http://www.w3.org/2005/Atom" version="2.0">
		<channel>
			<title>Example</title>
			<link>https://example.org/</link>
			<atom:link href="https://example.org/rss" type="application/rss+xml" rel="self"></atom:link>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.FeedURL != "https://example.org/rss" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "https://example.org/" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}
}

func TestParseEntryWithAtomAuthor(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss xmlns:atom="http://www.w3.org/2005/Atom" version="2.0">
		<channel>
			<title>Example</title>
			<link>https://example.org/</link>
			<atom:link href="https://example.org/rss" type="application/rss+xml" rel="self"></atom:link>
			<item>
				<title>Test</title>
				<link>https://example.org/item</link>
				<author xmlns:author="http://www.w3.org/2005/Atom">
					<name>Foo Bar</name>
					<title>Vice President</title>
					<department/>
					<company>FooBar Inc.</company>
				</author>
			</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Author != "Foo Bar" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseEntryWithDublinCoreAuthor(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/">
		<channel>
			<title>Example</title>
			<link>https://example.org/</link>
			<item>
				<title>Test</title>
				<link>https://example.org/item</link>
				<dc:creator>Me (me@example.com)</dc:creator>
			</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Author != "Me (me@example.com)" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseEntryWithItunesAuthor(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">
		<channel>
			<title>Example</title>
			<link>https://example.org/</link>
			<item>
				<title>Test</title>
				<link>https://example.org/item</link>
				<itunes:author>Someone</itunes:author>
			</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Author != "Someone" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseFeedWithItunesAuthor(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">
		<channel>
			<title>Example</title>
			<link>https://example.org/</link>
			<itunes:author>Someone</itunes:author>
			<item>
				<title>Test</title>
				<link>https://example.org/item</link>
			</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Author != "Someone" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseEntryWithDublinCoreDate(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
				<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/">
				<channel>
					<title>Example</title>
					<link>http://example.org/</link>
					<item>
						<title>Item 1</title>
						<link>http://example.org/item1</link>
						<description>Description.</description>
						<guid isPermaLink="false">UUID</guid>
						<dc:date>2002-09-29T23:40:06-05:00</dc:date>
					</item>
				</channel>
			</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	location, _ := time.LoadLocation("EST")
	expectedDate := time.Date(2002, time.September, 29, 23, 40, 06, 0, location)
	if !feed.Entries[0].Date.Equal(expectedDate) {
		t.Errorf("Incorrect entry date, got: %v, want: %v", feed.Entries[0].Date, expectedDate)
	}
}

func TestParseEntryWithContentEncoded(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/">
		<channel>
			<title>Example</title>
			<link>http://example.org/</link>
			<item>
				<title>Item 1</title>
				<link>http://example.org/item1</link>
				<description>Description.</description>
				<guid isPermaLink="false">UUID</guid>
				<content:encoded><![CDATA[<p><a href="http://www.example.org/">Example</a>.</p>]]></content:encoded>
			</item>
		</channel>
	</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Content != `<p><a href="http://www.example.org/">Example</a>.</p>` {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[0].Content)
	}
}

func TestParseEntryWithFeedBurnerLink(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0" xmlns:feedburner="http://rssnamespace.org/feedburner/ext/1.0">
		<channel>
			<title>Example</title>
			<link>http://example.org/</link>
			<item>
				<title>Item 1</title>
				<link>http://example.org/item1</link>
				<feedburner:origLink>http://example.org/original</feedburner:origLink>
			</item>
		</channel>
	</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].URL != "http://example.org/original" {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[0].URL)
	}
}

func TestParseEntryTitleWithWhitespaces(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rss version="2.0">
	<channel>
		<title>Example</title>
		<link>http://example.org</link>
		<item>
			<title>
				Some Title
			</title>
			<link>http://www.example.org/entries/1</link>
			<pubDate>Fri, 15 Jul 2005 00:00:00 -0500</pubDate>
		</item>
	</channel>
	</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Title != "Some Title" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseEntryWithEnclosures(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0">
		<channel>
		<title>My Podcast Feed</title>
		<link>http://example.org</link>
		<author>some.email@example.org</author>
		<item>
			<title>Podcasting with RSS</title>
			<link>http://www.example.org/entries/1</link>
			<description>An overview of RSS podcasting</description>
			<pubDate>Fri, 15 Jul 2005 00:00:00 -0500</pubDate>
			<guid isPermaLink="true">http://www.example.org/entries/1</guid>
			<enclosure url="http://www.example.org/myaudiofile.mp3"
					length="12345"
					type="audio/mpeg" />
		</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].URL != "http://www.example.org/entries/1" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if len(feed.Entries[0].Enclosures) != 1 {
		t.Errorf("Incorrect number of enclosures, got: %d", len(feed.Entries[0].Enclosures))
	}

	if feed.Entries[0].Enclosures[0].URL != "http://www.example.org/myaudiofile.mp3" {
		t.Errorf("Incorrect enclosure URL, got: %s", feed.Entries[0].Enclosures[0].URL)
	}

	if feed.Entries[0].Enclosures[0].MimeType != "audio/mpeg" {
		t.Errorf("Incorrect enclosure type, got: %s", feed.Entries[0].Enclosures[0].MimeType)
	}

	if feed.Entries[0].Enclosures[0].Size != 12345 {
		t.Errorf("Incorrect enclosure length, got: %d", feed.Entries[0].Enclosures[0].Size)
	}
}

func TestParseEntryWithFeedBurnerEnclosures(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0" xmlns:feedburner="http://rssnamespace.org/feedburner/ext/1.0">
		<channel>
		<title>My Example Feed</title>
		<link>http://example.org</link>
		<author>some.email@example.org</author>
		<item>
			<title>Example Item</title>
			<link>http://www.example.org/entries/1</link>
			<enclosure
				url="http://feedproxy.google.com/~r/example/~5/lpMyFSCvubs/File.mp3"
				length="76192460"
				type="audio/mpeg" />
			<feedburner:origEnclosureLink>http://example.org/67ca416c-f22a-4228-a681-68fc9998ec10/File.mp3</feedburner:origEnclosureLink>
		</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].URL != "http://www.example.org/entries/1" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if len(feed.Entries[0].Enclosures) != 1 {
		t.Errorf("Incorrect number of enclosures, got: %d", len(feed.Entries[0].Enclosures))
	}

	if feed.Entries[0].Enclosures[0].URL != "http://example.org/67ca416c-f22a-4228-a681-68fc9998ec10/File.mp3" {
		t.Errorf("Incorrect enclosure URL, got: %s", feed.Entries[0].Enclosures[0].URL)
	}

	if feed.Entries[0].Enclosures[0].MimeType != "audio/mpeg" {
		t.Errorf("Incorrect enclosure type, got: %s", feed.Entries[0].Enclosures[0].MimeType)
	}

	if feed.Entries[0].Enclosures[0].Size != 76192460 {
		t.Errorf("Incorrect enclosure length, got: %d", feed.Entries[0].Enclosures[0].Size)
	}
}

func TestParseEntryWithRelativeURL(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0">
		<channel>
			<link>https://example.org/</link>
			<item>
				<link>item.html</link>
			</item>
		</channel>
		</rss>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Title != "https://example.org/item.html" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseInvalidXml(t *testing.T) {
	data := `garbage`
	_, err := Parse(bytes.NewBufferString(data))
	if err == nil {
		t.Error("Parse should returns an error")
	}
}

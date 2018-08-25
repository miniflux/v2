// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package atom // import "miniflux.app/reader/atom"

import (
	"bytes"
	"testing"
	"time"
)

func TestParseAtomSample(t *testing.T) {
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

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "Example Feed" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "http://example.org/" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if !feed.Entries[0].Date.Equal(time.Date(2003, time.December, 13, 18, 30, 2, 0, time.UTC)) {
		t.Errorf("Incorrect entry date, got: %v", feed.Entries[0].Date)
	}

	if feed.Entries[0].Hash != "3841e5cf232f5111fc5841e9eba5f4b26d95e7d7124902e0f7272729d65601a6" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[0].Hash)
	}

	if feed.Entries[0].URL != "http://example.org/2003/12/13/atom03" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].Title != "Atom-Powered Robots Run Amok" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}

	if feed.Entries[0].Content != "Some text." {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[0].Content)
	}

	if feed.Entries[0].Author != "John Doe" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseFeedWithoutTitle(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom">
			<link rel="alternate" type="text/html" href="https://example.org/"/>
			<link rel="self" type="application/atom+xml" href="https://example.org/feed"/>
			<updated>2003-12-13T18:30:02Z</updated>
		</feed>`

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
	<feed xmlns="http://www.w3.org/2005/Atom">

	  <title>Example Feed</title>
	  <link href="http://example.org/"/>
	  <updated>2003-12-13T18:30:02Z</updated>
	  <author>
		<name>John Doe</name>
	  </author>
	  <id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>

	  <entry>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Title != "http://example.org/2003/12/13/atom03" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseFeedURL(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link rel="alternate" type="text/html" href="https://example.org/"/>
	  <link rel="self" type="application/atom+xml" href="https://example.org/feed"/>
	  <updated>2003-12-13T18:30:02Z</updated>
	</feed>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.SiteURL != "https://example.org/" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if feed.FeedURL != "https://example.org/feed" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}
}

func TestParseEntryWithRelativeURL(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title>Test</title>
		<link href="something.html"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].URL != "http://example.org/something.html" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}
}

func TestParseEntryTitleWithWhitespaces(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title>
			Some Title
		</title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Title != "Some Title" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseEntryTitleWithHTMLAndCDATA(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="html"><![CDATA[Test &#8220;Test&#8221;]]></title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Title != "Test “Test”" {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[0].Title)
	}
}

func TestParseEntryTitleWithHTML(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="html">&lt;code&gt;Test&lt;/code&gt; Test</title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Title != "Test Test" {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[0].Title)
	}
}

func TestParseEntryTitleWithXHTML(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="xhtml"><code>Test</code> Test</title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Title != "Test Test" {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[0].Title)
	}
}

func TestParseEntryWithAuthorName(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
		<author>
			<name>Me</name>
			<email>me@localhost</email>
		</author>
	  </entry>

	</feed>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Author != "Me" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseEntryWithoutAuthorName(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
		<author>
			<name/>
			<email>me@localhost</email>
		</author>
	  </entry>

	</feed>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Author != "me@localhost" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseEntryWithEnclosures(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
		<id>http://www.example.org/myfeed</id>
		<title>My Podcast Feed</title>
		<updated>2005-07-15T12:00:00Z</updated>
		<author>
		<name>John Doe</name>
		</author>
		<link href="http://example.org" />
		<link rel="self" href="http://example.org/myfeed" />
		<entry>
			<id>http://www.example.org/entries/1</id>
			<title>Atom 1.0</title>
			<updated>2005-07-15T12:00:00Z</updated>
			<link href="http://www.example.org/entries/1" />
			<summary>An overview of Atom 1.0</summary>
			<link rel="enclosure"
					type="audio/mpeg"
					title="MP3"
					href="http://www.example.org/myaudiofile.mp3"
					length="1234" />
			<link rel="enclosure"
					type="application/x-bittorrent"
					title="BitTorrent"
					href="http://www.example.org/myaudiofile.torrent"
					length="4567" />
			<content type="xhtml">
				<div xmlns="http://www.w3.org/1999/xhtml">
				<h1>Show Notes</h1>
				<ul>
					<li>00:01:00 -- Introduction</li>
					<li>00:15:00 -- Talking about Atom 1.0</li>
					<li>00:30:00 -- Wrapping up</li>
				</ul>
				</div>
			</content>
		</entry>
  	</feed>`

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

	if len(feed.Entries[0].Enclosures) != 2 {
		t.Errorf("Incorrect number of enclosures, got: %d", len(feed.Entries[0].Enclosures))
	}

	if feed.Entries[0].Enclosures[0].URL != "http://www.example.org/myaudiofile.mp3" {
		t.Errorf("Incorrect enclosure URL, got: %s", feed.Entries[0].Enclosures[0].URL)
	}

	if feed.Entries[0].Enclosures[0].MimeType != "audio/mpeg" {
		t.Errorf("Incorrect enclosure type, got: %s", feed.Entries[0].Enclosures[0].MimeType)
	}

	if feed.Entries[0].Enclosures[0].Size != 1234 {
		t.Errorf("Incorrect enclosure length, got: %d", feed.Entries[0].Enclosures[0].Size)
	}

	if feed.Entries[0].Enclosures[1].URL != "http://www.example.org/myaudiofile.torrent" {
		t.Errorf("Incorrect enclosure URL, got: %s", feed.Entries[0].Enclosures[1].URL)
	}

	if feed.Entries[0].Enclosures[1].MimeType != "application/x-bittorrent" {
		t.Errorf("Incorrect enclosure type, got: %s", feed.Entries[0].Enclosures[1].MimeType)
	}

	if feed.Entries[0].Enclosures[1].Size != 4567 {
		t.Errorf("Incorrect enclosure length, got: %d", feed.Entries[0].Enclosures[1].Size)
	}
}

func TestParseEntryWithPublished(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<published>2003-12-13T18:30:02Z</published>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if !feed.Entries[0].Date.Equal(time.Date(2003, time.December, 13, 18, 30, 2, 0, time.UTC)) {
		t.Errorf("Incorrect entry date, got: %v", feed.Entries[0].Date)
	}
}

func TestParseEntryWithPublishedAndUpdated(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<published>2002-11-12T18:30:02Z</published>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if !feed.Entries[0].Date.Equal(time.Date(2003, time.December, 13, 18, 30, 2, 0, time.UTC)) {
		t.Errorf("Incorrect entry date, got: %v", feed.Entries[0].Date)
	}
}

func TestParseInvalidXml(t *testing.T) {
	data := `garbage`
	_, err := Parse(bytes.NewBufferString(data))
	if err == nil {
		t.Error("Parse should returns an error")
	}
}

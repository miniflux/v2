// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

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

	feed, err := Parse("http://example.org/feed.xml", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "Example Feed" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "http://example.org/feed.xml" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "http://example.org/" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if feed.IconURL != "" {
		t.Errorf("Incorrect icon URL, got: %s", feed.IconURL)
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

	if feed.Entries[0].CommentsURL != "" {
		t.Errorf("Incorrect entry Comments URL, got: %s", feed.Entries[0].CommentsURL)
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

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "https://example.org/" {
		t.Errorf("Incorrect feed title, got: %s", feed.Title)
	}
}

func TestParseEntryWithoutTitleButWithURL(t *testing.T) {
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
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Title != "http://example.org/2003/12/13/atom03" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseEntryWithoutTitleButWithSummary(t *testing.T) {
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

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Title != "Some text." {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseEntryWithoutTitleButWithXHTMLContent(t *testing.T) {
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
		<content type="xhtml">
			<div xmlns="http://www.w3.org/1999/xhtml">AT&amp;T bought <b>by SBC</b>!</div>
		</content>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Title != "AT&T bought by SBC!" {
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

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.SiteURL != "https://example.org/" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if feed.FeedURL != "https://example.org/feed" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}
}

func TestParseFeedWithRelativeURL(t *testing.T) {
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

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
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

	feed, err := Parse("https://example.net/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].URL != "http://example.org/something.html" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}
}

func TestParseEntryURLWithTextHTMLType(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title>Test</title>
		<link href="http://example.org/something.html" type="text/html"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.net/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].URL != "http://example.org/something.html" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}
}

func TestParseEntryURLWithNoRelAndNoType(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title>Test</title>
		<link href="http://example.org/something.html"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.net/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].URL != "http://example.org/something.html" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}
}

func TestParseEntryURLWithAlternateRel(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title>Test</title>
		<link href="http://example.org/something.html" rel="alternate"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.net/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
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

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Title != "Some Title" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseEntryWithPlainTextTitle(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="text">AT&amp;T bought by SBC!</title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	  <entry>
		<title>AT&amp;T bought by SBC!</title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	expected := `AT&T bought by SBC!`
	for i := 0; i < 2; i++ {
		if feed.Entries[i].Title != expected {
			t.Errorf("Incorrect title for entry #%d, got: %q", i, feed.Entries[i].Title)
		}
	}
}

func TestParseEntryWithHTMLTitle(t *testing.T) {
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

	  <entry>
		<title type="html"><![CDATA[Test &#8220;Test&#8221;]]></title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	  <entry>
		<title>
			<![CDATA[Entry title with space around CDATA]]>
		</title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Title != "<code>Test</code> Test" {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[0].Title)
	}

	if feed.Entries[1].Title != "Test “Test”" {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[1].Title)
	}

	if feed.Entries[2].Title != "Entry title with space around CDATA" {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[2].Title)
	}
}

func TestParseEntryWithXHTMLTitle(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="xhtml">
			<div xmlns="http://www.w3.org/1999/xhtml">
				This is <b>XHTML</b> content.
	 		</div>
		</title>
		<link href="http://example.org/b"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Title != `This is <b>XHTML</b> content.` {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[1].Title)
	}
}

func TestParseEntryWithEmptyXHTMLTitle(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="xhtml">
			<div xmlns="http://www.w3.org/1999/xhtml"/>
		</title>
		<link href="http://example.org/entry"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Title != `http://example.org/entry` {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[0].Title)
	}
}

func TestParseEntryWithXHTMLTitleWithoutDiv(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="xhtml">
		  test
		</title>
		<link href="http://example.org/entry"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Title != `test` {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[0].Title)
	}
}

func TestParseEntryWithNumericCharacterReferenceTitle(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title>&#931; &#xDF;</title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Title != "Σ ß" {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[0].Title)
	}
}

func TestParseEntryWithDoubleEncodedEntitiesTitle(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title>&amp;#39;AT&amp;amp;T&amp;#39;</title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Title != `&#39;AT&amp;T&#39;` {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[0].Title)
	}
}

func TestParseEntryWithXHTMLSummary(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="xhtml">Example</title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary type="xhtml"><div xmlns="http://www.w3.org/1999/xhtml"><p>Test: <code>std::unique_ptr&lt;S&gt;</code></p></div></summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Content != `<p>Test: <code>std::unique_ptr&lt;S&gt;</code></p>` {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[1].Content)
	}
}

func TestParseEntryWithHTMLSummary(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/1"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary type="html">&lt;code&gt;std::unique_ptr&amp;lt;S&amp;gt;&lt;/code&gt;</summary>
	  </entry>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/2"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary type="text/html">&lt;code&gt;std::unique_ptr&amp;lt;S&amp;gt;&lt;/code&gt;</summary>
	  </entry>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/3"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary type="html"><![CDATA[<code>std::unique_ptr&lt;S&gt;</code>]]></summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	expected := `<code>std::unique_ptr&lt;S&gt;</code>`
	for i := 0; i < 3; i++ {
		if feed.Entries[i].Content != expected {
			t.Errorf("Incorrect content for entry #%d, got: %q", i, feed.Entries[i].Content)
		}
	}
}

func TestParseEntryWithTextSummary(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/a"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>AT&amp;T &lt;S&gt;</summary>
	  </entry>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/b"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary type="text">AT&amp;T &lt;S&gt;</summary>
	  </entry>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/c"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary type="text/plain">AT&amp;T &lt;S&gt;</summary>
	  </entry>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/d"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary type="text"><![CDATA[AT&T <S>]]></summary>
	  </entry>
	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	expected := `AT&amp;T &lt;S&gt;`
	for i := 0; i < 4; i++ {
		if feed.Entries[i].Content != expected {
			t.Errorf("Incorrect content for entry #%d, got: %q", i, feed.Entries[i].Content)
		}
	}
}

func TestParseEntryWithTextContent(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/a"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<content>AT&amp;T &lt;S&gt;</content>
	  </entry>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/b"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<content type="text">AT&amp;T &lt;S&gt;</content>
	  </entry>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/c"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<content type="text/plain">AT&amp;T &lt;S&gt;</content>
	  </entry>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/d"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<content><![CDATA[AT&T <S>]]></content>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	expected := `AT&amp;T &lt;S&gt;`
	for i := 0; i < 4; i++ {
		if feed.Entries[i].Content != expected {
			t.Errorf("Incorrect content for entry #%d, got: %q", i, feed.Entries[i].Content)
		}
	}
}

func TestParseEntryWithHTMLContent(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/a"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<content type="html">AT&amp;amp;T bought &lt;b&gt;by SBC&lt;/b&gt;!</content>
	  </entry>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/b"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<content type="text/html">AT&amp;amp;T bought &lt;b&gt;by SBC&lt;/b&gt;!</content>
	  </entry>

	  <entry>
		<title type="html">Example</title>
		<link href="http://example.org/c"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<content type="html"><![CDATA[AT&amp;T bought <b>by SBC</b>!]]></content>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	expected := `AT&amp;T bought <b>by SBC</b>!`
	for i := 0; i < 3; i++ {
		if feed.Entries[i].Content != expected {
			t.Errorf("Incorrect content for entry #%d, got: %q", i, feed.Entries[i].Content)
		}
	}
}

func TestParseEntryWithXHTMLContent(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<title>Example</title>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<content type="xhtml">
			<div xmlns="http://www.w3.org/1999/xhtml">AT&amp;T bought <b>by SBC</b>!</div>
		</content>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Content != `AT&amp;T bought <b>by SBC</b>!` {
		t.Errorf("Incorrect entry content, got: %q", feed.Entries[0].Content)
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

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
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

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Author != "me@localhost" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseEntryWithMultipleAuthors(t *testing.T) {
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
			<name>Alice</name>
		</author>
		<author>
			<name>Bob</name>
		</author>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Author != "Alice, Bob" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseEntryWithoutAuthor(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>
	  <author>
		<name>John Doe</name>
	  </author>

	  <entry>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Author != "John Doe" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseFeedWithMultipleAuthors(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>
	  <author>
		<name>Alice</name>
	  </author>
	  <author>
		<name>Bob</name>
	  </author>

	  <entry>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Author != "Alice, Bob" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseFeedWithoutAuthor(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>

	  <entry>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Author != "" {
		t.Errorf("Incorrect entry author, got: %q", feed.Entries[0].Author)
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

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].URL != "http://www.example.org/entries/1" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if len(feed.Entries[0].Enclosures) != 2 {
		t.Fatalf("Incorrect number of enclosures, got: %d", len(feed.Entries[0].Enclosures))
	}

	expectedResults := []struct {
		url      string
		mimeType string
		size     int64
	}{
		{"http://www.example.org/myaudiofile.mp3", "audio/mpeg", 1234},
		{"http://www.example.org/myaudiofile.torrent", "application/x-bittorrent", 4567},
	}

	for index, enclosure := range feed.Entries[0].Enclosures {
		if expectedResults[index].url != enclosure.URL {
			t.Errorf(`Unexpected enclosure URL, got %q instead of %q`, enclosure.URL, expectedResults[index].url)
		}

		if expectedResults[index].mimeType != enclosure.MimeType {
			t.Errorf(`Unexpected enclosure type, got %q instead of %q`, enclosure.MimeType, expectedResults[index].mimeType)
		}

		if expectedResults[index].size != enclosure.Size {
			t.Errorf(`Unexpected enclosure size, got %d instead of %d`, enclosure.Size, expectedResults[index].size)
		}
	}
}

func TestParseEntryWithoutEnclosureURL(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
		<id>http://www.example.org/myfeed</id>
		<title>My Podcast Feed</title>
		<updated>2005-07-15T12:00:00Z</updated>
		<link href="http://example.org" />
		<link rel="self" href="http://example.org/myfeed" />
		<entry>
			<id>http://www.example.org/entries/1</id>
			<title>Atom 1.0</title>
			<updated>2005-07-15T12:00:00Z</updated>
			<link href="http://www.example.org/entries/1" />
			<summary>An overview of Atom 1.0</summary>
			<link rel="enclosure" href="" length="0" />
			<content type="xhtml">Test</content>
		</entry>
  	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].URL != "http://www.example.org/entries/1" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if len(feed.Entries[0].Enclosures) != 0 {
		t.Fatalf("Incorrect number of enclosures, got: %d", len(feed.Entries[0].Enclosures))
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

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
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

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if !feed.Entries[0].Date.Equal(time.Date(2002, time.November, 12, 18, 30, 2, 0, time.UTC)) {
		t.Errorf("Incorrect entry date, got: %v", feed.Entries[0].Date)
	}
}

func TestParseInvalidXml(t *testing.T) {
	data := `garbage`
	_, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err == nil {
		t.Error("Parse should returns an error")
	}
}

func TestParseTitleWithSingleQuote(t *testing.T) {
	data := `
		<?xml version="1.0" encoding="utf-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom">
			<title>' or ’</title>
			<link href="http://example.org/"/>
		</feed>
	`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "' or ’" {
		t.Errorf(`Incorrect title, got: %q`, feed.Title)
	}
}

func TestParseTitleWithEncodedSingleQuote(t *testing.T) {
	data := `
		<?xml version="1.0" encoding="utf-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom">
			<title type="html">Test&#39;s Blog</title>
			<link href="http://example.org/"/>
		</feed>
	`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "Test's Blog" {
		t.Errorf(`Incorrect title, got: %q`, feed.Title)
	}
}

func TestParseTitleWithSingleQuoteAndHTMLType(t *testing.T) {
	data := `
		<?xml version="1.0" encoding="utf-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom">
			<title type="html">O’Hara</title>
			<link href="http://example.org/"/>
		</feed>
	`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "O’Hara" {
		t.Errorf(`Incorrect title, got: %q`, feed.Title)
	}
}

func TestParseWithHTMLEntity(t *testing.T) {
	data := `
		<?xml version="1.0" encoding="utf-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom">
			<title>Example &nbsp; Feed</title>
			<link href="http://example.org/"/>
		</feed>
	`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "Example \u00a0 Feed" {
		t.Errorf(`Incorrect title, got: %q`, feed.Title)
	}
}

func TestParseWithInvalidCharacterEntity(t *testing.T) {
	data := `
		<?xml version="1.0" encoding="utf-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom">
			<title>Example Feed</title>
			<link href="http://example.org/a&b"/>
		</feed>
	`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.SiteURL != "http://example.org/a&b" {
		t.Errorf(`Incorrect URL, got: %q`, feed.SiteURL)
	}
}

func TestParseMediaGroup(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/">
		<id>http://www.example.org/myfeed</id>
		<title>My Video Feed</title>
		<updated>2005-07-15T12:00:00Z</updated>
		<link href="http://example.org" />
		<link rel="self" href="http://example.org/myfeed" />
		<entry>
			<id>http://www.example.org/entries/1</id>
			<title>Some Video</title>
			<updated>2005-07-15T12:00:00Z</updated>
			<link href="http://www.example.org/entries/1" />
			<media:group>
				<media:title>Another title</media:title>
				<media:content url="https://www.youtube.com/v/abcd" type="application/x-shockwave-flash" width="640" height="390"/>
				<media:thumbnail url="https://example.org/thumbnail.jpg" width="480" height="360"/>
				<media:description>Some description
A website: http://example.org/</media:description>
			</media:group>
		</entry>
  	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].URL != "http://www.example.org/entries/1" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].Content != `Some description<br>A website: <a href="http://example.org/">http://example.org/</a>` {
		t.Errorf("Incorrect entry content, got: %q", feed.Entries[0].Content)
	}

	if len(feed.Entries[0].Enclosures) != 2 {
		t.Fatalf("Incorrect number of enclosures, got: %d", len(feed.Entries[0].Enclosures))
	}

	expectedResults := []struct {
		url      string
		mimeType string
		size     int64
	}{
		{"https://example.org/thumbnail.jpg", "image/*", 0},
		{"https://www.youtube.com/v/abcd", "application/x-shockwave-flash", 0},
	}

	for index, enclosure := range feed.Entries[0].Enclosures {
		if expectedResults[index].url != enclosure.URL {
			t.Errorf(`Unexpected enclosure URL, got %q instead of %q`, enclosure.URL, expectedResults[index].url)
		}

		if expectedResults[index].mimeType != enclosure.MimeType {
			t.Errorf(`Unexpected enclosure type, got %q instead of %q`, enclosure.MimeType, expectedResults[index].mimeType)
		}

		if expectedResults[index].size != enclosure.Size {
			t.Errorf(`Unexpected enclosure size, got %d instead of %d`, enclosure.Size, expectedResults[index].size)
		}
	}
}

func TestParseMediaElements(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/">
		<id>http://www.example.org/myfeed</id>
		<title>My Video Feed</title>
		<updated>2005-07-15T12:00:00Z</updated>
		<link href="http://example.org" />
		<link rel="self" href="http://example.org/myfeed" />
		<entry>
			<id>http://www.example.org/entries/1</id>
			<title>Some Video</title>
			<updated>2005-07-15T12:00:00Z</updated>
			<link href="http://www.example.org/entries/1" />
			<media:title>Another title</media:title>
			<media:content url="https://www.youtube.com/v/abcd" type="application/x-shockwave-flash" width="640" height="390"/>
			<media:thumbnail url="https://example.org/thumbnail.jpg" width="480" height="360"/>
			<media:description>Some description
A website: http://example.org/</media:description>
		</entry>
  	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].URL != "http://www.example.org/entries/1" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].Content != `Some description<br>A website: <a href="http://example.org/">http://example.org/</a>` {
		t.Errorf("Incorrect entry content, got: %q", feed.Entries[0].Content)
	}

	if len(feed.Entries[0].Enclosures) != 2 {
		t.Fatalf("Incorrect number of enclosures, got: %d", len(feed.Entries[0].Enclosures))
	}

	expectedResults := []struct {
		url      string
		mimeType string
		size     int64
	}{
		{"https://example.org/thumbnail.jpg", "image/*", 0},
		{"https://www.youtube.com/v/abcd", "application/x-shockwave-flash", 0},
	}

	for index, enclosure := range feed.Entries[0].Enclosures {
		if expectedResults[index].url != enclosure.URL {
			t.Errorf(`Unexpected enclosure URL, got %q instead of %q`, enclosure.URL, expectedResults[index].url)
		}

		if expectedResults[index].mimeType != enclosure.MimeType {
			t.Errorf(`Unexpected enclosure type, got %q instead of %q`, enclosure.MimeType, expectedResults[index].mimeType)
		}

		if expectedResults[index].size != enclosure.Size {
			t.Errorf(`Unexpected enclosure size, got %d instead of %d`, enclosure.Size, expectedResults[index].size)
		}
	}
}

func TestParseRepliesLinkRelationWithHTMLType(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom"
			xmlns:thr="http://purl.org/syndication/thread/1.0">
		<id>http://www.example.org/myfeed</id>
		<title>My Example Feed</title>
		<updated>2005-07-28T12:00:00Z</updated>
		<link href="http://www.example.org/myfeed" />
		<author><name>James</name></author>
		<entry>
			<id>tag:entries.com,2005:1</id>
			<title>My original entry</title>
			<updated>2006-03-01T12:12:12Z</updated>
			<link href="http://www.example.org/entries/1" />
			<link rel="replies"
				type="application/atom+xml"
				href="http://www.example.org/mycommentsfeed.xml"
				thr:count="10" thr:updated="2005-07-28T12:10:00Z" />
			<link rel="replies"
				type="text/html"
				href="http://www.example.org/comments.html"
				thr:count="10" thr:updated="2005-07-28T12:10:00Z" />
			<summary>This is my original entry</summary>
		</entry>
	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].URL != "http://www.example.org/entries/1" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].CommentsURL != "http://www.example.org/comments.html" {
		t.Errorf("Incorrect entry comments URL, got: %s", feed.Entries[0].CommentsURL)
	}
}

func TestParseRepliesLinkRelationWithXHTMLType(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom"
			xmlns:thr="http://purl.org/syndication/thread/1.0">
		<id>http://www.example.org/myfeed</id>
		<title>My Example Feed</title>
		<updated>2005-07-28T12:00:00Z</updated>
		<link href="http://www.example.org/myfeed" />
		<author><name>James</name></author>
		<entry>
			<id>tag:entries.com,2005:1</id>
			<title>My original entry</title>
			<updated>2006-03-01T12:12:12Z</updated>
			<link href="http://www.example.org/entries/1" />
			<link rel="replies"
				type="application/atom+xml"
				href="http://www.example.org/mycommentsfeed.xml"
				thr:count="10" thr:updated="2005-07-28T12:10:00Z" />
			<link rel="replies"
				type="application/xhtml+xml"
				href="http://www.example.org/comments.xhtml"
				thr:count="10" thr:updated="2005-07-28T12:10:00Z" />
			<summary>This is my original entry</summary>
		</entry>
	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].URL != "http://www.example.org/entries/1" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].CommentsURL != "http://www.example.org/comments.xhtml" {
		t.Errorf("Incorrect entry comments URL, got: %s", feed.Entries[0].CommentsURL)
	}
}

func TestParseRepliesLinkRelationWithNoType(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom"
			xmlns:thr="http://purl.org/syndication/thread/1.0">
		<id>http://www.example.org/myfeed</id>
		<title>My Example Feed</title>
		<updated>2005-07-28T12:00:00Z</updated>
		<link href="http://www.example.org/myfeed" />
		<author><name>James</name></author>
		<entry>
			<id>tag:entries.com,2005:1</id>
			<title>My original entry</title>
			<updated>2006-03-01T12:12:12Z</updated>
			<link href="http://www.example.org/entries/1" />
			<link rel="replies"
				href="http://www.example.org/mycommentsfeed.xml"
				thr:count="10" thr:updated="2005-07-28T12:10:00Z" />
			<summary>This is my original entry</summary>
		</entry>
	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].URL != "http://www.example.org/entries/1" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].CommentsURL != "" {
		t.Errorf("Incorrect entry comments URL, got: %s", feed.Entries[0].CommentsURL)
	}
}

func TestAbsoluteCommentsURL(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom"
			xmlns:thr="http://purl.org/syndication/thread/1.0">
		<id>http://www.example.org/myfeed</id>
		<title>My Example Feed</title>
		<updated>2005-07-28T12:00:00Z</updated>
		<link href="http://www.example.org/myfeed" />
		<author><name>James</name></author>
		<entry>
			<id>tag:entries.com,2005:1</id>
			<title>My original entry</title>
			<updated>2006-03-01T12:12:12Z</updated>
			<link href="http://www.example.org/entries/1" />
			<link rel="replies"
				type="text/html"
				href="invalid url"
				thr:count="10" thr:updated="2005-07-28T12:10:00Z" />
			<summary>This is my original entry</summary>
		</entry>
	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].URL != "http://www.example.org/entries/1" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].CommentsURL != "" {
		t.Errorf("Incorrect entry comments URL, got: %s", feed.Entries[0].CommentsURL)
	}
}

func TestParseFeedWithCategories(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
	  <title>Example Feed</title>
	  <link href="http://example.org/"/>
	  <author>
		<name>Alice</name>
	  </author>
	  <author>
		<name>Bob</name>
	  </author>

	  <entry>
		<link href="http://example.org/2003/12/13/atom03"/>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<updated>2003-12-13T18:30:02Z</updated>
		<summary>Some text.</summary>
		<category term='Tech' />
		<category term='Technology' label='Science' />
	  </entry>

	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries[0].Tags) != 2 {
		t.Errorf("Incorrect number of tags, got: %d", len(feed.Entries[0].Tags))
	}

	expected := "Tech"
	result := feed.Entries[0].Tags[0]
	if result != expected {
		t.Errorf("Incorrect entry category, got %q instead of %q", result, expected)
	}

	expected = "Science"
	result = feed.Entries[0].Tags[1]
	if result != expected {
		t.Errorf("Incorrect entry category, got %q instead of %q", result, expected)
	}
}

func TestParseFeedWithIconURL(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom">
		<title>Example Feed</title>
		<link href="http://example.org/"/>
		<icon>http://example.org/icon.png</icon>
	</feed>`

	feed, err := Parse("https://example.org/", bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if feed.IconURL != "http://example.org/icon.png" {
		t.Errorf("Incorrect icon URL, got: %s", feed.IconURL)
	}
}

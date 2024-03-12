// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package atom // import "miniflux.app/v2/internal/reader/atom"

import (
	"bytes"
	"testing"
	"time"
)

func TestParseAtom03(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed version="0.3" xmlns="http://purl.org/atom/ns#">
		<title>dive into mark</title>
		<link rel="alternate" type="text/html" href="http://diveintomark.org/"/>
		<modified>2003-12-13T18:30:02Z</modified>
		<author><name>Mark Pilgrim</name></author>
		<entry>
			<title>Atom 0.3 snapshot</title>
			<link rel="alternate" type="text/html" href="http://diveintomark.org/2003/12/13/atom03"/>
			<id>tag:diveintomark.org,2003:3.2397</id>
			<issued>2003-12-13T08:29:29-04:00</issued>
			<modified>2003-12-13T18:30:02Z</modified>
			<summary type="text/plain">It&apos;s a test</summary>
			<content type="text/html" mode="escaped"><![CDATA[<p>HTML content</p>]]></content>
		</entry>
	</feed>`

	feed, err := Parse("http://diveintomark.org/", bytes.NewReader([]byte(data)), "0.3")
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "dive into mark" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "http://diveintomark.org/" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "http://diveintomark.org/" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	tz := time.FixedZone("Test Case Time", -int((4 * time.Hour).Seconds()))
	if !feed.Entries[0].Date.Equal(time.Date(2003, time.December, 13, 8, 29, 29, 0, tz)) {
		t.Errorf("Incorrect entry date, got: %v", feed.Entries[0].Date)
	}

	if feed.Entries[0].Hash != "b70d30334b808f32e66eb19fabb263525cecd18f205720b583e84f7f295cf728" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[0].Hash)
	}

	if feed.Entries[0].URL != "http://diveintomark.org/2003/12/13/atom03" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].Title != "Atom 0.3 snapshot" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}

	if feed.Entries[0].Content != "<p>HTML content</p>" {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[0].Content)
	}

	if feed.Entries[0].Author != "Mark Pilgrim" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseAtom03WithoutFeedTitle(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed version="0.3" xmlns="http://purl.org/atom/ns#">
		<link rel="alternate" type="text/html" href="http://diveintomark.org/"/>
		<modified>2003-12-13T18:30:02Z</modified>
		<author><name>Mark Pilgrim</name></author>
		<entry>
			<title>Atom 0.3 snapshot</title>
			<link rel="alternate" type="text/html" href="http://diveintomark.org/2003/12/13/atom03"/>
			<id>tag:diveintomark.org,2003:3.2397</id>
		</entry>
	</feed>`

	feed, err := Parse("http://diveintomark.org/", bytes.NewReader([]byte(data)), "0.3")
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "http://diveintomark.org/" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}
}

func TestParseAtom03WithoutEntryTitleButWithLink(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed version="0.3" xmlns="http://purl.org/atom/ns#">
		<title>dive into mark</title>
		<link rel="alternate" type="text/html" href="http://diveintomark.org/"/>
		<modified>2003-12-13T18:30:02Z</modified>
		<author><name>Mark Pilgrim</name></author>
		<entry>
			<link rel="alternate" type="text/html" href="http://diveintomark.org/2003/12/13/atom03"/>
			<id>tag:diveintomark.org,2003:3.2397</id>
		</entry>
	</feed>`

	feed, err := Parse("http://diveintomark.org/", bytes.NewReader([]byte(data)), "0.3")
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Title != "http://diveintomark.org/2003/12/13/atom03" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseAtom03WithoutEntryTitleButWithSummary(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed version="0.3" xmlns="http://purl.org/atom/ns#">
		<title>dive into mark</title>
		<link rel="alternate" type="text/html" href="http://diveintomark.org/"/>
		<modified>2003-12-13T18:30:02Z</modified>
		<author><name>Mark Pilgrim</name></author>
		<entry>
			<link rel="alternate" type="text/html" href="http://diveintomark.org/2003/12/13/atom03"/>
			<id>tag:diveintomark.org,2003:3.2397</id>
			<summary type="text/plain">It&apos;s a test</summary>
		</entry>
	</feed>`

	feed, err := Parse("http://diveintomark.org/", bytes.NewReader([]byte(data)), "0.3")
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Title != "It's a test" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseAtom03WithoutEntryTitleButWithXMLContent(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed version="0.3" xmlns="http://purl.org/atom/ns#">
		<title>dive into mark</title>
		<link rel="alternate" type="text/html" href="http://diveintomark.org/"/>
		<modified>2003-12-13T18:30:02Z</modified>
		<author><name>Mark Pilgrim</name></author>
		<entry>
			<link rel="alternate" type="text/html" href="http://diveintomark.org/2003/12/13/atom03"/>
			<id>tag:diveintomark.org,2003:3.2397</id>
			<content mode="xml" type="text/html"><p>Some text.</p></content>
		</entry>
	</feed>`

	feed, err := Parse("http://diveintomark.org/", bytes.NewReader([]byte(data)), "0.3")
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Title != "Some text." {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}
}

func TestParseAtom03WithSummaryOnly(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed version="0.3" xmlns="http://purl.org/atom/ns#">
		<title>dive into mark</title>
		<link rel="alternate" type="text/html" href="http://diveintomark.org/"/>
		<modified>2003-12-13T18:30:02Z</modified>
		<author><name>Mark Pilgrim</name></author>
		<entry>
			<title>Atom 0.3 snapshot</title>
			<link rel="alternate" type="text/html" href="http://diveintomark.org/2003/12/13/atom03"/>
			<id>tag:diveintomark.org,2003:3.2397</id>
			<issued>2003-12-13T08:29:29-04:00</issued>
			<modified>2003-12-13T18:30:02Z</modified>
			<summary type="text/plain">It&apos;s a test</summary>
		</entry>
	</feed>`

	feed, err := Parse("http://diveintomark.org/", bytes.NewReader([]byte(data)), "0.3")
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Content != "It&#39;s a test" {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[0].Content)
	}
}

func TestParseAtom03WithXMLContent(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed version="0.3" xmlns="http://purl.org/atom/ns#">
		<title>dive into mark</title>
		<link rel="alternate" type="text/html" href="http://diveintomark.org/"/>
		<modified>2003-12-13T18:30:02Z</modified>
		<author><name>Mark Pilgrim</name></author>
		<entry>
			<title>Atom 0.3 snapshot</title>
			<link rel="alternate" type="text/html" href="http://diveintomark.org/2003/12/13/atom03"/>
			<id>tag:diveintomark.org,2003:3.2397</id>
			<issued>2003-12-13T08:29:29-04:00</issued>
			<modified>2003-12-13T18:30:02Z</modified>
			<content mode="xml" type="text/html"><p>Some text.</p></content>
		</entry>
	</feed>`

	feed, err := Parse("http://diveintomark.org/", bytes.NewReader([]byte(data)), "0.3")
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Content != "<p>Some text.</p>" {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[0].Content)
	}
}

func TestParseAtom03WithBase64Content(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<feed version="0.3" xmlns="http://purl.org/atom/ns#">
		<title>dive into mark</title>
		<link rel="alternate" type="text/html" href="http://diveintomark.org/"/>
		<modified>2003-12-13T18:30:02Z</modified>
		<author><name>Mark Pilgrim</name></author>
		<entry>
			<title>Atom 0.3 snapshot</title>
			<link rel="alternate" type="text/html" href="http://diveintomark.org/2003/12/13/atom03"/>
			<id>tag:diveintomark.org,2003:3.2397</id>
			<issued>2003-12-13T08:29:29-04:00</issued>
			<modified>2003-12-13T18:30:02Z</modified>
			<content mode="base64" type="text/html">PHA+U29tZSB0ZXh0LjwvcD4=</content>
		</entry>
	</feed>`

	feed, err := Parse("http://diveintomark.org/", bytes.NewReader([]byte(data)), "0.3")
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Content != "<p>Some text.</p>" {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[0].Content)
	}
}

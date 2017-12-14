// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rdf

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/miniflux/miniflux/errors"
)

func TestParseRDFSample(t *testing.T) {
	data := `
	<?xml version="1.0"?>

	<rdf:RDF
	  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	  xmlns="http://purl.org/rss/1.0/"
	>

	  <channel rdf:about="http://www.xml.com/xml/news.rss">
		<title>XML.com</title>
		<link>http://xml.com/pub</link>
		<description>
		  XML.com features a rich mix of information and services
		  for the XML community.
		</description>

		<image rdf:resource="http://xml.com/universal/images/xml_tiny.gif" />

		<items>
		  <rdf:Seq>
			<rdf:li resource="http://xml.com/pub/2000/08/09/xslt/xslt.html" />
			<rdf:li resource="http://xml.com/pub/2000/08/09/rdfdb/index.html" />
		  </rdf:Seq>
		</items>

		<textinput rdf:resource="http://search.xml.com" />

	  </channel>

	  <image rdf:about="http://xml.com/universal/images/xml_tiny.gif">
		<title>XML.com</title>
		<link>http://www.xml.com</link>
		<url>http://xml.com/universal/images/xml_tiny.gif</url>
	  </image>

	  <item rdf:about="http://xml.com/pub/2000/08/09/xslt/xslt.html">
		<title>Processing Inclusions with XSLT</title>
		<link>http://xml.com/pub/2000/08/09/xslt/xslt.html</link>
		<description>
		 Processing document inclusions with general XML tools can be
		 problematic. This article proposes a way of preserving inclusion
		 information through SAX-based processing.
		</description>
	  </item>

	  <item rdf:about="http://xml.com/pub/2000/08/09/rdfdb/index.html">
		<title>Putting RDF to Work</title>
		<link>http://xml.com/pub/2000/08/09/rdfdb/index.html</link>
		<description>
		 Tool and API support for the Resource Description Framework
		 is slowly coming of age. Edd Dumbill takes a look at RDFDB,
		 one of the most exciting new RDF toolkits.
		</description>
	  </item>

	  <textinput rdf:about="http://search.xml.com">
		<title>Search XML.com</title>
		<description>Search XML.com's XML collection</description>
		<name>s</name>
		<link>http://search.xml.com</link>
	  </textinput>

	</rdf:RDF>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "XML.com" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "http://xml.com/pub" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if len(feed.Entries) != 2 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[1].Hash != "8aaeee5d3ab50351422fbded41078ee88c73bf1441085b16a8c09fd90a7db321" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[0].Hash)
	}

	if feed.Entries[1].URL != "http://xml.com/pub/2000/08/09/rdfdb/index.html" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[1].Title != "Putting RDF to Work" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}

	if strings.HasSuffix(feed.Entries[1].Content, "Tool and API support") {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[0].Content)
	}

	if feed.Entries[1].Date.Year() != time.Now().Year() {
		t.Errorf("Entry date should not be empty")
	}
}

func TestParseRDFSampleWithDublinCore(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>

	<rdf:RDF
	  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	  xmlns:dc="http://purl.org/dc/elements/1.1/"
	  xmlns:sy="http://purl.org/rss/1.0/modules/syndication/"
	  xmlns:co="http://purl.org/rss/1.0/modules/company/"
	  xmlns:ti="http://purl.org/rss/1.0/modules/textinput/"
	  xmlns="http://purl.org/rss/1.0/"
	>

	  <channel rdf:about="http://meerkat.oreillynet.com/?_fl=rss1.0">
		<title>Meerkat</title>
		<link>http://meerkat.oreillynet.com</link>
		<description>Meerkat: An Open Wire Service</description>
		<dc:publisher>The O'Reilly Network</dc:publisher>
		<dc:creator>Rael Dornfest (mailto:rael@oreilly.com)</dc:creator>
		<dc:rights>Copyright &#169; 2000 O'Reilly &amp; Associates, Inc.</dc:rights>
		<dc:date>2000-01-01T12:00+00:00</dc:date>
		<sy:updatePeriod>hourly</sy:updatePeriod>
		<sy:updateFrequency>2</sy:updateFrequency>
		<sy:updateBase>2000-01-01T12:00+00:00</sy:updateBase>

		<image rdf:resource="http://meerkat.oreillynet.com/icons/meerkat-powered.jpg" />

		<items>
		  <rdf:Seq>
			<rdf:li resource="http://c.moreover.com/click/here.pl?r123" />
		  </rdf:Seq>
		</items>

		<textinput rdf:resource="http://meerkat.oreillynet.com" />

	  </channel>

	  <image rdf:about="http://meerkat.oreillynet.com/icons/meerkat-powered.jpg">
		<title>Meerkat Powered!</title>
		<url>http://meerkat.oreillynet.com/icons/meerkat-powered.jpg</url>
		<link>http://meerkat.oreillynet.com</link>
	  </image>

	  <item rdf:about="http://c.moreover.com/click/here.pl?r123">
		<title>XML: A Disruptive Technology</title>
		<link>http://c.moreover.com/click/here.pl?r123</link>
		<dc:description>
		  XML is placing increasingly heavy loads on the existing technical
		  infrastructure of the Internet.
		</dc:description>
		<dc:publisher>The O'Reilly Network</dc:publisher>
		<dc:creator>Simon St.Laurent (mailto:simonstl@simonstl.com)</dc:creator>
		<dc:rights>Copyright &#169; 2000 O'Reilly &amp; Associates, Inc.</dc:rights>
		<dc:subject>XML</dc:subject>
		<co:name>XML.com</co:name>
		<co:market>NASDAQ</co:market>
		<co:symbol>XML</co:symbol>
	  </item>

	  <textinput rdf:about="http://meerkat.oreillynet.com">
		<title>Search Meerkat</title>
		<description>Search Meerkat's RSS Database...</description>
		<name>s</name>
		<link>http://meerkat.oreillynet.com/</link>
		<ti:function>search</ti:function>
		<ti:inputType>regex</ti:inputType>
	  </textinput>

	</rdf:RDF>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Title != "Meerkat" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "" {
		t.Errorf("Incorrect feed URL, got: %s", feed.FeedURL)
	}

	if feed.SiteURL != "http://meerkat.oreillynet.com" {
		t.Errorf("Incorrect site URL, got: %s", feed.SiteURL)
	}

	if len(feed.Entries) != 1 {
		t.Errorf("Incorrect number of entries, got: %d", len(feed.Entries))
	}

	if feed.Entries[0].Hash != "fa4ef7c300b175ca66f92f226b5dba5caa2a9619f031101bf56e5b884b02cd97" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[0].Hash)
	}

	if feed.Entries[0].URL != "http://c.moreover.com/click/here.pl?r123" {
		t.Errorf("Incorrect entry URL, got: %s", feed.Entries[0].URL)
	}

	if feed.Entries[0].Title != "XML: A Disruptive Technology" {
		t.Errorf("Incorrect entry title, got: %s", feed.Entries[0].Title)
	}

	if strings.HasSuffix(feed.Entries[0].Content, "XML is placing increasingly") {
		t.Errorf("Incorrect entry content, got: %s", feed.Entries[0].Content)
	}

	if feed.Entries[0].Author != "Simon St.Laurent (mailto:simonstl@simonstl.com)" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseItemWithOnlyFeedAuthor(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>

	<rdf:RDF
	  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	  xmlns:dc="http://purl.org/dc/elements/1.1/"
	  xmlns="http://purl.org/rss/1.0/"
	>

	  <channel rdf:about="http://meerkat.oreillynet.com/?_fl=rss1.0">
		<title>Meerkat</title>
		<link>http://meerkat.oreillynet.com</link>
		<dc:creator>Rael Dornfest (mailto:rael@oreilly.com)</dc:creator>
	  </channel>

	  <item rdf:about="http://c.moreover.com/click/here.pl?r123">
		<title>XML: A Disruptive Technology</title>
		<link>http://c.moreover.com/click/here.pl?r123</link>
		<dc:description>
		  XML is placing increasingly heavy loads on the existing technical
		  infrastructure of the Internet.
		</dc:description>
	  </item>
	</rdf:RDF>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Author != "Rael Dornfest (mailto:rael@oreilly.com)" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseItemRelativeURL(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/">
	  <channel>
			<title>Example</title>
			<link>http://example.org</link>
	  </channel>

	  <item>
			<title>Title</title>
			<description>Test</description>
			<link>something.html</link>
	  </item>
	</rdf:RDF>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].URL != "http://example.org/something.html" {
		t.Errorf("Incorrect entry url, got: %s", feed.Entries[0].URL)
	}
}

func TestParseItemWithoutLink(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>

	<rdf:RDF
	  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	  xmlns="http://purl.org/rss/1.0/"
	>

	  <channel rdf:about="http://meerkat.oreillynet.com/?_fl=rss1.0">
		<title>Meerkat</title>
		<link>http://meerkat.oreillynet.com</link>
	  </channel>

	  <item rdf:about="http://c.moreover.com/click/here.pl?r123">
		<title>Title</title>
		<description>Test</description>
	  </item>
	</rdf:RDF>`

	feed, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Error(err)
	}

	if feed.Entries[0].Hash != "37f5223ebd58639aa62a49afbb61df960efb7dc5db5181dfb3cedd9a49ad34c6" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[0].Hash)
	}

	if feed.Entries[0].URL != "http://meerkat.oreillynet.com" {
		t.Errorf("Incorrect entry url, got: %s", feed.Entries[0].URL)
	}
}

func TestParseInvalidXml(t *testing.T) {
	data := `garbage`
	_, err := Parse(bytes.NewBufferString(data))
	if err == nil {
		t.Error("Parse should returns an error")
	}

	if _, ok := err.(errors.LocalizedError); !ok {
		t.Error("The error returned must be a LocalizedError")
	}
}

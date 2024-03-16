// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rdf // import "miniflux.app/v2/internal/reader/rdf"

import (
	"bytes"
	"strings"
	"testing"
	"time"
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

	feed, err := Parse("http://xml.com/pub/rdf.xml", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "XML.com" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "http://xml.com/pub/rdf.xml" {
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

	feed, err := Parse("http://meerkat.oreillynet.com/feed.rdf", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "Meerkat" {
		t.Errorf("Incorrect title, got: %s", feed.Title)
	}

	if feed.FeedURL != "http://meerkat.oreillynet.com/feed.rdf" {
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

func TestParseRDFFeedWithEmptyTitle(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns="http://purl.org/rss/1.0/">
		<channel>
			<link>http://example.org/item</link>
		</channel>
		<item>
			<title>Example</title>
			<link>http://example.org/item</link>
			<description>Test</description>
		</item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org/feed", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "http://example.org/feed" {
		t.Errorf(`Incorrect title, got: %q`, feed.Title)
	}
}

func TestParseRDFFeedWithEmptyLink(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns="http://purl.org/rss/1.0/">
		<channel>
			<title>Example Feed</title>
		</channel>
		<item>
			<title>Example</title>
			<link>http://example.org/item</link>
			<description>Test</description>
		</item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org/feed", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.SiteURL != "http://example.org/feed" {
		t.Errorf(`Incorrect SiteURL, got: %q`, feed.SiteURL)
	}

	if feed.FeedURL != "http://example.org/feed" {
		t.Errorf(`Incorrect FeedURL, got: %q`, feed.FeedURL)
	}
}

func TestParseRDFFeedWithRelativeLink(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns="http://purl.org/rss/1.0/">
		<channel>
			<title>Example Feed</title>
			<link>/test/index.html</link>
		</channel>
		<item>
			<title>Example</title>
			<link>http://example.org/item</link>
			<description>Test</description>
		</item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org/feed", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.SiteURL != "http://example.org/test/index.html" {
		t.Errorf(`Incorrect SiteURL, got: %q`, feed.SiteURL)
	}

	if feed.FeedURL != "http://example.org/feed" {
		t.Errorf(`Incorrect FeedURL, got: %q`, feed.FeedURL)
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

	feed, err := Parse("http://meerkat.oreillynet.com", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Hash != "37f5223ebd58639aa62a49afbb61df960efb7dc5db5181dfb3cedd9a49ad34c6" {
		t.Errorf("Incorrect entry hash, got: %s", feed.Entries[0].Hash)
	}

	if feed.Entries[0].URL != "http://meerkat.oreillynet.com" {
		t.Errorf("Incorrect entry url, got: %s", feed.Entries[0].URL)
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

	feed, err := Parse("http://meerkat.oreillynet.com", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].URL != "http://example.org/something.html" {
		t.Errorf("Incorrect entry url, got: %s", feed.Entries[0].URL)
	}
}

func TestParseFeedWithURLWrappedInSpaces(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF xmlns:admin="http://webns.net/mvcb/" xmlns="http://purl.org/rss/1.0/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns:prism="http://purl.org/rss/1.0/modules/prism/" xmlns:taxo="http://purl.org/rss/1.0/modules/taxonomy/" xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:syn="http://purl.org/rss/1.0/modules/syndication/">
	<channel rdf:about="http://biorxiv.org">
		<title>bioRxiv Subject Collection: Bioengineering</title>
		<link>http://biorxiv.org</link>
		<description>
			This feed contains articles for bioRxiv Subject Collection "Bioengineering"
		</description>
		<items>
			<rdf:Seq>
				<rdf:li rdf:resource="http://biorxiv.org/cgi/content/short/857789v1?rss=1"/>
			</rdf:Seq>
		</items>
		<prism:eIssn/>
		<prism:publicationName>bioRxiv</prism:publicationName>
		<prism:issn/>
		<image rdf:resource=""/>
	</channel>
	<image rdf:about="">
		<title>bioRxiv</title>
		<url/>
		<link>http://biorxiv.org</link>
	</image>
	<item rdf:about="http://biorxiv.org/cgi/content/short/857789v1?rss=1">
		<title>
			<![CDATA[
			Microscale Collagen and Fibroblast Interactions Enhance Primary Human Hepatocyte Functions in 3-Dimensional Models
			]]>
		</title>
		<link>
			http://biorxiv.org/cgi/content/short/857789v1?rss=1
		</link>
		<description><![CDATA[
		Human liver models that are 3-dimensional (3D) in architecture are proving to be indispensable for diverse applications, including compound metabolism and toxicity screening during preclinical drug development, to model human liver diseases for the discovery of novel therapeutics, and for cell-based therapies in the clinic; however, further development of such models is needed to maintain high levels of primary human hepatocyte (PHH) functions for weeks to months in vitro. Therefore, here we determined how microscale 3D collagen-I presentation and fibroblast interaction could affect the long-term functions of PHHs. High-throughput droplet microfluidics was utilized to rapidly generate reproducibly-sized (~300 micron diameter) microtissues containing PHHs encapsulated in collagen-I +/- supportive fibroblasts, namely 3T3-J2 murine embryonic fibroblasts or primary human hepatic stellate cells (HSCs); self-assembled spheroids and bulk collagen gels (macrogels) containing PHHs served as gold-standard controls. Hepatic functions (e.g. albumin and cytochrome-P450 or CYP activities) and gene expression were subsequently measured for up to 6 weeks. We found that collagen-based 3D microtissues rescued PHH functions within static multi-well plates at 2- to 30-fold higher levels than self-assembled spheroids or macrogels. Further coating of PHH microtissues with 3T3-J2s led to higher hepatic functions than when the two cell types were either coencapsulated together or when HSCs were used for the coating instead. Additionally, the 3T3-J2-coated PHH microtissues displayed 6+ weeks of relatively stable hepatic gene expression and function at levels similar to freshly thawed PHHs. Lastly, microtissues responded in a clinically-relevant manner to drug-mediated CYP induction or hepatotoxicity. In conclusion, fibroblast-coated collagen microtissues containing PHHs display hepatic functions for 6+ weeks without any fluid perfusion at higher levels than spheroids and macrogels, and such microtissues can be used to assess drug-mediated CYP induction and hepatotoxicity. Ultimately, microtissues may find broader utility for modeling liver diseases and as building blocks for cell-based therapies.
		]]></description>
		<dc:creator><![CDATA[ Kukla, D., Crampton, A., Wood, D., Khetani, S. ]]></dc:creator>
		<dc:date>2019-11-29</dc:date>
		<dc:identifier>doi:10.1101/857789</dc:identifier>
		<dc:title><![CDATA[Microscale Collagen and Fibroblast Interactions Enhance Primary Human Hepatocyte Functions in 3-Dimensional Models]]></dc:title>
		<dc:publisher>Cold Spring Harbor Laboratory</dc:publisher>
		<prism:publicationDate>2019-11-29</prism:publicationDate>
		<prism:section></prism:section>
	</item>
	</rdf:RDF>`

	feed, err := Parse("http://biorxiv.org", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.SiteURL != "http://biorxiv.org" {
		t.Errorf(`Incorrect URL, got: %q`, feed.SiteURL)
	}

	if len(feed.Entries) != 1 {
		t.Fatalf(`Unexpected number of entries, got %d`, len(feed.Entries))
	}

	if feed.Entries[0].URL != `http://biorxiv.org/cgi/content/short/857789v1?rss=1` {
		t.Errorf(`Unexpected entry URL, got %q`, feed.Entries[0].URL)
	}
}

func TestParseRDFItemWitEmptyTitleElement(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns="http://purl.org/rss/1.0/">
		<channel>
			<title>Example Feed</title>
			<link>http://example.org/</link>
		</channel>
		<item>
			<title> </title>
			<link>http://example.org/item</link>
			<description>Test</description>
		</item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org/", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Fatalf(`Unexpected number of entries, got %d`, len(feed.Entries))
	}

	expected := `http://example.org/item`
	result := feed.Entries[0].Title
	if result != expected {
		t.Errorf(`Unexpected entry title, got %q instead of %q`, result, expected)
	}
}

func TestParseRDFItemWithDublinCoreTitleElement(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns="http://purl.org/rss/1.0/"
		xmlns:dc="http://purl.org/dc/elements/1.1/">
		<channel>
			<title>Example Feed</title>
			<link>http://example.org/</link>
		</channel>
		<item>
			<dc:title>Dublin Core Title</dc:title>
			<link>http://example.org/</link>
			<description>Test</description>
		</item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org/", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Fatalf(`Unexpected number of entries, got %d`, len(feed.Entries))
	}

	expected := `Dublin Core Title`
	result := feed.Entries[0].Title
	if result != expected {
		t.Errorf(`Unexpected entry title, got %q instead of %q`, result, expected)
	}
}

func TestParseRDFItemWithDuplicateTitleElement(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns="http://purl.org/rss/1.0/"
		xmlns:dc="http://purl.org/dc/elements/1.1/">
		<channel>
			<title>Example Feed</title>
			<link>http://example.org/</link>
		</channel>
		<item>
			<title>Item Title</title>
			<dc:title/>
			<link>http://example.org/</link>
			<description>Test</description>
		</item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org/", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Fatalf(`Unexpected number of entries, got %d`, len(feed.Entries))
	}

	expected := `Item Title`
	result := feed.Entries[0].Title
	if result != expected {
		t.Errorf(`Unexpected entry title, got %q instead of %q`, result, expected)
	}
}

func TestParseItemWithEncodedHTMLTitle(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/">
	  <channel>
			<title>Example</title>
			<link>http://example.org</link>
	  </channel>

	  <item>
			<title>AT&amp;amp;T</title>
			<description>Test</description>
			<link>http://example.org/test.html</link>
	  </item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Title != `AT&T` {
		t.Errorf("Incorrect entry title, got: %q", feed.Entries[0].Title)
	}
}

func TestParseRDFWithContentEncoded(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns="http://purl.org/rss/1.0/"
		xmlns:content="http://purl.org/rss/1.0/modules/content/">
		<channel>
			<title>Example Feed</title>
			<link>http://example.org/</link>
		</channel>
		<item>
			<title>Item Title</title>
			<link>http://example.org/</link>
			<content:encoded><![CDATA[<p>Test</p>]]></content:encoded>
		</item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org/", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Fatalf(`Unexpected number of entries, got %d`, len(feed.Entries))
	}

	expected := `<p>Test</p>`
	result := feed.Entries[0].Content
	if result != expected {
		t.Errorf(`Unexpected entry content, got %q instead of %q`, result, expected)
	}
}

func TestParseRDFWithEncodedHTMLDescription(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF
		xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		xmlns="http://purl.org/rss/1.0/"
		xmlns:content="http://purl.org/rss/1.0/modules/content/">
		<channel>
			<title>Example Feed</title>
			<link>http://example.org/</link>
		</channel>
		<item>
			<title>Item Title</title>
			<link>http://example.org/</link>
			<description>AT&amp;amp;T &lt;img src="https://example.org/img.png"&gt;&lt;/a&gt;</description>
		</item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org/", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Fatalf(`Unexpected number of entries, got %d`, len(feed.Entries))
	}

	expected := `AT&amp;T <img src="https://example.org/img.png"></a>`
	result := feed.Entries[0].Content
	if result != expected {
		t.Errorf(`Unexpected entry content, got %v instead of %v`, result, expected)
	}
}

func TestParseItemWithoutDate(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/">
	  <channel>
			<title>Example</title>
			<link>http://example.org</link>
	  </channel>

	  <item>
			<title>Title</title>
			<description>Test</description>
			<link>http://example.org/test.html</link>
	  </item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	expectedDate := time.Now().In(time.Local)
	diff := expectedDate.Sub(feed.Entries[0].Date)
	if diff > time.Second {
		t.Errorf("Incorrect entry date, got: %v", diff)
	}
}

func TestParseItemWithDublicCoreDate(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:slash="http://purl.org/rss/1.0/modules/slash/">
	  <channel>
			<title>Example</title>
			<link>http://example.org</link>
	  </channel>

	  <item>
			<title>Title</title>
			<description>Test</description>
			<link>http://example.org/test.html</link>
			<dc:creator>Tester</dc:creator>
			<dc:date>2018-04-10T05:00:00+00:00</dc:date>
	  </item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	expectedDate := time.Date(2018, time.April, 10, 5, 0, 0, 0, time.UTC)
	if !feed.Entries[0].Date.Equal(expectedDate) {
		t.Errorf("Incorrect entry date, got: %v, want: %v", feed.Entries[0].Date, expectedDate)
	}
}

func TestParseItemWithInvalidDublicCoreDate(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:slash="http://purl.org/rss/1.0/modules/slash/">
	  <channel>
			<title>Example</title>
			<link>http://example.org</link>
	  </channel>

	  <item>
			<title>Title</title>
			<description>Test</description>
			<link>http://example.org/test.html</link>
			<dc:creator>Tester</dc:creator>
			<dc:date>20-04-10T05:00:00+00:00</dc:date>
	  </item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	expectedDate := time.Now().In(time.Local)
	diff := expectedDate.Sub(feed.Entries[0].Date)
	if diff > time.Second {
		t.Errorf("Incorrect entry date, got: %v", diff)
	}
}

func TestParseItemWithEncodedHTMLInDCCreatorField(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:slash="http://purl.org/rss/1.0/modules/slash/">
	  <channel>
			<title>Example</title>
			<link>http://example.org</link>
	  </channel>

	  <item>
			<title>Title</title>
			<description>Test</description>
			<link>http://example.org/test.html</link>
			<dc:creator>&lt;a href=&quot;http://example.org/author1&quot;>Author 1&lt;/a&gt; (University 1), &lt;a href=&quot;http://example.org/author2&quot;>Author 2&lt;/a&gt; (University 2)</dc:creator>
			<dc:date>2018-04-10T05:00:00+00:00</dc:date>
	  </item>
	</rdf:RDF>`

	feed, err := Parse("http://example.org", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	expectedAuthor := "Author 1 (University 1), Author 2 (University 2)"
	if feed.Entries[0].Author != expectedAuthor {
		t.Errorf("Incorrect entry author, got: %s, want: %s", feed.Entries[0].Author, expectedAuthor)
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

	feed, err := Parse("http://meerkat.oreillynet.com", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Entries[0].Author != "Rael Dornfest (mailto:rael@oreilly.com)" {
		t.Errorf("Incorrect entry author, got: %s", feed.Entries[0].Author)
	}
}

func TestParseInvalidXml(t *testing.T) {
	data := `garbage`
	_, err := Parse("http://example.org", bytes.NewReader([]byte(data)))
	if err == nil {
		t.Fatal("Parse should returns an error")
	}
}

func TestParseFeedWithHTMLEntity(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/">
	  <channel>
			<title>Example &nbsp; Feed</title>
			<link>http://example.org</link>
	  </channel>
	</rdf:RDF>`

	feed, err := Parse("http://example.org", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.Title != "Example \u00a0 Feed" {
		t.Errorf(`Incorrect title, got: %q`, feed.Title)
	}
}

func TestParseFeedWithInvalidCharacterEntity(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/">
	  <channel>
			<title>Example Feed</title>
			<link>http://example.org/a&b</link>
	  </channel>
	</rdf:RDF>`

	feed, err := Parse("http://example.org", bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}

	if feed.SiteURL != "http://example.org/a&b" {
		t.Errorf(`Incorrect URL, got: %q`, feed.SiteURL)
	}
}

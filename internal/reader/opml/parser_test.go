// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/v2/internal/reader/opml"

import (
	"bytes"
	"testing"
)

func TestParseOpmlWithoutCategories(t *testing.T) {
	data := `<?xml version="1.0" encoding="ISO-8859-1"?>
	<opml version="2.0">
		<head>
			<title>mySubscriptions.opml</title>
		</head>
		<body>
			<outline text="CNET News.com" description="Tech news and business reports by CNET News.com. Focused on information technology, core topics include computers, hardware, software, networking, and Internet media." htmlUrl="http://news.com.com/" language="unknown" title="CNET News.com" type="rss" version="RSS2" xmlUrl="http://news.com.com/2547-1_3-0-5.xml"/>
			<outline text="washingtonpost.com - Politics" description="Politics" htmlUrl="http://www.washingtonpost.com/wp-dyn/politics?nav=rss_politics" language="unknown" title="washingtonpost.com - Politics" type="rss" version="RSS2" xmlUrl="http://www.washingtonpost.com/wp-srv/politics/rssheadlines.xml"/>
			<outline text="Scobleizer: Microsoft Geek Blogger" description="Robert Scoble's look at geek and Microsoft life." htmlUrl="http://radio.weblogs.com/0001011/" language="unknown" title="Scobleizer: Microsoft Geek Blogger" type="rss" version="RSS2" xmlUrl="http://radio.weblogs.com/0001011/rss.xml"/>
			<outline text="Yahoo! News: Technology" description="Technology" htmlUrl="http://news.yahoo.com/news?tmpl=index&amp;cid=738" language="unknown" title="Yahoo! News: Technology" type="rss" version="RSS2" xmlUrl="http://rss.news.yahoo.com/rss/tech"/>
			<outline text="Workbench" description="Programming and publishing news and comment" htmlUrl="http://www.cadenhead.org/workbench/" language="unknown" title="Workbench" type="rss" version="RSS2" xmlUrl="http://www.cadenhead.org/workbench/rss.xml"/>
			<outline text="Christian Science Monitor | Top Stories" description="Read the front page stories of csmonitor.com." htmlUrl="http://csmonitor.com" language="unknown" title="Christian Science Monitor | Top Stories" type="rss" version="RSS" xmlUrl="http://www.csmonitor.com/rss/top.rss"/>
			<outline text="Dictionary.com Word of the Day" description="A new word is presented every day with its definition and example sentences from actual published works." htmlUrl="http://dictionary.reference.com/wordoftheday/" language="unknown" title="Dictionary.com Word of the Day" type="rss" version="RSS" xmlUrl="http://www.dictionary.com/wordoftheday/wotd.rss"/>
			<outline text="The Motley Fool" description="To Educate, Amuse, and Enrich" htmlUrl="http://www.fool.com" language="unknown" title="The Motley Fool" type="rss" version="RSS" xmlUrl="http://www.fool.com/xml/foolnews_rss091.xml"/>
			<outline text="InfoWorld: Top News" description="The latest on Top News from InfoWorld" htmlUrl="http://www.infoworld.com/news/index.html" language="unknown" title="InfoWorld: Top News" type="rss" version="RSS2" xmlUrl="http://www.infoworld.com/rss/news.xml"/>
			<outline text="NYT &gt; Business" description="Find breaking news &amp; business news on Wall Street, media &amp; advertising, international business, banking, interest rates, the stock market, currencies &amp; funds." htmlUrl="http://www.nytimes.com/pages/business/index.html?partner=rssnyt" language="unknown" title="NYT &gt; Business" type="rss" version="RSS2" xmlUrl="http://www.nytimes.com/services/xml/rss/nyt/Business.xml"/>
			<outline text="NYT &gt; Technology" description="" htmlUrl="http://www.nytimes.com/pages/technology/index.html?partner=rssnyt" language="unknown" title="NYT &gt; Technology" type="rss" version="RSS2" xmlUrl="http://www.nytimes.com/services/xml/rss/nyt/Technology.xml"/>
			<outline text="Scripting News" description="It's even worse than it appears." htmlUrl="http://www.scripting.com/" language="unknown" title="Scripting News" type="rss" version="RSS2" xmlUrl="http://www.scripting.com/rss.xml"/>
			<outline text="Wired News" description="Technology, and the way we do business, is changing the world we know. Wired News is a technology - and business-oriented news service feeding an intelligent, discerning audience. What role does technology play in the day-to-day living of your life? Wired News tells you. How has evolving technology changed the face of the international business world? Wired News puts you in the picture." htmlUrl="http://www.wired.com/" language="unknown" title="Wired News" type="rss" version="RSS" xmlUrl="http://www.wired.com/news_drop/netcenter/netcenter.rdf"/>
		</body>
	</opml>
	`

	var expected SubcriptionList
	expected = append(expected, &Subcription{Title: "CNET News.com", FeedURL: "http://news.com.com/2547-1_3-0-5.xml", SiteURL: "http://news.com.com/"})

	subscriptions, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) != 13 {
		t.Fatalf("Wrong number of subscriptions: %d instead of %d", len(subscriptions), 13)
	}

	if !subscriptions[0].Equals(expected[0]) {
		t.Errorf(`Subscription is different: "%v" vs "%v"`, subscriptions[0], expected[0])
	}
}

func TestParseOpmlWithCategories(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<opml version="2.0">
		<head>
			<title>mySubscriptions.opml</title>
		</head>
		<body>
			<outline text="My Category 1">
				<outline text="Feed 1" xmlUrl="http://example.org/feed1/" htmlUrl="http://example.org/1"/>
				<outline text="Feed 2" xmlUrl="http://example.org/feed2/" htmlUrl="http://example.org/2"/>
			</outline>
			<outline text="My Category 2">
			<outline text="Feed 3" xmlUrl="http://example.org/feed3/" htmlUrl="http://example.org/3"/>
		</outline>
		</body>
	</opml>
	`

	var expected SubcriptionList
	expected = append(expected, &Subcription{Title: "Feed 1", FeedURL: "http://example.org/feed1/", SiteURL: "http://example.org/1", CategoryName: "My Category 1"})
	expected = append(expected, &Subcription{Title: "Feed 2", FeedURL: "http://example.org/feed2/", SiteURL: "http://example.org/2", CategoryName: "My Category 1"})
	expected = append(expected, &Subcription{Title: "Feed 3", FeedURL: "http://example.org/feed3/", SiteURL: "http://example.org/3", CategoryName: "My Category 2"})

	subscriptions, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) != 3 {
		t.Fatalf("Wrong number of subscriptions: %d instead of %d", len(subscriptions), 3)
	}

	for i := range len(subscriptions) {
		if !subscriptions[i].Equals(expected[i]) {
			t.Errorf(`Subscription is different: "%v" vs "%v"`, subscriptions[i], expected[i])
		}
	}
}

func TestParseOpmlWithEmptyTitleAndEmptySiteURL(t *testing.T) {
	data := `<?xml version="1.0" encoding="ISO-8859-1"?>
	<opml version="2.0">
	<head>
	<title>mySubscriptions.opml</title>
	</head>
	<body>
		<outline xmlUrl="http://example.org/feed1/" htmlUrl="http://example.org/1"/>
		<outline xmlUrl="http://example.org/feed2/"/>
	</body>
	</opml>
	`

	var expected SubcriptionList
	expected = append(expected, &Subcription{Title: "http://example.org/1", FeedURL: "http://example.org/feed1/", SiteURL: "http://example.org/1", CategoryName: ""})
	expected = append(expected, &Subcription{Title: "http://example.org/feed2/", FeedURL: "http://example.org/feed2/", SiteURL: "http://example.org/feed2/", CategoryName: ""})

	subscriptions, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) != 2 {
		t.Fatalf("Wrong number of subscriptions: %d instead of %d", len(subscriptions), 2)
	}

	for i := range len(subscriptions) {
		if !subscriptions[i].Equals(expected[i]) {
			t.Errorf(`Subscription is different: "%v" vs "%v"`, subscriptions[i], expected[i])
		}
	}
}

func TestParseOpmlVersion1(t *testing.T) {
	data := `<?xml version="1.0"?>
	<opml version="1.0">
		<head>
			<title>mySubscriptions.opml</title>
			<dateCreated>Wed, 13 Mar 2019 11:51:41 GMT</dateCreated>
		</head>
		<body>
			<outline title="Category 1">
				<outline type="rss" title="Feed 1" xmlUrl="http://example.org/feed1/" htmlUrl="http://example.org/1"></outline>
			</outline>
			<outline title="Category 2">
				<outline type="rss" title="Feed 2" xmlUrl="http://example.org/feed2/" htmlUrl="http://example.org/2"></outline>
			</outline>
		</body>
	</opml>
	`

	var expected SubcriptionList
	expected = append(expected, &Subcription{Title: "Feed 1", FeedURL: "http://example.org/feed1/", SiteURL: "http://example.org/1", CategoryName: "Category 1"})
	expected = append(expected, &Subcription{Title: "Feed 2", FeedURL: "http://example.org/feed2/", SiteURL: "http://example.org/2", CategoryName: "Category 2"})

	subscriptions, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) != 2 {
		t.Fatalf("Wrong number of subscriptions: %d instead of %d", len(subscriptions), 2)
	}

	for i := range len(subscriptions) {
		if !subscriptions[i].Equals(expected[i]) {
			t.Errorf(`Subscription is different: "%v" vs "%v"`, subscriptions[i], expected[i])
		}
	}
}

func TestParseOpmlVersion1WithoutOuterOutline(t *testing.T) {
	data := `<?xml version="1.0"?>
	<opml version="1.0">
		<head>
			<title>mySubscriptions.opml</title>
			<dateCreated>Wed, 13 Mar 2019 11:51:41 GMT</dateCreated>
		</head>
		<body>
			<outline type="rss" title="Feed 1" xmlUrl="http://example.org/feed1/" htmlUrl="http://example.org/1"></outline>
			<outline type="rss" title="Feed 2" xmlUrl="http://example.org/feed2/" htmlUrl="http://example.org/2"></outline>
		</body>
	</opml>
	`

	var expected SubcriptionList
	expected = append(expected, &Subcription{Title: "Feed 1", FeedURL: "http://example.org/feed1/", SiteURL: "http://example.org/1", CategoryName: ""})
	expected = append(expected, &Subcription{Title: "Feed 2", FeedURL: "http://example.org/feed2/", SiteURL: "http://example.org/2", CategoryName: ""})

	subscriptions, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) != 2 {
		t.Fatalf("Wrong number of subscriptions: %d instead of %d", len(subscriptions), 2)
	}

	for i := range len(subscriptions) {
		if !subscriptions[i].Equals(expected[i]) {
			t.Errorf(`Subscription is different: "%v" vs "%v"`, subscriptions[i], expected[i])
		}
	}
}

func TestParseOpmlVersion1WithSeveralNestedOutlines(t *testing.T) {
	data := `<?xml version="1.0"?>
	<opml xmlns:rssowl="http://www.rssowl.org" version="1.1">
		<head>
			<title>RSSOwl Subscriptions</title>
			<dateCreated>星期二, 26 四月 2022 00:12:04 CST</dateCreated>
		</head>
		<body>
			<outline text="My Feeds" rssowl:isSet="true" rssowl:id="7">
				<outline text="Some Category" rssowl:isSet="false" rssowl:id="55">
					<outline type="rss" title="Feed 1" xmlUrl="http://example.org/feed1/" htmlUrl="http://example.org/1"></outline>
					<outline type="rss" title="Feed 2" xmlUrl="http://example.org/feed2/" htmlUrl="http://example.org/2"></outline>
				</outline>
				<outline text="Another Category" rssowl:isSet="false" rssowl:id="87">
					<outline type="rss" title="Feed 3" xmlUrl="http://example.org/feed3/" htmlUrl="http://example.org/3"></outline>
				</outline>
			</outline>
		</body>
	</opml>
	`

	var expected SubcriptionList
	expected = append(expected, &Subcription{Title: "Feed 1", FeedURL: "http://example.org/feed1/", SiteURL: "http://example.org/1", CategoryName: "Some Category"})
	expected = append(expected, &Subcription{Title: "Feed 2", FeedURL: "http://example.org/feed2/", SiteURL: "http://example.org/2", CategoryName: "Some Category"})
	expected = append(expected, &Subcription{Title: "Feed 3", FeedURL: "http://example.org/feed3/", SiteURL: "http://example.org/3", CategoryName: "Another Category"})

	subscriptions, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) != 3 {
		t.Fatalf("Wrong number of subscriptions: %d instead of %d", len(subscriptions), 3)
	}

	for i := range len(subscriptions) {
		if !subscriptions[i].Equals(expected[i]) {
			t.Errorf(`Subscription is different: "%v" vs "%v"`, subscriptions[i], expected[i])
		}
	}
}

func TestParseOpmlWithInvalidCharacterEntity(t *testing.T) {
	data := `<?xml version="1.0"?>
	<opml version="1.0">
		<head>
			<title>mySubscriptions.opml</title>
		</head>
		<body>
			<outline title="Feed 1">
				<outline type="rss" title="Feed 1" xmlUrl="http://example.org/feed1/a&b" htmlUrl="http://example.org/c&d"></outline>
			</outline>
		</body>
	</opml>
	`

	var expected SubcriptionList
	expected = append(expected, &Subcription{Title: "Feed 1", FeedURL: "http://example.org/feed1/a&b", SiteURL: "http://example.org/c&d", CategoryName: "Feed 1"})

	subscriptions, err := Parse(bytes.NewBufferString(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) != 1 {
		t.Fatalf("Wrong number of subscriptions: %d instead of %d", len(subscriptions), 1)
	}

	for i := range len(subscriptions) {
		if !subscriptions[i].Equals(expected[i]) {
			t.Errorf(`Subscription is different: "%v" vs "%v"`, subscriptions[i], expected[i])
		}
	}
}

func TestParseInvalidXML(t *testing.T) {
	data := `garbage`
	_, err := Parse(bytes.NewBufferString(data))
	if err == nil {
		t.Error("Parse should generate an error")
	}
}

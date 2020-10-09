// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package processor // import "miniflux.app/reader/processor"

import (
	"miniflux.app/reader/parser"
	"testing"
)

func TestKeeplistRules(t *testing.T) {
	data := `<?xml version="1.0"?>
	<rss version="2.0">
	<channel>
		<title>SomeGood News</title>
		<link>http://foo.bar/</link>
		<item>
			<title>Kitten News</title>
			<link>http://kitties.today/daily-kitten</link>
			<description>Kitten picture of the day.</description>
			<pubDate>Tue, 03 Jun 2003 09:39:21 GMT</pubDate>
			<guid>http://kitties.today</guid>
		</item>
		<item>
			<title>Daily Covid DoomScrolling News</title>
			<link>http://covid.doom/daily-panic-dose</link>
			<description>Did you know that you can get COVID IN YOUR DREAMS?.</description>
			<pubDate>Tue, 03 Jun 2020 09:39:21 GMT</pubDate>
			<guid>http://covid.doom</guid>
		</item>
	</channel>
	</rss>`

	feed, err := parser.ParseFeed(data)
	if err != nil {
		t.Error(err)
	}
	if len(feed.Entries) != 2 {
		t.Errorf("Error parsing feed")
	}

	//case insensitive
	feed.KeeplistRules = "(?i)kitten"
	filterFeedEntries(feed)
	if len(feed.Entries) != 1 {
		t.Errorf("Keeplist filter rule did not properly filter the feed")
	}
}

func TestBlocklistRules(t *testing.T) {
	data := `<?xml version="1.0"?>
	<rss version="2.0">
	<channel>
		<title>SomeGood News</title>
		<link>http://foo.bar/</link>
		<item>
			<title>Kitten News</title>
			<link>http://kitties.today/daily-kitten</link>
			<description>Kitten picture of the day.</description>
			<pubDate>Tue, 03 Jun 2003 09:39:21 GMT</pubDate>
			<guid>http://kitties.today</guid>
		</item>
		<item>
			<title>Daily Covid DoomScrolling News</title>
			<link>http://covid.doom/daily-panic-dose</link>
			<description>Did you know that you can get COVID IN YOUR DREAMS?.</description>
			<pubDate>Tue, 03 Jun 2020 09:39:21 GMT</pubDate>
			<guid>http://covid.doom</guid>
		</item>
	</channel>
	</rss>`

	feed, err := parser.ParseFeed(data)
	if err != nil {
		t.Error(err)
	}
	if len(feed.Entries) != 2 {
		t.Errorf("Error parsing feed")
	}

	//case insensitive
	feed.BlocklistRules = "(?i)covid"
	filterFeedEntries(feed)
	if len(feed.Entries) != 1 {
		t.Errorf("Keeplist filter rule did not properly filter the feed")
	}
}

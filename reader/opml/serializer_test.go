// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSerialize(t *testing.T) {
	var subscriptions SubcriptionList
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 1", FeedURL: "http://example.org/feed/1", SiteURL: "http://example.org/1", CategoryName: "Category 1"})
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 2", FeedURL: "http://example.org/feed/2", SiteURL: "http://example.org/2", CategoryName: "Category 1"})
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 3", FeedURL: "http://example.org/feed/3", SiteURL: "http://example.org/3", CategoryName: "Category 2"})

	output := Serialize(subscriptions)
	fmt.Println(output)
	feeds, err := Parse(bytes.NewBufferString(output))
	if err != nil {
		t.Error(err)
	}

	if len(feeds) != 3 {
		t.Errorf("Wrong number of subscriptions: %d instead of %d", len(feeds), 3)
	}

	found := false
	for _, feed := range feeds {
		if feed.Title == "Feed 1" && feed.CategoryName == "Category 1" &&
			feed.FeedURL == "http://example.org/feed/1" && feed.SiteURL == "http://example.org/1" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Serialized feed is incorrect")
	}
}

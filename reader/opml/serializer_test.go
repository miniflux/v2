// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml

import "testing"
import "bytes"

func TestSerialize(t *testing.T) {
	var subscriptions SubcriptionList
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 1", FeedURL: "http://example.org/feed/1", SiteURL: "http://example.org/1", CategoryName: "Category 1"})
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 2", FeedURL: "http://example.org/feed/2", SiteURL: "http://example.org/2", CategoryName: "Category 1"})
	subscriptions = append(subscriptions, &Subcription{Title: "Feed 3", FeedURL: "http://example.org/feed/3", SiteURL: "http://example.org/3", CategoryName: "Category 2"})

	output := Serialize(subscriptions)
	feeds, err := Parse(bytes.NewBufferString(output))
	if err != nil {
		t.Error(err)
	}

	if len(feeds) != 3 {
		t.Errorf("Wrong number of subscriptions: %d instead of %d", len(feeds), 3)
	}

	for i := 0; i < len(feeds); i++ {
		if !feeds[i].Equals(subscriptions[i]) {
			t.Errorf(`Subscription are different: "%v" vs "%v"`, subscriptions[i], feeds[i])
		}
	}
}

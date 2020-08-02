// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package subscription

import "testing"

func TestFindYoutubeChannelFeed(t *testing.T) {
	scenarios := map[string]string{
		"https://www.youtube.com/channel/UC-Qj80avWItNRjkZ41rzHyw": "https://www.youtube.com/feeds/videos.xml?channel_id=UC-Qj80avWItNRjkZ41rzHyw",
		"http://example.org/feed":                                  "http://example.org/feed",
	}

	for websiteURL, expectedFeedURL := range scenarios {
		result := findYoutubeChannelFeed(websiteURL)
		if result != expectedFeedURL {
			t.Errorf(`Unexpected Feed, got %s, instead of %s`, result, expectedFeedURL)
		}
	}
}

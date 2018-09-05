// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// +build integration

package tests

import (
	"testing"
)

func TestDiscoverSubscriptions(t *testing.T) {
	client := createClient(t)
	subscriptions, err := client.Discover(testWebsiteURL)
	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) != 1 {
		t.Fatalf(`Invalid number of subscriptions, got "%v" instead of "%v"`, len(subscriptions), 2)
	}

	if subscriptions[0].Title != testFeedTitle {
		t.Fatalf(`Invalid feed title, got "%v" instead of "%v"`, subscriptions[0].Title, testFeedTitle)
	}

	if subscriptions[0].Type != "atom" {
		t.Fatalf(`Invalid feed type, got "%v" instead of "%v"`, subscriptions[0].Type, "atom")
	}

	if subscriptions[0].URL != testFeedURL {
		t.Fatalf(`Invalid feed URL, got "%v" instead of "%v"`, subscriptions[0].URL, testFeedURL)
	}
}

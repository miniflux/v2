// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package tests

import (
	"testing"

	miniflux "miniflux.app/client"
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

	if subscriptions[0].Title != testSubscriptionTitle {
		t.Fatalf(`Invalid feed title, got "%v" instead of "%v"`, subscriptions[0].Title, testSubscriptionTitle)
	}

	if subscriptions[0].Type != "atom" {
		t.Fatalf(`Invalid feed type, got "%v" instead of "%v"`, subscriptions[0].Type, "atom")
	}

	if subscriptions[0].URL != testFeedURL {
		t.Fatalf(`Invalid feed URL, got "%v" instead of "%v"`, subscriptions[0].URL, testFeedURL)
	}
}

func TestDiscoverSubscriptionsWithInvalidURL(t *testing.T) {
	client := createClient(t)
	_, err := client.Discover("invalid")
	if err == nil {
		t.Fatal(`Invalid URLs should be rejected`)
	}
}

func TestDiscoverSubscriptionsWithNoSubscription(t *testing.T) {
	client := createClient(t)
	_, err := client.Discover(testBaseURL)
	if err != miniflux.ErrNotFound {
		t.Fatal(`A 404 should be returned when there is no subscription`)
	}
}

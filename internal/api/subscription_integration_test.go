// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"errors"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

func TestDiscoverSubscriptionsEndpoint(t *testing.T) {
	t.Parallel()

	fixture := servertest.NewFeedFixture(t)

	client := newAdminClient(t)
	subscriptions, err := client.Discover(fixture.WebsiteURL)
	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) == 0 {
		t.Fatalf(`Invalid number of subscriptions, got %d`, len(subscriptions))
	}

	if subscriptions[0].Title != servertest.SubscriptionTitle {
		t.Fatalf(`Invalid title, got %q`, subscriptions[0].Title)
	}

	if subscriptions[0].URL != fixture.FeedURL {
		t.Fatalf(`Invalid URL, got %q`, subscriptions[0].URL)
	}
}

func TestDiscoverSubscriptionsWithInvalidURL(t *testing.T) {
	t.Parallel()

	client := newAdminClient(t)
	_, err := client.Discover("invalid_url")
	if !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Discovering subscriptions with an invalid URL should raise an error, got %v`, err)
	}
}

func TestDiscoverSubscriptionsWithNoSubscription(t *testing.T) {
	t.Parallel()

	client := newAdminClient(t)
	if _, err := client.Discover(servertest.Require(t).URL); !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Discovering subscriptions with no subscription should raise a "not found" error, got %v`, err)
	}
}

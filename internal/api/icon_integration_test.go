// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"errors"
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestGetFeedIcon(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	feedID, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: fixture.feedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	icon, err := regularUserClient.FeedIcon(feedID)
	if err != nil {
		t.Fatal(err)
	}

	if icon == nil {
		t.Fatalf(`Invalid icon, got nil`)
	}

	if icon.MimeType != "image/png" {
		t.Fatalf(`Invalid mime type, got %q`, icon.MimeType)
	}

	if len(icon.Data) == 0 {
		t.Fatalf(`Invalid data, got empty`)
	}

	icon, err = regularUserClient.Icon(icon.ID)
	if err != nil {
		t.Fatal(err)
	}

	if icon == nil {
		t.Fatalf(`Invalid icon, got nil`)
	}

	if icon.MimeType != "image/png" {
		t.Fatalf(`Invalid mime type, got %q`, icon.MimeType)
	}

	if len(icon.Data) == 0 {
		t.Fatalf(`Invalid data, got empty`)
	}
}

func TestGetFeedIconWithInexistingFeedID(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	client := testConfig.newAdminClient()
	_, err := client.FeedIcon(123456789)
	if !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Fetching the icon of an inexisting feed should raise an error, got %v`, err)
	}
}

func TestGetIconWithInexistingIconID(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)

	client := testConfig.newAdminClient()
	_, err := client.Icon(123456789)
	if !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Fetching an inexisting icon should return a "not found" error, got %v`, err)
	}
}

func TestGetIconByIconIDFromAnotherUser(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	// The owner subscribes to a feed, which fetches and stores its icon.
	ownerClient, _ := testConfig.newRegularUser(t)

	feedID, err := ownerClient.CreateFeed(&miniflux.FeedCreationRequest{
		FeedURL: fixture.feedURL,
	})
	if err != nil {
		t.Fatal(err)
	}

	ownerIcon, err := ownerClient.FeedIcon(feedID)
	if err != nil {
		t.Fatal(err)
	}
	if ownerIcon == nil {
		t.Fatalf(`Invalid icon, got nil`)
	}

	// The owner can fetch its own icon by icon ID.
	if _, err := ownerClient.Icon(ownerIcon.ID); err != nil {
		t.Fatalf(`The owner should be able to fetch its own icon, got %v`, err)
	}

	// Another user without access to that feed must not be able to fetch the icon.
	otherClient, _ := testConfig.newRegularUser(t)

	if _, err := otherClient.Icon(ownerIcon.ID); !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Fetching an icon owned by another user should return a "not found" error, got %v`, err)
	}
}

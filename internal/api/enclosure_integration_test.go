// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"errors"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

func TestUpdateEnclosureEndpoint(t *testing.T) {
	t.Parallel()

	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	result := listFeedEntries(t, regularUserClient, feedID, nil)

	var enclosure *miniflux.Enclosure
	for _, entry := range result.Entries {
		if len(entry.Enclosures) > 0 {
			enclosure = entry.Enclosures[0]
			break
		}
	}

	if enclosure == nil {
		t.Fatal(`The second fixture entry should have an enclosure`)
	}

	err := regularUserClient.UpdateEnclosure(enclosure.ID, &miniflux.EnclosureUpdateRequest{
		MediaProgression: 20,
	})
	if err != nil {
		t.Fatal(err)
	}

	updatedEnclosure, err := regularUserClient.Enclosure(enclosure.ID)
	if err != nil {
		t.Fatal(err)
	}

	if updatedEnclosure.MediaProgression != 20 {
		t.Fatalf(`Failed to update media_progression, expected %d but got %d`, 20, updatedEnclosure.MediaProgression)
	}
}

func TestGetEnclosureEndpoint(t *testing.T) {
	t.Parallel()

	fixture := servertest.NewFeedFixture(t)

	regularUserClient, _ := newRegularUser(t)

	feedID := fixture.Subscribe(t, regularUserClient)

	result := listFeedEntries(t, regularUserClient, feedID, nil)

	var expectedEnclosure *miniflux.Enclosure
	for _, entry := range result.Entries {
		if len(entry.Enclosures) > 0 {
			expectedEnclosure = entry.Enclosures[0]
			break
		}
	}

	if expectedEnclosure == nil {
		t.Fatal(`The second fixture entry should have an enclosure`)
	}

	enclosure, err := regularUserClient.Enclosure(expectedEnclosure.ID)
	if err != nil {
		t.Fatal(err)
	}

	if enclosure.ID != expectedEnclosure.ID {
		t.Fatalf(`Invalid enclosureID, got %d while expecting %d`, enclosure.ID, expectedEnclosure.ID)
	}

	if _, err = regularUserClient.Enclosure(99999); !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Fetching an inexisting enclosure should raise an error, got %v`, err)
	}
}

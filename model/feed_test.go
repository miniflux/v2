// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import "testing"

func TestMergeFeedTitle(t *testing.T) {
	feed1 := &Feed{Title: "Feed 1"}
	feed2 := &Feed{Title: "Feed 2"}
	feed1.Merge(feed2)

	if feed1.Title != "Feed 2" {
		t.Fatal(`The title of feed1 should be merged`)
	}

	feed1 = &Feed{Title: "Feed 1"}
	feed2 = &Feed{}
	feed1.Merge(feed2)

	if feed1.Title != "Feed 1" {
		t.Fatal(`The title of feed1 should not be merged`)
	}

	feed1 = &Feed{Title: "Feed 1"}
	feed2 = &Feed{Title: "Feed 1"}
	feed1.Merge(feed2)

	if feed1.Title != "Feed 1" {
		t.Fatal(`The title of feed1 should not be changed`)
	}
}

func TestMergeFeedCategory(t *testing.T) {
	feed1 := &Feed{Category: &Category{ID: 222}}
	feed2 := &Feed{Category: &Category{ID: 333}}
	feed1.Merge(feed2)

	if feed1.Category.ID != 333 {
		t.Fatal(`The category of feed1 should be merged`)
	}

	feed1 = &Feed{Category: &Category{ID: 222}}
	feed2 = &Feed{}
	feed1.Merge(feed2)

	if feed1.Category.ID != 222 {
		t.Fatal(`The category of feed1 should not be merged`)
	}

	feed1 = &Feed{Category: &Category{ID: 222}}
	feed2 = &Feed{Category: &Category{ID: 0}}
	feed1.Merge(feed2)

	if feed1.Category.ID != 222 {
		t.Fatal(`The category of feed1 should not be merged`)
	}
}

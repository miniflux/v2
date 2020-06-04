// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

import "testing"

func TestValidateEntryStatus(t *testing.T) {
	for _, status := range []string{EntryStatusRead, EntryStatusUnread, EntryStatusMarked, EntryStatusRemoved} {
		if err := ValidateEntryStatus(status); err != nil {
			t.Error(`A valid status should not generate any error`)
		}
	}

	if err := ValidateEntryStatus("invalid"); err == nil {
		t.Error(`An invalid status should generate a error`)
	}
}

func TestValidateEntryOrder(t *testing.T) {
	for _, status := range []string{"id", "status", "changed_at", "published_at", "category_title", "category_id"} {
		if err := ValidateEntryOrder(status); err != nil {
			t.Error(`A valid order should not generate any error`)
		}
	}

	if err := ValidateEntryOrder("invalid"); err == nil {
		t.Error(`An invalid order should generate a error`)
	}
}

func TestValidateEntryDirection(t *testing.T) {
	for _, status := range []string{"asc", "desc"} {
		if err := ValidateDirection(status); err != nil {
			t.Error(`A valid direction should not generate any error`)
		}
	}

	if err := ValidateDirection("invalid"); err == nil {
		t.Error(`An invalid direction should generate a error`)
	}
}

func TestValidateRange(t *testing.T) {
	if err := ValidateRange(-1, 0); err == nil {
		t.Error(`An invalid offset should generate a error`)
	}

	if err := ValidateRange(0, -1); err == nil {
		t.Error(`An invalid limit should generate a error`)
	}

	if err := ValidateRange(42, 42); err != nil {
		t.Error(`A valid offset and limit should not generate any error`)
	}
}

func TestGetOppositeDirection(t *testing.T) {
	if OppositeDirection("asc") != "desc" {
		t.Errorf(`The opposite direction of "asc" should be "desc"`)
	}

	if OppositeDirection("desc") != "asc" {
		t.Errorf(`The opposite direction of "desc" should be "asc"`)
	}

	if OppositeDirection("invalid") != "asc" {
		t.Errorf(`An invalid direction should return "asc"`)
	}
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"reflect"
	"testing"
)

func TestEntryUpdateRequestPatchTags(t *testing.T) {
	entry := &Entry{Tags: []string{"original"}}
	(&EntryUpdateRequest{}).Patch(entry)
	if !reflect.DeepEqual(entry.Tags, []string{"original"}) {
		t.Fatalf(`Expected omitted tags to leave entry tags unchanged, got %v`, entry.Tags)
	}

	emptyTags := []string{}
	(&EntryUpdateRequest{Tags: &emptyTags}).Patch(entry)
	if len(entry.Tags) != 0 {
		t.Fatalf(`Expected empty tags to clear entry tags, got %v`, entry.Tags)
	}

	newTags := []string{"foo", "bar"}
	(&EntryUpdateRequest{Tags: &newTags}).Patch(entry)
	if !reflect.DeepEqual(entry.Tags, newTags) {
		t.Fatalf(`Expected tags to be replaced with %v, got %v`, newTags, entry.Tags)
	}
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestValidateEntriesStatusUpdateRequest(t *testing.T) {
	err := ValidateEntriesStatusUpdateRequest(&model.EntriesStatusUpdateRequest{
		Status:   model.EntryStatusRead,
		EntryIDs: []int64{int64(123), int64(456)},
	})
	if err != nil {
		t.Error(`A valid request should not be rejected`)
	}

	err = ValidateEntriesStatusUpdateRequest(&model.EntriesStatusUpdateRequest{
		Status: model.EntryStatusRead,
	})
	if err == nil {
		t.Error(`An empty list of entries is not valid`)
	}

	err = ValidateEntriesStatusUpdateRequest(&model.EntriesStatusUpdateRequest{
		Status:   "invalid",
		EntryIDs: []int64{int64(123)},
	})
	if err == nil {
		t.Error(`Only a valid status should be accepted`)
	}
}

func TestValidateEntryStatus(t *testing.T) {
	for _, status := range []string{model.EntryStatusRead, model.EntryStatusUnread, model.EntryStatusRemoved} {
		if err := ValidateEntryStatus(status); err != nil {
			t.Error(`A valid status should not generate any error`)
		}
	}

	if err := ValidateEntryStatus("invalid"); err == nil {
		t.Error(`An invalid status should generate a error`)
	}
}

func TestValidateEntryOrder(t *testing.T) {
	for _, status := range []string{"id", "status", "changed_at", "published_at", "created_at", "category_title", "category_id"} {
		if err := ValidateEntryOrder(status); err != nil {
			t.Error(`A valid order should not generate any error`)
		}
	}

	if err := ValidateEntryOrder("invalid"); err == nil {
		t.Error(`An invalid order should generate a error`)
	}
}

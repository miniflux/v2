// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/validator"

import (
	"fmt"

	"miniflux.app/model"
)

// ValidateEntriesStatusUpdateRequest validates a status update for a list of entries.
func ValidateEntriesStatusUpdateRequest(request *model.EntriesStatusUpdateRequest) error {
	if len(request.EntryIDs) == 0 {
		return fmt.Errorf(`The list of entries cannot be empty`)
	}

	return ValidateEntryStatus(request.Status)
}

// ValidateEntryStatus makes sure the entry status is valid.
func ValidateEntryStatus(status string) error {
	switch status {
	case model.EntryStatusRead, model.EntryStatusUnread, model.EntryStatusRemoved:
		return nil
	}

	return fmt.Errorf(`Invalid entry status, valid status values are: "%s", "%s" and "%s"`, model.EntryStatusRead, model.EntryStatusUnread, model.EntryStatusRemoved)
}

// ValidateEntryOrder makes sure the sorting order is valid.
func ValidateEntryOrder(order string) error {
	switch order {
	case "id", "status", "changed_at", "published_at", "created_at", "category_title", "category_id", "title", "author":
		return nil
	}

	return fmt.Errorf(`Invalid entry order, valid order values are: "id", "status", "changed_at", "published_at", "created_at", "category_title", "category_id", "title", "author"`)
}

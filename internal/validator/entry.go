// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"fmt"

	"miniflux.app/v2/internal/model"
)

// ValidateEntriesStatusUpdateRequest validates a status update for a list of entries.
func ValidateEntriesStatusUpdateRequest(request *model.EntriesStatusUpdateRequest) error {
	if len(request.EntryIDs) == 0 {
		return fmt.Errorf(`the list of entries cannot be empty`)
	}

	return ValidateEntryStatus(request.Status)
}

// ValidateEntryStatus makes sure the entry status is valid.
func ValidateEntryStatus(status string) error {
	switch status {
	case model.EntryStatusRead, model.EntryStatusUnread, model.EntryStatusRemoved:
		return nil
	}

	return fmt.Errorf(`invalid entry status, valid status values are: "%s", "%s" and "%s"`, model.EntryStatusRead, model.EntryStatusUnread, model.EntryStatusRemoved)
}

// ValidateEntryOrder makes sure the sorting order is valid.
func ValidateEntryOrder(order string) error {
	switch order {
	case "id", "status", "changed_at", "published_at", "created_at", "category_title", "category_id", "title", "author":
		return nil
	}

	return fmt.Errorf(`invalid entry order, valid order values are: "id", "status", "changed_at", "published_at", "created_at", "category_title", "category_id", "title", "author"`)
}

// ValidateEntryModification makes sure the entry modification is valid.
func ValidateEntryModification(request *model.EntryUpdateRequest) error {
	if request.Title != nil && *request.Title == "" {
		return fmt.Errorf(`the entry title cannot be empty`)
	}

	if request.Content != nil && *request.Content == "" {
		return fmt.Errorf(`the entry content cannot be empty`)
	}

	return nil
}

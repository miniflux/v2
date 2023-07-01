// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/validator"

import (
	"miniflux.app/model"
	"miniflux.app/storage"
)

// ValidateFeedCreation validates feed creation.
func ValidateFeedCreation(store *storage.Storage, userID int64, request *model.FeedCreationRequest) *ValidationError {
	if request.FeedURL == "" || request.CategoryID <= 0 {
		return NewValidationError("error.feed_mandatory_fields")
	}

	if !IsValidURL(request.FeedURL) {
		return NewValidationError("error.invalid_feed_url")
	}

	if store.FeedURLExists(userID, request.FeedURL) {
		return NewValidationError("error.feed_already_exists")
	}

	if !store.CategoryIDExists(userID, request.CategoryID) {
		return NewValidationError("error.feed_category_not_found")
	}

	if !IsValidRegex(request.BlocklistRules) {
		return NewValidationError("error.feed_invalid_blocklist_rule")
	}

	if !IsValidRegex(request.KeeplistRules) {
		return NewValidationError("error.feed_invalid_keeplist_rule")
	}

	return nil
}

// ValidateFeedModification validates feed modification.
func ValidateFeedModification(store *storage.Storage, userID int64, request *model.FeedModificationRequest) *ValidationError {
	if request.FeedURL != nil {
		if *request.FeedURL == "" {
			return NewValidationError("error.feed_url_not_empty")
		}

		if !IsValidURL(*request.FeedURL) {
			return NewValidationError("error.invalid_feed_url")
		}
	}

	if request.SiteURL != nil {
		if *request.SiteURL == "" {
			return NewValidationError("error.site_url_not_empty")
		}

		if !IsValidURL(*request.SiteURL) {
			return NewValidationError("error.invalid_site_url")
		}
	}

	if request.Title != nil {
		if *request.Title == "" {
			return NewValidationError("error.feed_title_not_empty")
		}
	}

	if request.CategoryID != nil {
		if !store.CategoryIDExists(userID, *request.CategoryID) {
			return NewValidationError("error.feed_category_not_found")
		}
	}

	if request.BlocklistRules != nil {
		if !IsValidRegex(*request.BlocklistRules) {
			return NewValidationError("error.feed_invalid_blocklist_rule")
		}
	}

	if request.KeeplistRules != nil {
		if !IsValidRegex(*request.KeeplistRules) {
			return NewValidationError("error.feed_invalid_keeplist_rule")
		}
	}

	return nil
}

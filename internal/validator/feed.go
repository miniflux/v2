// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
)

// ValidateFeedCreation validates feed creation.
func ValidateFeedCreation(store *storage.Storage, userID int64, request *model.FeedCreationRequest) *locale.LocalizedError {
	if request.FeedURL == "" || request.CategoryID <= 0 {
		return locale.NewLocalizedError("error.feed_mandatory_fields")
	}

	if !IsValidURL(request.FeedURL) {
		return locale.NewLocalizedError("error.invalid_feed_url")
	}

	if store.FeedURLExists(userID, request.FeedURL) {
		return locale.NewLocalizedError("error.feed_already_exists")
	}

	if !store.CategoryIDExists(userID, request.CategoryID) {
		return locale.NewLocalizedError("error.feed_category_not_found")
	}

	if !IsValidRegex(request.BlocklistRules) {
		return locale.NewLocalizedError("error.feed_invalid_blocklist_rule")
	}

	if !IsValidRegex(request.KeeplistRules) {
		return locale.NewLocalizedError("error.feed_invalid_keeplist_rule")
	}

	return nil
}

// ValidateFeedModification validates feed modification.
func ValidateFeedModification(store *storage.Storage, userID, feedID int64, request *model.FeedModificationRequest) *locale.LocalizedError {
	if request.FeedURL != "" {
		if request.FeedURL == "" {
			return locale.NewLocalizedError("error.feed_url_not_empty")
		}

		if !IsValidURL(request.FeedURL) {
			return locale.NewLocalizedError("error.invalid_feed_url")
		}

		if store.AnotherFeedURLExists(userID, feedID, request.FeedURL) {
			return locale.NewLocalizedError("error.feed_already_exists")
		}
	}

	if request.SiteURL != "" {
		if request.SiteURL == "" {
			return locale.NewLocalizedError("error.site_url_not_empty")
		}

		if !IsValidURL(request.SiteURL) {
			return locale.NewLocalizedError("error.invalid_site_url")
		}
	}

	if request.Title != "" {
		if request.Title == "" {
			return locale.NewLocalizedError("error.feed_title_not_empty")
		}
	}

	if request.CategoryID > 0 {
		if !store.CategoryIDExists(userID, request.CategoryID) {
			return locale.NewLocalizedError("error.feed_category_not_found")
		}
	}

	if request.BlocklistRules != "" {
		if !IsValidRegex(request.BlocklistRules) {
			return locale.NewLocalizedError("error.feed_invalid_blocklist_rule")
		}
	}

	if request.KeeplistRules != "" {
		if !IsValidRegex(request.KeeplistRules) {
			return locale.NewLocalizedError("error.feed_invalid_keeplist_rule")
		}
	}

	return nil
}

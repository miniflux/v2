// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"log/slog"

	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/urllib"
)

// ValidateFeedCreation validates feed creation.
func ValidateFeedCreation(store *storage.Storage, userID int64, request *model.FeedCreationRequest) *locale.LocalizedError {
	if request.FeedURL == "" || request.CategoryID <= 0 {
		return locale.NewLocalizedError("error.feed_mandatory_fields")
	}

	if !urllib.IsAbsoluteURL(request.FeedURL) {
		return locale.NewLocalizedError("error.invalid_feed_url")
	}

	if store.FeedURLExists(userID, request.FeedURL) {
		return locale.NewLocalizedError("error.feed_already_exists")
	}

	categoryExists, err := store.CategoryIDExists(userID, request.CategoryID)
	if err != nil {
		// *locale.LocalizedError (unlike *locale.LocalizedErrorWrapper elsewhere
		// in this codebase) has no way to carry an underlying error, so a
		// genuine backend failure here can't be distinguished from "category
		// not found" in the value returned to the caller. Log it so the
		// failure is at least observable, and fall back to the existing
		// user-facing message rather than changing this function's public
		// return type as part of this PR.
		slog.Error("validator: unable to check if feed category exists",
			slog.Int64("user_id", userID),
			slog.Int64("category_id", request.CategoryID),
			slog.Any("error", err),
		)
		return locale.NewLocalizedError("error.feed_category_not_found")
	}
	if !categoryExists {
		return locale.NewLocalizedError("error.feed_category_not_found")
	}

	if !IsValidRegex(request.BlocklistRules) {
		return locale.NewLocalizedError("error.feed_invalid_blocklist_rule")
	}

	if !IsValidRegex(request.KeeplistRules) {
		return locale.NewLocalizedError("error.feed_invalid_keeplist_rule")
	}

	if request.BlockFilterEntryRules != "" {
		if err := IsValidFilterRules(request.BlockFilterEntryRules, "block"); err != nil {
			return err
		}
	}

	if request.KeepFilterEntryRules != "" {
		if err := IsValidFilterRules(request.KeepFilterEntryRules, "keep"); err != nil {
			return err
		}
	}

	if request.ProxyURL != "" && !urllib.IsValidProxyURL(request.ProxyURL) {
		return locale.NewLocalizedError("error.invalid_feed_proxy_url")
	}

	return nil
}

// ValidateFeedModification validates feed modification.
func ValidateFeedModification(store *storage.Storage, userID, feedID int64, request *model.FeedModificationRequest) *locale.LocalizedError {
	if request.FeedURL != nil {
		if *request.FeedURL == "" {
			return locale.NewLocalizedError("error.feed_url_not_empty")
		}

		if !urllib.IsAbsoluteURL(*request.FeedURL) {
			return locale.NewLocalizedError("error.invalid_feed_url")
		}

		if store.AnotherFeedURLExists(userID, feedID, *request.FeedURL) {
			return locale.NewLocalizedError("error.feed_already_exists")
		}
	}

	if request.SiteURL != nil {
		if *request.SiteURL == "" {
			return locale.NewLocalizedError("error.site_url_not_empty")
		}

		if !urllib.IsAbsoluteURL(*request.SiteURL) {
			return locale.NewLocalizedError("error.invalid_site_url")
		}
	}

	if request.Title != nil {
		if *request.Title == "" {
			return locale.NewLocalizedError("error.feed_title_not_empty")
		}
	}

	if request.CategoryID != nil {
		categoryExists, err := store.CategoryIDExists(userID, *request.CategoryID)
		if err != nil {
			// See the identical rationale in ValidateFeedCreation above.
			slog.Error("validator: unable to check if feed category exists",
				slog.Int64("user_id", userID),
				slog.Int64("category_id", *request.CategoryID),
				slog.Any("error", err),
			)
			return locale.NewLocalizedError("error.feed_category_not_found")
		}
		if !categoryExists {
			return locale.NewLocalizedError("error.feed_category_not_found")
		}
	}

	if request.BlocklistRules != nil {
		if !IsValidRegex(*request.BlocklistRules) {
			return locale.NewLocalizedError("error.feed_invalid_blocklist_rule")
		}
	}

	if request.KeeplistRules != nil {
		if !IsValidRegex(*request.KeeplistRules) {
			return locale.NewLocalizedError("error.feed_invalid_keeplist_rule")
		}
	}

	if request.BlockFilterEntryRules != nil && *request.BlockFilterEntryRules != "" {
		if err := IsValidFilterRules(*request.BlockFilterEntryRules, "block"); err != nil {
			return err
		}
	}

	if request.KeepFilterEntryRules != nil && *request.KeepFilterEntryRules != "" {
		if err := IsValidFilterRules(*request.KeepFilterEntryRules, "keep"); err != nil {
			return err
		}
	}

	if request.ProxyURL != nil && *request.ProxyURL != "" {
		if !urllib.IsValidProxyURL(*request.ProxyURL) {
			return locale.NewLocalizedError("error.invalid_feed_proxy_url")
		}
	}

	return nil
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
)

// ValidateCategoryCreation validates category creation.
func ValidateCategoryCreation(store *storage.Storage, userID int64, request *model.CategoryRequest) *locale.LocalizedError {
	if request.Title == "" {
		return locale.NewLocalizedError("error.title_required")
	}

	if store.CategoryTitleExists(userID, request.Title) {
		return locale.NewLocalizedError("error.category_already_exists")
	}

	return nil
}

// ValidateCategoryModification validates category modification.
func ValidateCategoryModification(store *storage.Storage, userID, categoryID int64, request *model.CategoryRequest) *locale.LocalizedError {
	if request.Title == "" {
		return locale.NewLocalizedError("error.title_required")
	}

	if store.AnotherCategoryExists(userID, categoryID, request.Title) {
		return locale.NewLocalizedError("error.category_already_exists")
	}

	return nil
}

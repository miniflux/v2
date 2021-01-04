// Copyright 2021 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package validator // import "miniflux.app/validator"

import (
	"miniflux.app/model"
	"miniflux.app/storage"
)

// ValidateCategoryCreation validates category creation.
func ValidateCategoryCreation(store *storage.Storage, userID int64, request *model.CategoryRequest) *ValidationError {
	if request.Title == "" {
		return NewValidationError("error.title_required")
	}

	if store.CategoryTitleExists(userID, request.Title) {
		return NewValidationError("error.category_already_exists")
	}

	return nil
}

// ValidateCategoryModification validates category modification.
func ValidateCategoryModification(store *storage.Storage, userID, categoryID int64, request *model.CategoryRequest) *ValidationError {
	if request.Title == "" {
		return NewValidationError("error.title_required")
	}

	if store.AnotherCategoryExists(userID, categoryID, request.Title) {
		return NewValidationError("error.category_already_exists")
	}

	return nil
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/view"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	categoryID := request.RouteInt64Param(r, "categoryID")
	category, err := h.store.Category(request.UserID(r), categoryID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if category == nil {
		response.HTMLNotFound(w, r)
		return
	}

	categoryForm := form.NewCategoryForm(r)

	view := view.New(h.tpl, r)
	view.Set("form", categoryForm)
	view.Set("category", category)
	view.Set("menu", "categories")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	categoryRequest := &model.CategoryModificationRequest{
		Title:        new(categoryForm.Title),
		HideGlobally: new(categoryForm.HideGlobally),
	}

	if validationErr := validator.ValidateCategoryModification(h.store, user.ID, category.ID, categoryRequest); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(user.Language))
		response.HTML(w, r, view.Render("edit_category"))
		return
	}

	categoryRequest.Patch(category)
	if err := h.store.UpdateCategory(category); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	response.HTMLRedirect(w, r, h.routePath("/category/%d/feeds", categoryID))
}

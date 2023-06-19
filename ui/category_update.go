// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
	"miniflux.app/validator"
)

func (h *handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	loggedUser, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	categoryID := request.RouteInt64Param(r, "categoryID")
	category, err := h.store.Category(request.UserID(r), categoryID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if category == nil {
		html.NotFound(w, r)
		return
	}

	categoryForm := form.NewCategoryForm(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", categoryForm)
	view.Set("category", category)
	view.Set("menu", "categories")
	view.Set("user", loggedUser)
	view.Set("countUnread", h.store.CountUnreadEntries(loggedUser.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(loggedUser.ID))

	categoryRequest := &model.CategoryRequest{
		Title:        categoryForm.Title,
		HideGlobally: categoryForm.HideGlobally,
	}

	if validationErr := validator.ValidateCategoryModification(h.store, loggedUser.ID, category.ID, categoryRequest); validationErr != nil {
		view.Set("errorMessage", validationErr.TranslationKey)
		html.OK(w, r, view.Render("create_category"))
		return
	}

	categoryRequest.Patch(category)
	if err := h.store.UpdateCategory(category); err != nil {
		logger.Error("[UI:UpdateCategory] %v", err)
		view.Set("errorMessage", "error.unable_to_update_category")
		html.OK(w, r, view.Render("edit_category"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "categoryFeeds", "categoryID", categoryID))
}

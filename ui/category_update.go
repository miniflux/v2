// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
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
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	if err := categoryForm.Validate(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("edit_category"))
		return
	}

	if h.store.AnotherCategoryExists(user.ID, category.ID, categoryForm.Title) {
		view.Set("errorMessage", "error.category_already_exists")
		html.OK(w, r, view.Render("edit_category"))
		return
	}

	err = h.store.UpdateCategory(categoryForm.Merge(category))
	if err != nil {
		logger.Error("[UI:UpdateCategory] %v", err)
		view.Set("errorMessage", "error.unable_to_update_category")
		html.OK(w, r, view.Render("edit_category"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "categoryFeeds", "categoryID", categoryID))
}

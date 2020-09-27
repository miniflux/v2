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
	"miniflux.app/model"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) saveCategory(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	categoryForm := form.NewCategoryForm(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", categoryForm)
	view.Set("menu", "categories")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	if err := categoryForm.Validate(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("create_category"))
		return
	}

	duplicateCategory, err := h.store.CategoryByTitle(user.ID, categoryForm.Title)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if duplicateCategory != nil {
		view.Set("errorMessage", "error.category_already_exists")
		html.OK(w, r, view.Render("create_category"))
		return
	}

	category := model.Category{
		Title:  categoryForm.Title,
		UserID: user.ID,
	}

	if err = h.store.CreateCategory(&category); err != nil {
		logger.Error("[UI:SaveCategory] %v", err)
		view.Set("errorMessage", "error.unable_to_create_category")
		html.OK(w, r, view.Render("create_category"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "categories"))
}

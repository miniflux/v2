// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/http/request"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// SaveCategory validate and save the new category into the database.
func (c *Controller) SaveCategory(w http.ResponseWriter, r *http.Request) {
	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	categoryForm := form.NewCategoryForm(r)

	sess := session.New(c.store, request.SessionID(r))
	view := view.New(c.tpl, r, sess)
	view.Set("form", categoryForm)
	view.Set("menu", "categories")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))

	if err := categoryForm.Validate(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("create_category"))
		return
	}

	duplicateCategory, err := c.store.CategoryByTitle(user.ID, categoryForm.Title)
	if err != nil {
		html.ServerError(w, err)
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

	if err = c.store.CreateCategory(&category); err != nil {
		logger.Error("[Controller:CreateCategory] %v", err)
		view.Set("errorMessage", "error.unable_to_create_category")
		html.OK(w, r, view.Render("create_category"))
		return
	}

	response.Redirect(w, r, route.Path(c.router, "categories"))
}

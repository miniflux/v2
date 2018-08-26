// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/context"
	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// UpdateCategory validates and updates a category.
func (c *Controller) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)

	user, err := c.store.UserByID(ctx.UserID())
	if err != nil {
		html.ServerError(w, err)
		return
	}

	categoryID, err := request.IntParam(r, "categoryID")
	if err != nil {
		html.BadRequest(w, err)
		return
	}

	category, err := c.store.Category(ctx.UserID(), categoryID)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	if category == nil {
		html.NotFound(w)
		return
	}

	categoryForm := form.NewCategoryForm(r)

	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)
	view.Set("form", categoryForm)
	view.Set("category", category)
	view.Set("menu", "categories")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))

	if err := categoryForm.Validate(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("edit_category"))
		return
	}

	if c.store.AnotherCategoryExists(user.ID, category.ID, categoryForm.Title) {
		view.Set("errorMessage", "This category already exists.")
		html.OK(w, r, view.Render("edit_category"))
		return
	}

	err = c.store.UpdateCategory(categoryForm.Merge(category))
	if err != nil {
		logger.Error("[Controller:UpdateCategory] %v", err)
		view.Set("errorMessage", "Unable to update this category.")
		html.OK(w, r, view.Render("edit_category"))
		return
	}

	response.Redirect(w, r, route.Path(c.router, "categories"))
}

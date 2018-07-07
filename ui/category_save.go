// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/http/route"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/ui/form"
	"github.com/miniflux/miniflux/ui/session"
	"github.com/miniflux/miniflux/ui/view"
)

// SaveCategory validate and save the new category into the database.
func (c *Controller) SaveCategory(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)

	user, err := c.store.UserByID(ctx.UserID())
	if err != nil {
		html.ServerError(w, err)
		return
	}

	categoryForm := form.NewCategoryForm(r)

	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)
	view.Set("form", categoryForm)
	view.Set("menu", "categories")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))

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
		view.Set("errorMessage", "This category already exists.")
		html.OK(w, r, view.Render("create_category"))
		return
	}

	category := model.Category{
		Title:  categoryForm.Title,
		UserID: user.ID,
	}

	if err = c.store.CreateCategory(&category); err != nil {
		logger.Error("[Controller:CreateCategory] %v", err)
		view.Set("errorMessage", "Unable to create this category.")
		html.OK(w, r, view.Render("create_category"))
		return
	}

	response.Redirect(w, r, route.Path(c.router, "categories"))
}

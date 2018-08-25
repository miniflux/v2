// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/context"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/model"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// CategoryEntries shows all entries for the given category.
func (c *Controller) CategoryEntries(w http.ResponseWriter, r *http.Request) {
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

	offset := request.QueryIntParam(r, "offset", 0)
	builder := c.store.NewEntryQueryBuilder(user.ID)
	builder.WithCategoryID(category.ID)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(user.EntryDirection)
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithOffset(offset)
	builder.WithLimit(nbItemsPerPage)

	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, err)
		return
	}

	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)
	view.Set("category", category)
	view.Set("total", count)
	view.Set("entries", entries)
	view.Set("pagination", c.getPagination(route.Path(c.router, "categoryEntries", "categoryID", category.ID), count, offset))
	view.Set("menu", "categories")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))
	view.Set("hasSaveEntry", c.store.HasSaveEntry(user.ID))

	html.OK(w, r, view.Render("category_entries"))
}

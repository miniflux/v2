// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/server/core"
)

// ShowStarredPage renders the page with all starred entries.
func (c *Controller) ShowStarredPage(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()
	offset := request.QueryIntegerParam("offset", 0)

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	builder := c.store.NewEntryQueryBuilder(user.ID)
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithStarred()
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(user.EntryDirection)
	builder.WithOffset(offset)
	builder.WithLimit(nbItemsPerPage)

	entries, err := builder.GetEntries()
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("starred", args.Merge(tplParams{
		"entries":    entries,
		"total":      count,
		"pagination": c.getPagination(ctx.Route("starred"), count, offset),
		"menu":       "starred",
	}))
}

// ToggleBookmark handles Ajax request to toggle bookmark value.
func (c *Controller) ToggleBookmark(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()
	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	if err := c.store.ToggleBookmark(user.ID, entryID); err != nil {
		logger.Error("[Controller:UpdateEntryStatus] %v", err)
		response.JSON().ServerError(nil)
		return
	}

	response.JSON().Standard("OK")
}

// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"errors"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/ui/payload"
	"log"
)

func (c *Controller) ShowFeedEntry(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()
	sortingDirection := model.DefaultSortingDirection

	entryID, err := request.GetIntegerParam("entryID")
	if err != nil {
		response.Html().BadRequest(err)
		return
	}

	feedID, err := request.GetIntegerParam("feedID")
	if err != nil {
		response.Html().BadRequest(err)
		return
	}

	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	if entry == nil {
		response.Html().NotFound()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithFeedID(feedID)
	builder.WithCondition("e.id", "!=", entryID)
	builder.WithCondition("e.published_at", "<=", entry.Date)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(model.DefaultSortingDirection)
	nextEntry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithFeedID(feedID)
	builder.WithCondition("e.id", "!=", entryID)
	builder.WithCondition("e.published_at", ">=", entry.Date)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(model.GetOppositeDirection(sortingDirection))
	prevEntry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = ctx.GetRoute("feedEntry", "feedID", feedID, "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = ctx.GetRoute("feedEntry", "feedID", feedID, "entryID", prevEntry.ID)
	}

	if entry.Status == model.EntryStatusUnread {
		err = c.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
		if err != nil {
			log.Println(err)
			response.Html().ServerError(nil)
			return
		}
	}

	response.Html().Render("entry", args.Merge(tplParams{
		"entry":          entry,
		"prevEntry":      prevEntry,
		"nextEntry":      nextEntry,
		"nextEntryRoute": nextEntryRoute,
		"prevEntryRoute": prevEntryRoute,
		"menu":           "feeds",
	}))
}

func (c *Controller) ShowCategoryEntry(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()
	sortingDirection := model.DefaultSortingDirection

	categoryID, err := request.GetIntegerParam("categoryID")
	if err != nil {
		response.Html().BadRequest(err)
		return
	}

	entryID, err := request.GetIntegerParam("entryID")
	if err != nil {
		response.Html().BadRequest(err)
		return
	}

	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithCategoryID(categoryID)
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	if entry == nil {
		response.Html().NotFound()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithCategoryID(categoryID)
	builder.WithCondition("e.id", "!=", entryID)
	builder.WithCondition("e.published_at", "<=", entry.Date)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(sortingDirection)
	nextEntry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithCategoryID(categoryID)
	builder.WithCondition("e.id", "!=", entryID)
	builder.WithCondition("e.published_at", ">=", entry.Date)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(model.GetOppositeDirection(sortingDirection))
	prevEntry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = ctx.GetRoute("categoryEntry", "categoryID", categoryID, "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = ctx.GetRoute("categoryEntry", "categoryID", categoryID, "entryID", prevEntry.ID)
	}

	if entry.Status == model.EntryStatusUnread {
		err = c.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
		if err != nil {
			log.Println(err)
			response.Html().ServerError(nil)
			return
		}
	}

	response.Html().Render("entry", args.Merge(tplParams{
		"entry":          entry,
		"prevEntry":      prevEntry,
		"nextEntry":      nextEntry,
		"nextEntryRoute": nextEntryRoute,
		"prevEntryRoute": prevEntryRoute,
		"menu":           "categories",
	}))
}

func (c *Controller) ShowUnreadEntry(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()
	sortingDirection := model.DefaultSortingDirection

	entryID, err := request.GetIntegerParam("entryID")
	if err != nil {
		response.Html().BadRequest(err)
		return
	}

	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	if entry == nil {
		response.Html().NotFound()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithCondition("e.id", "!=", entryID)
	builder.WithCondition("e.published_at", "<=", entry.Date)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(sortingDirection)
	nextEntry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithCondition("e.id", "!=", entryID)
	builder.WithCondition("e.published_at", ">=", entry.Date)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(model.GetOppositeDirection(sortingDirection))
	prevEntry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = ctx.GetRoute("unreadEntry", "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = ctx.GetRoute("unreadEntry", "entryID", prevEntry.ID)
	}

	if entry.Status == model.EntryStatusUnread {
		err = c.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
		if err != nil {
			log.Println(err)
			response.Html().ServerError(nil)
			return
		}
	}

	response.Html().Render("entry", args.Merge(tplParams{
		"entry":          entry,
		"prevEntry":      prevEntry,
		"nextEntry":      nextEntry,
		"nextEntryRoute": nextEntryRoute,
		"prevEntryRoute": prevEntryRoute,
		"menu":           "unread",
	}))
}

func (c *Controller) ShowReadEntry(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()
	sortingDirection := model.DefaultSortingDirection

	entryID, err := request.GetIntegerParam("entryID")
	if err != nil {
		response.Html().BadRequest(err)
		return
	}

	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	if entry == nil {
		response.Html().NotFound()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithStatus(model.EntryStatusRead)
	builder.WithCondition("e.id", "!=", entryID)
	builder.WithCondition("e.published_at", "<=", entry.Date)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(sortingDirection)
	nextEntry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithStatus(model.EntryStatusRead)
	builder.WithCondition("e.id", "!=", entryID)
	builder.WithCondition("e.published_at", ">=", entry.Date)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(model.GetOppositeDirection(sortingDirection))
	prevEntry, err := builder.GetEntry()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = ctx.GetRoute("readEntry", "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = ctx.GetRoute("readEntry", "entryID", prevEntry.ID)
	}

	response.Html().Render("entry", args.Merge(tplParams{
		"entry":          entry,
		"prevEntry":      prevEntry,
		"nextEntry":      nextEntry,
		"nextEntryRoute": nextEntryRoute,
		"prevEntryRoute": prevEntryRoute,
		"menu":           "history",
	}))
}

func (c *Controller) UpdateEntriesStatus(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	entryIDs, status, err := payload.DecodeEntryStatusPayload(request.GetBody())
	if err != nil {
		log.Println(err)
		response.Json().BadRequest(nil)
		return
	}

	if len(entryIDs) == 0 {
		response.Html().BadRequest(errors.New("The list of entryID is empty"))
		return
	}

	err = c.store.SetEntriesStatus(user.ID, entryIDs, status)
	if err != nil {
		log.Println(err)
		response.Html().ServerError(nil)
		return
	}

	response.Json().Standard("OK")
}

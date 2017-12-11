// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"errors"
	"log"

	"github.com/miniflux/miniflux2/integration"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/scraper"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/ui/payload"
	"github.com/miniflux/miniflux2/storage"
)

// FetchContent downloads the original HTML page and returns relevant contents.
func (c *Controller) FetchContent(ctx *core.Context, request *core.Request, response *core.Response) {
	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	user := ctx.LoggedUser()
	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		response.JSON().ServerError(err)
		return
	}

	if entry == nil {
		response.JSON().NotFound(errors.New("Entry not found"))
		return
	}

	content, err := scraper.Fetch(entry.URL)
	if err != nil {
		response.JSON().ServerError(err)
		return
	}

	if len(content) > len(entry.Content) {
		entry.Content = content
		c.store.UpdateEntryContent(entry)
	} else {
		content = entry.Content
	}

	response.JSON().Created(map[string]string{"content": content})
}

// SaveEntry send the link to external services.
func (c *Controller) SaveEntry(ctx *core.Context, request *core.Request, response *core.Response) {
	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	user := ctx.LoggedUser()
	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		response.JSON().ServerError(err)
		return
	}

	if entry == nil {
		response.JSON().NotFound(errors.New("Entry not found"))
		return
	}

	settings, err := c.store.Integration(user.ID)
	if err != nil {
		response.JSON().ServerError(err)
		return
	}

	go func() {
		integration.SendEntry(entry, settings)
	}()

	response.JSON().Created(map[string]string{"message": "saved"})
}

// ShowFeedEntry shows a single feed entry in "feed" mode.
func (c *Controller) ShowFeedEntry(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithFeedID(feedID)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	if entry == nil {
		response.HTML().NotFound()
		return
	}

	if entry.Status == model.EntryStatusUnread {
		err = c.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
		if err != nil {
			log.Println(err)
			response.HTML().ServerError(nil)
			return
		}
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithFeedID(feedID)

	prevEntry, nextEntry, err := c.getEntryPrevNext(user, builder, entry.ID)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = ctx.Route("feedEntry", "feedID", feedID, "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = ctx.Route("feedEntry", "feedID", feedID, "entryID", prevEntry.ID)
	}

	response.HTML().Render("entry", args.Merge(tplParams{
		"entry":          entry,
		"prevEntry":      prevEntry,
		"nextEntry":      nextEntry,
		"nextEntryRoute": nextEntryRoute,
		"prevEntryRoute": prevEntryRoute,
		"menu":           "feeds",
	}))
}

// ShowCategoryEntry shows a single feed entry in "category" mode.
func (c *Controller) ShowCategoryEntry(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	categoryID, err := request.IntegerParam("categoryID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithCategoryID(categoryID)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	if entry == nil {
		response.HTML().NotFound()
		return
	}

	if entry.Status == model.EntryStatusUnread {
		err = c.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
		if err != nil {
			log.Println(err)
			response.HTML().ServerError(nil)
			return
		}
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithCategoryID(categoryID)

	prevEntry, nextEntry, err := c.getEntryPrevNext(user, builder, entry.ID)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = ctx.Route("categoryEntry", "categoryID", categoryID, "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = ctx.Route("categoryEntry", "categoryID", categoryID, "entryID", prevEntry.ID)
	}

	response.HTML().Render("entry", args.Merge(tplParams{
		"entry":          entry,
		"prevEntry":      prevEntry,
		"nextEntry":      nextEntry,
		"nextEntryRoute": nextEntryRoute,
		"prevEntryRoute": prevEntryRoute,
		"menu":           "categories",
	}))
}

// ShowUnreadEntry shows a single feed entry in "unread" mode.
func (c *Controller) ShowUnreadEntry(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	if entry == nil {
		response.HTML().NotFound()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithStatus(model.EntryStatusUnread)

	prevEntry, nextEntry, err := c.getEntryPrevNext(user, builder, entry.ID)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = ctx.Route("unreadEntry", "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = ctx.Route("unreadEntry", "entryID", prevEntry.ID)
	}

	// We change the status here, otherwise we cannot get the pagination for unread items.
	if entry.Status == model.EntryStatusUnread {
		err = c.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
		if err != nil {
			log.Println(err)
			response.HTML().ServerError(nil)
			return
		}
	}

	response.HTML().Render("entry", args.Merge(tplParams{
		"entry":          entry,
		"prevEntry":      prevEntry,
		"nextEntry":      nextEntry,
		"nextEntryRoute": nextEntryRoute,
		"prevEntryRoute": prevEntryRoute,
		"menu":           "unread",
	}))
}

// ShowReadEntry shows a single feed entry in "history" mode.
func (c *Controller) ShowReadEntry(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	entryID, err := request.IntegerParam("entryID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	if entry == nil {
		response.HTML().NotFound()
		return
	}

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	builder = c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithStatus(model.EntryStatusRead)

	prevEntry, nextEntry, err := c.getEntryPrevNext(user, builder, entry.ID)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = ctx.Route("readEntry", "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = ctx.Route("readEntry", "entryID", prevEntry.ID)
	}

	response.HTML().Render("entry", args.Merge(tplParams{
		"entry":          entry,
		"prevEntry":      prevEntry,
		"nextEntry":      nextEntry,
		"nextEntryRoute": nextEntryRoute,
		"prevEntryRoute": prevEntryRoute,
		"menu":           "history",
	}))
}

// UpdateEntriesStatus handles Ajax request to update the status for a list of entries.
func (c *Controller) UpdateEntriesStatus(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	entryIDs, status, err := payload.DecodeEntryStatusPayload(request.Body())
	if err != nil {
		log.Println(err)
		response.JSON().BadRequest(nil)
		return
	}

	if len(entryIDs) == 0 {
		response.JSON().BadRequest(errors.New("The list of entryID is empty"))
		return
	}

	err = c.store.SetEntriesStatus(user.ID, entryIDs, status)
	if err != nil {
		log.Println(err)
		response.JSON().ServerError(nil)
		return
	}

	response.JSON().Standard("OK")
}

func (c *Controller) getEntryPrevNext(user *model.User, builder *storage.EntryQueryBuilder, entryID int64) (prev *model.Entry, next *model.Entry, err error) {
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(user.EntryDirection)

	entries, err := builder.GetEntries()
	if err != nil {
		return nil, nil, err
	}

	n := len(entries)
	for i := 0; i < n; i++ {
		if entries[i].ID == entryID {
			if i-1 > 0 {
				prev = entries[i-1]
			}

			if i+1 < n {
				next = entries[i+1]
			}
		}
	}

	return prev, next, nil
}

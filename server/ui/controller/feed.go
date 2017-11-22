// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"errors"
	"log"

	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/ui/form"
)

// ShowFeedsPage shows the page with all subscriptions.
func (c *Controller) ShowFeedsPage(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	feeds, err := c.store.GetFeeds(user.ID)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	response.Html().Render("feeds", args.Merge(tplParams{
		"feeds": feeds,
		"total": len(feeds),
		"menu":  "feeds",
	}))
}

// ShowFeedEntries shows all entries for the given feed.
func (c *Controller) ShowFeedEntries(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()
	offset := request.GetQueryIntegerParam("offset", 0)

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	feed, err := c.getFeedFromURL(request, response, user)
	if err != nil {
		return
	}

	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithFeedID(feed.ID)
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(model.DefaultSortingDirection)
	builder.WithOffset(offset)
	builder.WithLimit(NbItemsPerPage)

	entries, err := builder.GetEntries()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	response.Html().Render("feed_entries", args.Merge(tplParams{
		"feed":       feed,
		"entries":    entries,
		"total":      count,
		"pagination": c.getPagination(ctx.GetRoute("feedEntries", "feedID", feed.ID), count, offset),
		"menu":       "feeds",
	}))
}

// EditFeed shows the form to modify a subscription.
func (c *Controller) EditFeed(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	feed, err := c.getFeedFromURL(request, response, user)
	if err != nil {
		return
	}

	args, err := c.getFeedFormTemplateArgs(ctx, user, feed, nil)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	response.Html().Render("edit_feed", args)
}

// UpdateFeed update a subscription and redirect to the feed entries page.
func (c *Controller) UpdateFeed(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	feed, err := c.getFeedFromURL(request, response, user)
	if err != nil {
		return
	}

	feedForm := form.NewFeedForm(request.GetRequest())
	args, err := c.getFeedFormTemplateArgs(ctx, user, feed, feedForm)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	if err := feedForm.ValidateModification(); err != nil {
		response.Html().Render("edit_feed", args.Merge(tplParams{
			"errorMessage": err.Error(),
		}))
		return
	}

	err = c.store.UpdateFeed(feedForm.Merge(feed))
	if err != nil {
		log.Println(err)
		response.Html().Render("edit_feed", args.Merge(tplParams{
			"errorMessage": "Unable to update this feed.",
		}))
		return
	}

	response.Redirect(ctx.GetRoute("feedEntries", "feedID", feed.ID))
}

// RemoveFeed delete a subscription from the database and redirect to the list of feeds page.
func (c *Controller) RemoveFeed(ctx *core.Context, request *core.Request, response *core.Response) {
	feedID, err := request.GetIntegerParam("feedID")
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	user := ctx.GetLoggedUser()
	if err := c.store.RemoveFeed(user.ID, feedID); err != nil {
		response.Html().ServerError(err)
		return
	}

	response.Redirect(ctx.GetRoute("feeds"))
}

// RefreshFeed refresh a subscription and redirect to the feed entries page.
func (c *Controller) RefreshFeed(ctx *core.Context, request *core.Request, response *core.Response) {
	feedID, err := request.GetIntegerParam("feedID")
	if err != nil {
		response.Html().BadRequest(err)
		return
	}

	user := ctx.GetLoggedUser()
	if err := c.feedHandler.RefreshFeed(user.ID, feedID); err != nil {
		log.Println("[UI:RefreshFeed]", err)
	}

	response.Redirect(ctx.GetRoute("feedEntries", "feedID", feedID))
}

func (c *Controller) getFeedFromURL(request *core.Request, response *core.Response, user *model.User) (*model.Feed, error) {
	feedID, err := request.GetIntegerParam("feedID")
	if err != nil {
		response.Html().BadRequest(err)
		return nil, err
	}

	feed, err := c.store.GetFeedById(user.ID, feedID)
	if err != nil {
		response.Html().ServerError(err)
		return nil, err
	}

	if feed == nil {
		response.Html().NotFound()
		return nil, errors.New("Feed not found")
	}

	return feed, nil
}

func (c *Controller) getFeedFormTemplateArgs(ctx *core.Context, user *model.User, feed *model.Feed, feedForm *form.FeedForm) (tplParams, error) {
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		return nil, err
	}

	categories, err := c.store.GetCategories(user.ID)
	if err != nil {
		return nil, err
	}

	if feedForm == nil {
		args["form"] = form.FeedForm{
			SiteURL:    feed.SiteURL,
			FeedURL:    feed.FeedURL,
			Title:      feed.Title,
			CategoryID: feed.Category.ID,
		}
	} else {
		args["form"] = feedForm
	}

	args["categories"] = categories
	args["feed"] = feed
	args["menu"] = "feeds"
	return args, nil
}

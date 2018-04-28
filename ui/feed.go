// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"errors"

	"github.com/miniflux/miniflux/http/handler"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/ui/form"
)

// RefreshAllFeeds refresh all feeds in the background for the current user.
func (c *Controller) RefreshAllFeeds(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()
	jobs, err := c.store.NewUserBatch(user.ID, c.store.CountFeeds(user.ID))
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	go func() {
		c.pool.Push(jobs)
	}()

	response.Redirect(ctx.Route("feeds"))
}

// ShowFeedsPage shows the page with all subscriptions.
func (c *Controller) ShowFeedsPage(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	feeds, err := c.store.Feeds(user.ID)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("feeds", ctx.UserLanguage(), args.Merge(tplParams{
		"feeds": feeds,
		"total": len(feeds),
		"menu":  "feeds",
	}))
}

// ShowFeedEntries shows all entries for the given feed.
func (c *Controller) ShowFeedEntries(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()
	offset := request.QueryIntegerParam("offset", 0)

	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	feed, err := c.getFeedFromURL(request, response, user)
	if err != nil {
		return
	}

	builder := c.store.NewEntryQueryBuilder(user.ID)
	builder.WithFeedID(feed.ID)
	builder.WithoutStatus(model.EntryStatusRemoved)
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

	response.HTML().Render("feed_entries", ctx.UserLanguage(), args.Merge(tplParams{
		"feed":       feed,
		"entries":    entries,
		"total":      count,
		"pagination": c.getPagination(ctx.Route("feedEntries", "feedID", feed.ID), count, offset),
		"menu":       "feeds",
	}))
}

// EditFeed shows the form to modify a subscription.
func (c *Controller) EditFeed(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()

	feed, err := c.getFeedFromURL(request, response, user)
	if err != nil {
		return
	}

	args, err := c.getFeedFormTemplateArgs(ctx, user, feed, nil)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("edit_feed", ctx.UserLanguage(), args)
}

// UpdateFeed update a subscription and redirect to the feed entries page.
func (c *Controller) UpdateFeed(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	user := ctx.LoggedUser()

	feed, err := c.getFeedFromURL(request, response, user)
	if err != nil {
		return
	}

	feedForm := form.NewFeedForm(request.Request())
	args, err := c.getFeedFormTemplateArgs(ctx, user, feed, feedForm)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	if err := feedForm.ValidateModification(); err != nil {
		response.HTML().Render("edit_feed", ctx.UserLanguage(), args.Merge(tplParams{
			"errorMessage": err.Error(),
		}))
		return
	}

	err = c.store.UpdateFeed(feedForm.Merge(feed))
	if err != nil {
		logger.Error("[Controller:EditFeed] %v", err)
		response.HTML().Render("edit_feed", ctx.UserLanguage(), args.Merge(tplParams{
			"errorMessage": "Unable to update this feed.",
		}))
		return
	}

	response.Redirect(ctx.Route("feedEntries", "feedID", feed.ID))
}

// RemoveFeed delete a subscription from the database and redirect to the list of feeds page.
func (c *Controller) RemoveFeed(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	user := ctx.LoggedUser()
	if err := c.store.RemoveFeed(user.ID, feedID); err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.Redirect(ctx.Route("feeds"))
}

// RefreshFeed refresh a subscription and redirect to the feed entries page.
func (c *Controller) RefreshFeed(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	user := ctx.LoggedUser()
	if err := c.feedHandler.RefreshFeed(user.ID, feedID); err != nil {
		logger.Error("[Controller:RefreshFeed] %v", err)
	}

	response.Redirect(ctx.Route("feedEntries", "feedID", feedID))
}

func (c *Controller) getFeedFromURL(request *handler.Request, response *handler.Response, user *model.User) (*model.Feed, error) {
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.HTML().BadRequest(err)
		return nil, err
	}

	feed, err := c.store.FeedByID(user.ID, feedID)
	if err != nil {
		response.HTML().ServerError(err)
		return nil, err
	}

	if feed == nil {
		response.HTML().NotFound()
		return nil, errors.New("Feed not found")
	}

	return feed, nil
}

func (c *Controller) getFeedFormTemplateArgs(ctx *handler.Context, user *model.User, feed *model.Feed, feedForm *form.FeedForm) (tplParams, error) {
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		return nil, err
	}

	categories, err := c.store.Categories(user.ID)
	if err != nil {
		return nil, err
	}

	if feedForm == nil {
		args["form"] = form.FeedForm{
			SiteURL:      feed.SiteURL,
			FeedURL:      feed.FeedURL,
			Title:        feed.Title,
			ScraperRules: feed.ScraperRules,
			RewriteRules: feed.RewriteRules,
			Crawler:      feed.Crawler,
			CategoryID:   feed.Category.ID,
		}
	} else {
		args["form"] = feedForm
	}

	args["categories"] = categories
	args["feed"] = feed
	args["menu"] = "feeds"
	return args, nil
}

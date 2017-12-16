// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/reader/subscription"
	"github.com/miniflux/miniflux/server/core"
	"github.com/miniflux/miniflux/server/ui/form"
)

// Bookmarklet prefill the form to add a subscription from the URL provided by the bookmarklet.
func (c *Controller) Bookmarklet(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()
	args, err := c.getSubscriptionFormTemplateArgs(ctx, user)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	bookmarkletURL := request.QueryStringParam("uri", "")
	response.HTML().Render("add_subscription", args.Merge(tplParams{
		"form": &form.SubscriptionForm{URL: bookmarkletURL},
	}))
}

// AddSubscription shows the form to add a new feed.
func (c *Controller) AddSubscription(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	args, err := c.getSubscriptionFormTemplateArgs(ctx, user)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	response.HTML().Render("add_subscription", args)
}

// SubmitSubscription try to find a feed from the URL provided by the user.
func (c *Controller) SubmitSubscription(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	args, err := c.getSubscriptionFormTemplateArgs(ctx, user)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	subscriptionForm := form.NewSubscriptionForm(request.Request())
	if err := subscriptionForm.Validate(); err != nil {
		response.HTML().Render("add_subscription", args.Merge(tplParams{
			"form":         subscriptionForm,
			"errorMessage": err.Error(),
		}))
		return
	}

	subscriptions, err := subscription.FindSubscriptions(subscriptionForm.URL)
	if err != nil {
		logger.Error("[Controller:SubmitSubscription] %v", err)
		response.HTML().Render("add_subscription", args.Merge(tplParams{
			"form":         subscriptionForm,
			"errorMessage": err,
		}))
		return
	}

	logger.Info("[UI:SubmitSubscription] %s", subscriptions)

	n := len(subscriptions)
	switch {
	case n == 0:
		response.HTML().Render("add_subscription", args.Merge(tplParams{
			"form":         subscriptionForm,
			"errorMessage": "Unable to find any subscription.",
		}))
	case n == 1:
		feed, err := c.feedHandler.CreateFeed(user.ID, subscriptionForm.CategoryID, subscriptions[0].URL, subscriptionForm.Crawler)
		if err != nil {
			response.HTML().Render("add_subscription", args.Merge(tplParams{
				"form":         subscriptionForm,
				"errorMessage": err,
			}))
			return
		}

		response.Redirect(ctx.Route("feedEntries", "feedID", feed.ID))
	case n > 1:
		response.HTML().Render("choose_subscription", args.Merge(tplParams{
			"categoryID":    subscriptionForm.CategoryID,
			"subscriptions": subscriptions,
		}))
	}
}

// ChooseSubscription shows a page to choose a subscription.
func (c *Controller) ChooseSubscription(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.LoggedUser()

	args, err := c.getSubscriptionFormTemplateArgs(ctx, user)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	subscriptionForm := form.NewSubscriptionForm(request.Request())
	if err := subscriptionForm.Validate(); err != nil {
		response.HTML().Render("add_subscription", args.Merge(tplParams{
			"form":         subscriptionForm,
			"errorMessage": err.Error(),
		}))
		return
	}

	feed, err := c.feedHandler.CreateFeed(user.ID, subscriptionForm.CategoryID, subscriptionForm.URL, subscriptionForm.Crawler)
	if err != nil {
		response.HTML().Render("add_subscription", args.Merge(tplParams{
			"form":         subscriptionForm,
			"errorMessage": err,
		}))
		return
	}

	response.Redirect(ctx.Route("feedEntries", "feedID", feed.ID))
}

func (c *Controller) getSubscriptionFormTemplateArgs(ctx *core.Context, user *model.User) (tplParams, error) {
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		return nil, err
	}

	categories, err := c.store.Categories(user.ID)
	if err != nil {
		return nil, err
	}

	args["categories"] = categories
	args["menu"] = "feeds"
	return args, nil
}

// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/subscription"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/ui/form"
	"log"
)

func (c *Controller) AddSubscription(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	args, err := c.getSubscriptionFormTemplateArgs(ctx, user)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	response.Html().Render("add_subscription", args)
}

func (c *Controller) SubmitSubscription(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	args, err := c.getSubscriptionFormTemplateArgs(ctx, user)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	subscriptionForm := form.NewSubscriptionForm(request.Request())
	if err := subscriptionForm.Validate(); err != nil {
		response.Html().Render("add_subscription", args.Merge(tplParams{
			"form":         subscriptionForm,
			"errorMessage": err.Error(),
		}))
		return
	}

	subscriptions, err := subscription.FindSubscriptions(subscriptionForm.URL)
	if err != nil {
		log.Println(err)
		response.Html().Render("add_subscription", args.Merge(tplParams{
			"form":         subscriptionForm,
			"errorMessage": err,
		}))
		return
	}

	log.Println("[UI:SubmitSubscription]", subscriptions)

	n := len(subscriptions)
	switch {
	case n == 0:
		response.Html().Render("add_subscription", args.Merge(tplParams{
			"form":         subscriptionForm,
			"errorMessage": "Unable to find any subscription.",
		}))
	case n == 1:
		feed, err := c.feedHandler.CreateFeed(user.ID, subscriptionForm.CategoryID, subscriptions[0].URL)
		if err != nil {
			response.Html().Render("add_subscription", args.Merge(tplParams{
				"form":         subscriptionForm,
				"errorMessage": err,
			}))
			return
		}

		response.Redirect(ctx.GetRoute("feedEntries", "feedID", feed.ID))
	case n > 1:
		response.Html().Render("choose_subscription", args.Merge(tplParams{
			"categoryID":    subscriptionForm.CategoryID,
			"subscriptions": subscriptions,
		}))
	}
}

func (c *Controller) ChooseSubscription(ctx *core.Context, request *core.Request, response *core.Response) {
	user := ctx.GetLoggedUser()

	args, err := c.getSubscriptionFormTemplateArgs(ctx, user)
	if err != nil {
		response.Html().ServerError(err)
		return
	}

	subscriptionForm := form.NewSubscriptionForm(request.Request())
	if err := subscriptionForm.Validate(); err != nil {
		response.Html().Render("add_subscription", args.Merge(tplParams{
			"form":         subscriptionForm,
			"errorMessage": err.Error(),
		}))
		return
	}

	feed, err := c.feedHandler.CreateFeed(user.ID, subscriptionForm.CategoryID, subscriptionForm.URL)
	if err != nil {
		response.Html().Render("add_subscription", args.Merge(tplParams{
			"form":         subscriptionForm,
			"errorMessage": err,
		}))
		return
	}

	response.Redirect(ctx.GetRoute("feedEntries", "feedID", feed.ID))
}

func (c *Controller) getSubscriptionFormTemplateArgs(ctx *core.Context, user *model.User) (tplParams, error) {
	args, err := c.getCommonTemplateArgs(ctx)
	if err != nil {
		return nil, err
	}

	categories, err := c.store.GetCategories(user.ID)
	if err != nil {
		return nil, err
	}

	args["categories"] = categories
	args["menu"] = "feeds"
	return args, nil
}

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
	"github.com/miniflux/miniflux/reader/subscription"
	"github.com/miniflux/miniflux/ui/form"
	"github.com/miniflux/miniflux/ui/session"
	"github.com/miniflux/miniflux/ui/view"
)

// SubmitSubscription try to find a feed from the URL provided by the user.
func (c *Controller) SubmitSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	sess := session.New(c.store, ctx)
	v := view.New(c.tpl, ctx, sess)

	user, err := c.store.UserByID(ctx.UserID())
	if err != nil {
		html.ServerError(w, err)
		return
	}

	categories, err := c.store.Categories(user.ID)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	v.Set("categories", categories)
	v.Set("menu", "feeds")
	v.Set("user", user)
	v.Set("countUnread", c.store.CountUnreadEntries(user.ID))

	subscriptionForm := form.NewSubscriptionForm(r)
	if err := subscriptionForm.Validate(); err != nil {
		v.Set("form", subscriptionForm)
		v.Set("errorMessage", err.Error())
		html.OK(w, v.Render("add_subscription"))
		return
	}

	subscriptions, err := subscription.FindSubscriptions(
		subscriptionForm.URL,
		subscriptionForm.Username,
		subscriptionForm.Password,
	)
	if err != nil {
		logger.Error("[Controller:SubmitSubscription] %v", err)
		v.Set("form", subscriptionForm)
		v.Set("errorMessage", err)
		html.OK(w, v.Render("add_subscription"))
		return
	}

	logger.Debug("[UI:SubmitSubscription] %s", subscriptions)

	n := len(subscriptions)
	switch {
	case n == 0:
		v.Set("form", subscriptionForm)
		v.Set("errorMessage", "Unable to find any subscription.")
		html.OK(w, v.Render("add_subscription"))
	case n == 1:
		feed, err := c.feedHandler.CreateFeed(
			user.ID,
			subscriptionForm.CategoryID,
			subscriptions[0].URL,
			subscriptionForm.Crawler,
			subscriptionForm.Username,
			subscriptionForm.Password,
		)
		if err != nil {
			v.Set("form", subscriptionForm)
			v.Set("errorMessage", err)
			html.OK(w, v.Render("add_subscription"))
			return
		}

		response.Redirect(w, r, route.Path(c.router, "feedEntries", "feedID", feed.ID))
	case n > 1:
		v := view.New(c.tpl, ctx, sess)
		v.Set("subscriptions", subscriptions)
		v.Set("form", subscriptionForm)
		v.Set("menu", "feeds")
		v.Set("user", user)
		v.Set("countUnread", c.store.CountUnreadEntries(user.ID))

		html.OK(w, v.Render("choose_subscription"))
	}
}

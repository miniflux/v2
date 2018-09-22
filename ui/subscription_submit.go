// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/client"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/request"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/reader/subscription"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// SubmitSubscription try to find a feed from the URL provided by the user.
func (c *Controller) SubmitSubscription(w http.ResponseWriter, r *http.Request) {
	sess := session.New(c.store, request.SessionID(r))
	v := view.New(c.tpl, r, sess)

	user, err := c.store.UserByID(request.UserID(r))
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
	v.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))
	v.Set("defaultUserAgent", client.DefaultUserAgent)

	subscriptionForm := form.NewSubscriptionForm(r)
	if err := subscriptionForm.Validate(); err != nil {
		v.Set("form", subscriptionForm)
		v.Set("errorMessage", err.Error())
		html.OK(w, r, v.Render("add_subscription"))
		return
	}

	subscriptions, err := subscription.FindSubscriptions(
		subscriptionForm.URL,
		subscriptionForm.UserAgent,
		subscriptionForm.Username,
		subscriptionForm.Password,
	)
	if err != nil {
		logger.Error("[Controller:SubmitSubscription] %v", err)
		v.Set("form", subscriptionForm)
		v.Set("errorMessage", err)
		html.OK(w, r, v.Render("add_subscription"))
		return
	}

	logger.Debug("[UI:SubmitSubscription] %s", subscriptions)

	n := len(subscriptions)
	switch {
	case n == 0:
		v.Set("form", subscriptionForm)
		v.Set("errorMessage", "error.subscription_not_found")
		html.OK(w, r, v.Render("add_subscription"))
	case n == 1:
		feed, err := c.feedHandler.CreateFeed(
			user.ID,
			subscriptionForm.CategoryID,
			subscriptions[0].URL,
			subscriptionForm.Crawler,
			subscriptionForm.UserAgent,
			subscriptionForm.Username,
			subscriptionForm.Password,
		)
		if err != nil {
			v.Set("form", subscriptionForm)
			v.Set("errorMessage", err)
			html.OK(w, r, v.Render("add_subscription"))
			return
		}

		response.Redirect(w, r, route.Path(c.router, "feedEntries", "feedID", feed.ID))
	case n > 1:
		v := view.New(c.tpl, r, sess)
		v.Set("subscriptions", subscriptions)
		v.Set("form", subscriptionForm)
		v.Set("menu", "feeds")
		v.Set("user", user)
		v.Set("countUnread", c.store.CountUnreadEntries(user.ID))
		v.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))

		html.OK(w, r, v.Render("choose_subscription"))
	}
}

// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/client"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/reader/subscription"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) submitSubscription(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	v := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	categories, err := h.store.Categories(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	v.Set("categories", categories)
	v.Set("menu", "feeds")
	v.Set("user", user)
	v.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	v.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))
	v.Set("defaultUserAgent", client.DefaultUserAgent)

	subscriptionForm := form.NewSubscriptionForm(r)
	if err := subscriptionForm.Validate(); err != nil {
		v.Set("form", subscriptionForm)
		v.Set("errorMessage", err.Error())
		html.OK(w, r, v.Render("add_subscription"))
		return
	}

	subscriptions, findErr := subscription.FindSubscriptions(
		subscriptionForm.URL,
		subscriptionForm.UserAgent,
		subscriptionForm.Username,
		subscriptionForm.Password,
	)
	if findErr != nil {
		logger.Error("[UI:SubmitSubscription] %s", findErr)
		v.Set("form", subscriptionForm)
		v.Set("errorMessage", findErr)
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
		feed, err := h.feedHandler.CreateFeed(
			user.ID,
			subscriptionForm.CategoryID,
			subscriptions[0].URL,
			subscriptionForm.Crawler,
			subscriptionForm.UserAgent,
			subscriptionForm.Username,
			subscriptionForm.Password,
			subscriptionForm.ScraperRules,
			subscriptionForm.RewriteRules,
		)
		if err != nil {
			v.Set("form", subscriptionForm)
			v.Set("errorMessage", err)
			html.OK(w, r, v.Render("add_subscription"))
			return
		}

		html.Redirect(w, r, route.Path(h.router, "feedEntries", "feedID", feed.ID))
	case n > 1:
		v := view.New(h.tpl, r, sess)
		v.Set("subscriptions", subscriptions)
		v.Set("form", subscriptionForm)
		v.Set("menu", "feeds")
		v.Set("user", user)
		v.Set("countUnread", h.store.CountUnreadEntries(user.ID))
		v.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))

		html.OK(w, r, v.Render("choose_subscription"))
	}
}

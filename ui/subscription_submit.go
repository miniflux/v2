// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/config"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/model"
	feedHandler "miniflux.app/reader/handler"
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
	v.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	v.Set("defaultUserAgent", config.Opts.HTTPClientUserAgent())
	v.Set("hasProxyConfigured", config.Opts.HasHTTPClientProxyConfigured())

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
		subscriptionForm.Cookie,
		subscriptionForm.Username,
		subscriptionForm.Password,
		subscriptionForm.FetchViaProxy,
		subscriptionForm.AllowSelfSignedCertificates,
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
		feed, err := feedHandler.CreateFeed(h.store, user.ID, &model.FeedCreationRequest{
			CategoryID:                  subscriptionForm.CategoryID,
			FeedURL:                     subscriptions[0].URL,
			Crawler:                     subscriptionForm.Crawler,
			AllowSelfSignedCertificates: subscriptionForm.AllowSelfSignedCertificates,
			ApplyFilterToContent:        subscriptionForm.ApplyFilterToContent,
			UserAgent:                   subscriptionForm.UserAgent,
			Cookie:                      subscriptionForm.Cookie,
			Username:                    subscriptionForm.Username,
			Password:                    subscriptionForm.Password,
			ScraperRules:                subscriptionForm.ScraperRules,
			RewriteRules:                subscriptionForm.RewriteRules,
			BlocklistRules:              subscriptionForm.BlocklistRules,
			KeeplistRules:               subscriptionForm.KeeplistRules,
			FetchViaProxy:               subscriptionForm.FetchViaProxy,
		})
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
		v.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
		v.Set("hasProxyConfigured", config.Opts.HasHTTPClientProxyConfigured())

		html.OK(w, r, v.Render("choose_subscription"))
	}
}

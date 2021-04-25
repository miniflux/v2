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
	"miniflux.app/model"
	feedHandler "miniflux.app/reader/handler"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) showChooseSubscriptionPage(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

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

	view.Set("categories", categories)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("defaultUserAgent", config.Opts.HTTPClientUserAgent())

	subscriptionForm := form.NewSubscriptionForm(r)
	if err := subscriptionForm.Validate(); err != nil {
		view.Set("form", subscriptionForm)
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("add_subscription"))
		return
	}

	feed, err := feedHandler.CreateFeed(h.store, user.ID, &model.FeedCreationRequest{
		CategoryID:                  subscriptionForm.CategoryID,
		FeedURL:                     subscriptionForm.URL,
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
		view.Set("form", subscriptionForm)
		view.Set("errorMessage", err)
		html.OK(w, r, view.Render("add_subscription"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "feedEntries", "feedID", feed.ID))
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	feedHandler "miniflux.app/v2/internal/reader/handler"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showChooseSubscriptionPage(w http.ResponseWriter, r *http.Request) {
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

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("categories", categories)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("defaultUserAgent", config.Opts.HTTPClientUserAgent())

	subscriptionForm := form.NewSubscriptionForm(r)
	if validationErr := subscriptionForm.Validate(); validationErr != nil {
		view.Set("form", subscriptionForm)
		view.Set("errorMessage", validationErr.Translate(user.Language))
		html.OK(w, r, view.Render("add_subscription"))
		return
	}

	feed, localizedError := feedHandler.CreateFeed(h.store, user.ID, &model.FeedCreationRequest{
		CategoryID:                  subscriptionForm.CategoryID,
		FeedURL:                     subscriptionForm.URL,
		Crawler:                     subscriptionForm.Crawler,
		AllowSelfSignedCertificates: subscriptionForm.AllowSelfSignedCertificates,
		UserAgent:                   subscriptionForm.UserAgent,
		Cookie:                      subscriptionForm.Cookie,
		Username:                    subscriptionForm.Username,
		Password:                    subscriptionForm.Password,
		ScraperRules:                subscriptionForm.ScraperRules,
		RewriteRules:                subscriptionForm.RewriteRules,
		UrlRewriteRules:             subscriptionForm.UrlRewriteRules,
		BlocklistRules:              subscriptionForm.BlocklistRules,
		KeeplistRules:               subscriptionForm.KeeplistRules,
		KeepFilterEntryRules:        subscriptionForm.KeepFilterEntryRules,
		BlockFilterEntryRules:       subscriptionForm.BlockFilterEntryRules,
		FetchViaProxy:               subscriptionForm.FetchViaProxy,
		DisableHTTP2:                subscriptionForm.DisableHTTP2,
		ProxyURL:                    subscriptionForm.ProxyURL,
	})
	if localizedError != nil {
		view.Set("form", subscriptionForm)
		view.Set("errorMessage", localizedError.Translate(user.Language))
		html.OK(w, r, view.Render("add_subscription"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "feedEntries", "feedID", feed.ID))
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"
	feedHandler "miniflux.app/v2/internal/reader/handler"
	"miniflux.app/v2/internal/reader/subscription"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) submitSubscription(w http.ResponseWriter, r *http.Request) {
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
	v := view.New(h.tpl, r, sess)
	v.Set("categories", categories)
	v.Set("menu", "feeds")
	v.Set("user", user)
	v.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	v.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	v.Set("defaultUserAgent", config.Opts.HTTPClientUserAgent())
	v.Set("hasProxyConfigured", config.Opts.HasHTTPClientProxyURLConfigured())

	subscriptionForm := form.NewSubscriptionForm(r)
	if validationErr := subscriptionForm.Validate(); validationErr != nil {
		v.Set("form", subscriptionForm)
		v.Set("errorMessage", validationErr.Translate(user.Language))
		html.OK(w, r, v.Render("add_subscription"))
		return
	}

	var rssBridgeURL string
	var rssBridgeToken string
	if intg, err := h.store.Integration(user.ID); err == nil && intg != nil && intg.RSSBridgeEnabled {
		rssBridgeURL = intg.RSSBridgeURL
		rssBridgeToken = intg.RSSBridgeToken
	}

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)
	requestBuilder.WithCustomFeedProxyURL(subscriptionForm.ProxyURL)
	requestBuilder.WithCustomApplicationProxyURL(config.Opts.HTTPClientProxyURL())
	requestBuilder.UseCustomApplicationProxyURL(subscriptionForm.FetchViaProxy)
	requestBuilder.WithUserAgent(subscriptionForm.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(subscriptionForm.Cookie)
	requestBuilder.WithUsernameAndPassword(subscriptionForm.Username, subscriptionForm.Password)
	requestBuilder.IgnoreTLSErrors(subscriptionForm.AllowSelfSignedCertificates)
	requestBuilder.DisableHTTP2(subscriptionForm.DisableHTTP2)

	subscriptionFinder := subscription.NewSubscriptionFinder(requestBuilder)
	subscriptions, localizedError := subscriptionFinder.FindSubscriptions(
		subscriptionForm.URL,
		rssBridgeURL,
		rssBridgeToken,
	)
	if localizedError != nil {
		v.Set("form", subscriptionForm)
		v.Set("errorMessage", localizedError.Translate(user.Language))
		html.OK(w, r, v.Render("add_subscription"))
		return
	}

	n := len(subscriptions)
	switch {
	case n == 0:
		v.Set("form", subscriptionForm)
		v.Set("errorMessage", locale.NewLocalizedError("error.subscription_not_found").Translate(user.Language))
		html.OK(w, r, v.Render("add_subscription"))
	case n == 1 && subscriptionFinder.IsFeedAlreadyDownloaded():
		feed, localizedError := feedHandler.CreateFeedFromSubscriptionDiscovery(h.store, user.ID, &model.FeedCreationRequestFromSubscriptionDiscovery{
			Content:      subscriptionFinder.FeedResponseInfo().Content,
			ETag:         subscriptionFinder.FeedResponseInfo().ETag,
			LastModified: subscriptionFinder.FeedResponseInfo().LastModified,
			FeedCreationRequest: model.FeedCreationRequest{
				CategoryID:                  subscriptionForm.CategoryID,
				FeedURL:                     subscriptions[0].URL,
				AllowSelfSignedCertificates: subscriptionForm.AllowSelfSignedCertificates,
				Crawler:                     subscriptionForm.Crawler,
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
			},
		})
		if localizedError != nil {
			v.Set("form", subscriptionForm)
			v.Set("errorMessage", localizedError.Translate(user.Language))
			html.OK(w, r, v.Render("add_subscription"))
			return
		}

		html.Redirect(w, r, route.Path(h.router, "feedEntries", "feedID", feed.ID))
	case n == 1 && !subscriptionFinder.IsFeedAlreadyDownloaded():
		feed, localizedError := feedHandler.CreateFeed(h.store, user.ID, &model.FeedCreationRequest{
			CategoryID:                  subscriptionForm.CategoryID,
			FeedURL:                     subscriptions[0].URL,
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
			v.Set("form", subscriptionForm)
			v.Set("errorMessage", localizedError.Translate(user.Language))
			html.OK(w, r, v.Render("add_subscription"))
			return
		}

		html.Redirect(w, r, route.Path(h.router, "feedEntries", "feedID", feed.ID))
	case n > 1:
		view := view.New(h.tpl, r, sess)
		view.Set("subscriptions", subscriptions)
		view.Set("form", subscriptionForm)
		view.Set("menu", "feeds")
		view.Set("user", user)
		view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
		view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
		view.Set("hasProxyConfigured", config.Opts.HasHTTPClientProxyURLConfigured())

		html.OK(w, r, view.Render("choose_subscription"))
	}
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"regexp"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

// Best effort url extraction regexp
var urlRe = regexp.MustCompile(`(?i)(?:https?://)?[0-9a-z.]+[.][a-z]+(?::[0-9]+)?(?:/[^ ]+|/)?`)

func (h *handler) bookmarklet(w http.ResponseWriter, r *http.Request) {
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

	bookmarkletURL := request.QueryStringParam(r, "uri", "")

	// Extract URL from text supplied by Web Share Target API.
	//
	// This is because Android intents have no concept of URL, so apps
	// just shove a URL directly into the EXTRA_TEXT intent field.
	//
	// See https://bugs.chromium.org/p/chromium/issues/detail?id=789379.
	text := request.QueryStringParam(r, "text", "")
	if text != "" && bookmarkletURL == "" {
		bookmarkletURL = urlRe.FindString(text)
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", form.SubscriptionForm{URL: bookmarkletURL})
	view.Set("categories", categories)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("defaultUserAgent", config.Opts.HTTPClientUserAgent())
	view.Set("hasProxyConfigured", config.Opts.HasHTTPClientProxyURLConfigured())

	html.OK(w, r, view.Render("add_subscription"))
}

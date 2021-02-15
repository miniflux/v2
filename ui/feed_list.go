// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
	"miniflux.app/validator"
)

func (h *handler) showFeedsPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	feedSortedBy := request.QueryStringParam(r, "feed_sorted_by", user.FeedSortedBy)
	feedDirection := request.QueryStringParam(r, "feed_direction", user.FeedDirection)

	if validationErr := validator.ValidateFeedSortedBy(feedSortedBy); validationErr != nil {
		html.ServerError(w, r, validationErr.Error())
		return
	}
	if validationErr := validator.ValidateFeedDirection(feedDirection); validationErr != nil {
		html.ServerError(w, r, validationErr.Error())
		return
	}

	builder := h.store.NewFeedQueryBuilder(user.ID)
	builder.WithCounters()
	builder.WithOrder(feedSortedBy)
	builder.WithDirection(feedDirection)
	feeds, err := builder.GetFeeds()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("feeds", feeds)
	view.Set("total", len(feeds))
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("feedSortedBy", feedSortedBy)
	view.Set("feedDirection", feedDirection)

	html.OK(w, r, view.Render("feeds"))
}

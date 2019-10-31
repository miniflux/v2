// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/model"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) showStatPage(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithStatus(model.EntryStatusUnread)
	countUnread, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	builder = h.store.NewEntryQueryBuilder(user.ID)
	builder.WithStarred()
	countStarred, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	unreadByFeed, err := h.store.UnreadStatByFeed(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	unreadByCategory, err := h.store.UnreadStatByCategory(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	starredByFeed, err := h.store.StarredStatByFeed(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	starredByCategory, err := h.store.StarredStatByCategory(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	view.Set("unreadByFeed", unreadByFeed)
	view.Set("unreadByCategory", unreadByCategory)
	view.Set("starredByFeed", starredByFeed)
	view.Set("starredByCategory", starredByCategory)
	view.Set("menu", "home")
	view.Set("user", user)
	view.Set("countUnread", countUnread)
	view.Set("countStarred", countStarred)
	view.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))
	view.Set("view", "masonry")

	html.OK(w, r, view.Render("stat"))
}

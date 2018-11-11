// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/reader/opml"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) uploadOPML(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		logger.Error("[Controller:UploadOPML] %v", err)
		html.Redirect(w, r, route.Path(h.router, "import"))
		return
	}
	defer file.Close()

	logger.Info(
		"[Controller:UploadOPML] User #%d uploaded this file: %s (%d bytes)",
		user.ID,
		fileHeader.Filename,
		fileHeader.Size,
	)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))

	if fileHeader.Size == 0 {
		view.Set("errorMessage", "error.empty_file")
		html.OK(w, r, view.Render("import"))
		return
	}

	if impErr := opml.NewHandler(h.store).Import(user.ID, file); impErr != nil {
		view.Set("errorMessage", impErr)
		html.OK(w, r, view.Render("import"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "feeds"))
}

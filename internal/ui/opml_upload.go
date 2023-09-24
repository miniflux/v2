// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/reader/opml"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) uploadOPML(w http.ResponseWriter, r *http.Request) {
	loggedUserID := request.UserID(r)
	user, err := h.store.UserByID(loggedUserID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		slog.Error("OPML file upload error",
			slog.Int64("user_id", loggedUserID),
			slog.Any("error", err),
		)

		html.Redirect(w, r, route.Path(h.router, "import"))
		return
	}
	defer file.Close()

	slog.Info("OPML file uploaded",
		slog.Int64("user_id", loggedUserID),
		slog.String("file_name", fileHeader.Filename),
		slog.Int64("file_size", fileHeader.Size),
	)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

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

func (h *handler) fetchOPML(w http.ResponseWriter, r *http.Request) {
	loggedUserID := request.UserID(r)
	user, err := h.store.UserByID(loggedUserID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		html.Redirect(w, r, route.Path(h.router, "import"))
		return
	}

	slog.Info("Fetching OPML file remotely",
		slog.Int64("user_id", loggedUserID),
		slog.String("opml_file_url", url),
	)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	clt := client.NewClientWithConfig(url, config.Opts)
	resp, err := clt.Get()
	if err != nil {
		view.Set("errorMessage", err)
		html.OK(w, r, view.Render("import"))
		return
	}

	if impErr := opml.NewHandler(h.store).Import(user.ID, resp.Body); impErr != nil {
		view.Set("errorMessage", impErr)
		html.OK(w, r, view.Render("import"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "feeds"))
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"runtime"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/ui/view"
	"miniflux.app/v2/internal/version"
)

func (h *handler) showAboutPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	dbSize, dbErr := h.store.DBSize()

	view := view.New(h.tpl, r)
	view.Set("version", version.Version)
	view.Set("commit", version.Commit)
	view.Set("build_date", version.BuildDate)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("globalConfigOptions", config.Opts.ConfigMap(true))
	view.Set("postgres_version", h.store.DatabaseVersion())
	view.Set("go_version", runtime.Version())

	if dbErr != nil {
		view.Set("db_usage", dbErr)
	} else {
		view.Set("db_usage", dbSize)
	}

	response.HTML(w, r, view.Render("about"))
}

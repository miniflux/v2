// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/context"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/reader/opml"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// UploadOPML handles OPML file importation.
func (c *Controller) UploadOPML(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)

	user, err := c.store.UserByID(ctx.UserID())
	if err != nil {
		html.ServerError(w, err)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		logger.Error("[Controller:UploadOPML] %v", err)
		response.Redirect(w, r, route.Path(c.router, "import"))
		return
	}
	defer file.Close()

	logger.Info(
		"[Controller:UploadOPML] User #%d uploaded this file: %s (%d bytes)",
		user.ID,
		fileHeader.Filename,
		fileHeader.Size,
	)

	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))

	if fileHeader.Size == 0 {
		view.Set("errorMessage", "This file is empty")
		html.OK(w, r, view.Render("import"))
		return
	}

	if impErr := opml.NewHandler(c.store).Import(user.ID, file); impErr != nil {
		view.Set("errorMessage", impErr)
		html.OK(w, r, view.Render("import"))
		return
	}

	response.Redirect(w, r, route.Path(c.router, "feeds"))
}

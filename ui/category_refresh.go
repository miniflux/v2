// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
)

func (h *handler) refreshCategoryEntriesPage(w http.ResponseWriter, r *http.Request) {
	categoryID := h.refreshCategory(w, r)
	html.Redirect(w, r, route.Path(h.router, "categoryEntries", "categoryID", categoryID))
}

func (h *handler) refreshCategoryFeedsPage(w http.ResponseWriter, r *http.Request) {
	categoryID := h.refreshCategory(w, r)
	html.Redirect(w, r, route.Path(h.router, "categoryFeeds", "categoryID", categoryID))
}

func (h *handler) refreshCategory(w http.ResponseWriter, r *http.Request) int64 {
	userID := request.UserID(r)
	categoryID := request.RouteInt64Param(r, "categoryID")

	jobs, err := h.store.NewCategoryBatch(userID, categoryID, h.store.CountFeeds(userID))
	if err != nil {
		html.ServerError(w, r, err)
		return 0
	}

	go func() {
		h.pool.Push(jobs)
	}()

	return categoryID
}

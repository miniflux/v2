// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response/json"
)

// FeedIcon returns a feed icon.
func (c *Controller) FeedIcon(w http.ResponseWriter, r *http.Request) {
	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	if !c.store.HasIcon(feedID) {
		json.NotFound(w, errors.New("This feed doesn't have any icon"))
		return
	}

	icon, err := c.store.IconByFeedID(context.New(r).UserID(), feedID)
	if err != nil {
		json.ServerError(w, errors.New("Unable to fetch feed icon"))
		return
	}

	if icon == nil {
		json.NotFound(w, errors.New("This feed doesn't have any icon"))
		return
	}

	json.OK(w, r, &feedIcon{
		ID:       icon.ID,
		MimeType: icon.MimeType,
		Data:     icon.DataURL(),
	})
}

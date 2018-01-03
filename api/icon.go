// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"

	"github.com/miniflux/miniflux/http/handler"
)

// FeedIcon returns a feed icon.
func (c *Controller) FeedIcon(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	userID := ctx.UserID()
	feedID, err := request.IntegerParam("feedID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	if !c.store.HasIcon(feedID) {
		response.JSON().NotFound(errors.New("This feed doesn't have any icon"))
		return
	}

	icon, err := c.store.IconByFeedID(userID, feedID)
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to fetch feed icon"))
		return
	}

	if icon == nil {
		response.JSON().NotFound(errors.New("This feed doesn't have any icon"))
		return
	}

	response.JSON().Standard(&feedIcon{
		ID:       icon.ID,
		MimeType: icon.MimeType,
		Data:     icon.DataURL(),
	})
}

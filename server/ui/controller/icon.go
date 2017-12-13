// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"time"

	"github.com/miniflux/miniflux/server/core"
)

// ShowIcon shows the feed icon.
func (c *Controller) ShowIcon(ctx *core.Context, request *core.Request, response *core.Response) {
	iconID, err := request.IntegerParam("iconID")
	if err != nil {
		response.HTML().BadRequest(err)
		return
	}

	icon, err := c.store.IconByID(iconID)
	if err != nil {
		response.HTML().ServerError(err)
		return
	}

	if icon == nil {
		response.HTML().NotFound()
		return
	}

	response.Cache(icon.MimeType, icon.Hash, icon.Content, 72*time.Hour)
}

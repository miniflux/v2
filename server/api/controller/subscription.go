// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"fmt"
	"github.com/miniflux/miniflux2/reader/subscription"
	"github.com/miniflux/miniflux2/server/api/payload"
	"github.com/miniflux/miniflux2/server/core"
)

// GetSubscriptions is the API handler to find subscriptions.
func (c *Controller) GetSubscriptions(ctx *core.Context, request *core.Request, response *core.Response) {
	websiteURL, err := payload.DecodeURLPayload(request.Body())
	if err != nil {
		response.Json().BadRequest(err)
		return
	}

	subscriptions, err := subscription.FindSubscriptions(websiteURL)
	if err != nil {
		response.Json().ServerError(errors.New("Unable to discover subscriptions"))
		return
	}

	if subscriptions == nil {
		response.Json().NotFound(fmt.Errorf("No subscription found"))
		return
	}

	response.Json().Standard(subscriptions)
}

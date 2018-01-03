// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"fmt"

	"github.com/miniflux/miniflux/http/handler"
	"github.com/miniflux/miniflux/reader/subscription"
)

// GetSubscriptions is the API handler to find subscriptions.
func (c *Controller) GetSubscriptions(ctx *handler.Context, request *handler.Request, response *handler.Response) {
	websiteURL, err := decodeURLPayload(request.Body())
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	subscriptions, err := subscription.FindSubscriptions(websiteURL)
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to discover subscriptions"))
		return
	}

	if subscriptions == nil {
		response.JSON().NotFound(fmt.Errorf("No subscription found"))
		return
	}

	response.JSON().Standard(subscriptions)
}

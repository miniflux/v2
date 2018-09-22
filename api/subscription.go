// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"fmt"
	"net/http"

	"miniflux.app/http/response/json"
	"miniflux.app/reader/subscription"
)

// GetSubscriptions is the API handler to find subscriptions.
func (c *Controller) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	subscriptionInfo, err := decodeURLPayload(r.Body)
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	subscriptions, err := subscription.FindSubscriptions(
		subscriptionInfo.URL,
		subscriptionInfo.UserAgent,
		subscriptionInfo.Username,
		subscriptionInfo.Password,
	)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	if subscriptions == nil {
		json.NotFound(w, fmt.Errorf("No subscription found"))
		return
	}

	json.OK(w, r, subscriptions)
}

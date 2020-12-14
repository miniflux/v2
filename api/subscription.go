// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"net/http"

	"miniflux.app/http/response/json"
	"miniflux.app/reader/subscription"
)

func (h *handler) getSubscriptions(w http.ResponseWriter, r *http.Request) {
	subscriptionRequest, bodyErr := decodeSubscriptionDiscoveryRequest(r.Body)
	if bodyErr != nil {
		json.BadRequest(w, r, bodyErr)
		return
	}

	subscriptions, finderErr := subscription.FindSubscriptions(
		subscriptionRequest.URL,
		subscriptionRequest.UserAgent,
		subscriptionRequest.Username,
		subscriptionRequest.Password,
		subscriptionRequest.FetchViaProxy,
	)
	if finderErr != nil {
		json.ServerError(w, r, finderErr)
		return
	}

	if subscriptions == nil {
		json.NotFound(w, r)
		return
	}

	json.OK(w, r, subscriptions)
}

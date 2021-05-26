// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	json_parser "encoding/json"
	"net/http"

	"miniflux.app/http/response/json"
	"miniflux.app/model"
	"miniflux.app/reader/subscription"
	"miniflux.app/validator"
)

func (h *handler) discoverSubscriptions(w http.ResponseWriter, r *http.Request) {
	var subscriptionDiscoveryRequest model.SubscriptionDiscoveryRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&subscriptionDiscoveryRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if validationErr := validator.ValidateSubscriptionDiscovery(&subscriptionDiscoveryRequest); validationErr != nil {
		json.BadRequest(w, r, validationErr.Error())
		return
	}

	subscriptions, finderErr := subscription.FindSubscriptions(
		subscriptionDiscoveryRequest.URL,
		subscriptionDiscoveryRequest.UserAgent,
		subscriptionDiscoveryRequest.Cookie,
		subscriptionDiscoveryRequest.Username,
		subscriptionDiscoveryRequest.Password,
		subscriptionDiscoveryRequest.FetchViaProxy,
		subscriptionDiscoveryRequest.AllowSelfSignedCertificates,
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

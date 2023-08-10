// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/api"

import (
	json_parser "encoding/json"
	"net/http"

	"miniflux.app/v2/http/response/json"
	"miniflux.app/v2/model"
	"miniflux.app/v2/reader/subscription"
	"miniflux.app/v2/validator"
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

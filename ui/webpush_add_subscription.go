package ui

import (
	"errors"
	"miniflux.app/logger"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
)

func (h *handler) addWebpushSubscription(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Reached subscription service!")
	subscription, err := decodeCreateWebpushSubscriptionPayload(r.Body)
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if subscription.Subscription == "" {
		json.BadRequest(w, r, errors.New("The subscription is required"))
		return
	}

	userID := request.UserID(r)
	subscription.UserID = userID

	if err := h.store.CreateWebpushSubscription(subscription); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, subscription)
}

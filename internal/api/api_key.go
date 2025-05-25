// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	json_parser "encoding/json"
	"errors"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) createAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	var apiKeyCreationRequest model.APIKeyCreationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&apiKeyCreationRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if validationErr := validator.ValidateAPIKeyCreation(h.store, userID, &apiKeyCreationRequest); validationErr != nil {
		json.BadRequest(w, r, validationErr.Error())
		return
	}

	apiKey, err := h.store.CreateAPIKey(userID, apiKeyCreationRequest.Description)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, apiKey)
}

func (h *handler) getAPIKeys(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	apiKeys, err := h.store.APIKeys(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	json.OK(w, r, apiKeys)
}

func (h *handler) deleteAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	apiKeyID := request.RouteInt64Param(r, "apiKeyID")

	if err := h.store.DeleteAPIKey(userID, apiKeyID); err != nil {
		if errors.Is(err, storage.ErrAPIKeyNotFound) {
			json.NotFound(w, r)
			return
		}
		json.ServerError(w, r, err)
		return
	}
	json.NoContent(w, r)
}

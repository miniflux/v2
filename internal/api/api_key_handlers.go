// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	json_parser "encoding/json"
	"errors"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) createAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	var apiKeyCreationRequest model.APIKeyCreationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&apiKeyCreationRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if validationErr := validator.ValidateAPIKeyCreation(h.store, userID, &apiKeyCreationRequest); validationErr != nil {
		response.JSONBadRequest(w, r, validationErr.Error())
		return
	}

	apiKey, err := h.store.CreateAPIKey(userID, apiKeyCreationRequest.Description)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSONCreated(w, r, apiKey)
}

func (h *handler) getAPIKeysHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	apiKeys, err := h.store.APIKeys(userID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	response.JSON(w, r, apiKeys)
}

func (h *handler) deleteAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	apiKeyID := request.RouteInt64Param(r, "apiKeyID")
	if apiKeyID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid API key ID"))
		return
	}

	if err := h.store.DeleteAPIKey(userID, apiKeyID); err != nil {
		if errors.Is(err, storage.ErrAPIKeyNotFound) {
			response.JSONNotFound(w, r)
			return
		}
		response.JSONServerError(w, r, err)
		return
	}
	response.NoContent(w, r)
}

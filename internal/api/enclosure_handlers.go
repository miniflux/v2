// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	json_parser "encoding/json"
	"errors"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) getEnclosureByIDHandler(w http.ResponseWriter, r *http.Request) {
	enclosureID := request.RouteInt64Param(r, "enclosureID")
	if enclosureID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid enclosure ID"))
		return
	}

	enclosure, err := h.store.GetEnclosure(enclosureID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if enclosure == nil {
		response.JSONNotFound(w, r)
		return
	}

	userID := request.UserID(r)
	if enclosure.UserID != userID {
		response.JSONNotFound(w, r)
		return
	}

	enclosure.ProxifyEnclosureURL(config.Opts.MediaProxyMode(), config.Opts.MediaProxyResourceTypes())

	response.JSON(w, r, enclosure)
}

func (h *handler) updateEnclosureByIDHandler(w http.ResponseWriter, r *http.Request) {
	enclosureID := request.RouteInt64Param(r, "enclosureID")
	if enclosureID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid enclosure ID"))
		return
	}

	var enclosureUpdateRequest model.EnclosureUpdateRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&enclosureUpdateRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if err := validator.ValidateEnclosureUpdateRequest(&enclosureUpdateRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	enclosure, err := h.store.GetEnclosure(enclosureID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if enclosure == nil {
		response.JSONNotFound(w, r)
		return
	}

	userID := request.UserID(r)
	if enclosure.UserID != userID {
		response.JSONNotFound(w, r)
		return
	}

	enclosure.MediaProgression = enclosureUpdateRequest.MediaProgression
	if err := h.store.UpdateEnclosure(enclosure); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.NoContent(w, r)
}

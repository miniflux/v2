// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	json2 "encoding/json"
	"io"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
)

type enclosurePositionSaveRequest struct {
	Progression int64 `json:"progression"`
}

func (h *handler) saveEnclosureProgression(w http.ResponseWriter, r *http.Request) {
	enclosureID := request.RouteInt64Param(r, "enclosureID")
	enclosure, err := h.store.GetEnclosure(enclosureID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	var postData enclosurePositionSaveRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json2.Unmarshal(body, &postData)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	enclosure.MediaProgression = postData.Progression

	err = h.store.UpdateEnclosure(enclosure)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, map[string]string{"message": "saved"})
}

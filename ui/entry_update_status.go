// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"errors"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
)

// UpdateEntriesStatus updates the status for a list of entries.
func (c *Controller) UpdateEntriesStatus(w http.ResponseWriter, r *http.Request) {
	entryIDs, status, err := decodeEntryStatusPayload(r.Body)
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if len(entryIDs) == 0 {
		json.BadRequest(w, r, errors.New("The list of entry IDs is empty"))
		return
	}

	err = c.store.SetEntriesStatus(request.UserID(r), entryIDs, status)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, "OK")
}

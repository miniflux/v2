// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"errors"
	"net/http"

	"miniflux.app/http/context"
	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/integration"
	"miniflux.app/model"
)

// SaveEntry send the link to external services.
func (c *Controller) SaveEntry(w http.ResponseWriter, r *http.Request) {
	entryID, err := request.IntParam(r, "entryID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	ctx := context.New(r)

	builder := c.store.NewEntryQueryBuilder(ctx.UserID())
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		json.ServerError(w, err)
		return
	}

	if entry == nil {
		json.NotFound(w, errors.New("Entry not found"))
		return
	}

	settings, err := c.store.Integration(ctx.UserID())
	if err != nil {
		json.ServerError(w, err)
		return
	}

	go func() {
		integration.SendEntry(c.cfg, entry, settings)
	}()

	json.Created(w, map[string]string{"message": "saved"})
}

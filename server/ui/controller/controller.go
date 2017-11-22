// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/feed"
	"github.com/miniflux/miniflux2/reader/opml"
	"github.com/miniflux/miniflux2/scheduler"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/storage"
)

type tplParams map[string]interface{}

func (t tplParams) Merge(d tplParams) tplParams {
	for k, v := range d {
		t[k] = v
	}

	return t
}

// Controller contains all HTTP handlers for the user interface.
type Controller struct {
	store       *storage.Storage
	pool        *scheduler.WorkerPool
	feedHandler *feed.Handler
	opmlHandler *opml.Handler
}

func (c *Controller) getCommonTemplateArgs(ctx *core.Context) (tplParams, error) {
	user := ctx.LoggedUser()
	builder := c.store.GetEntryQueryBuilder(user.ID, user.Timezone)
	builder.WithStatus(model.EntryStatusUnread)

	countUnread, err := builder.CountEntries()
	if err != nil {
		return nil, err
	}

	params := tplParams{
		"menu":        "",
		"user":        user,
		"countUnread": countUnread,
		"csrf":        ctx.CsrfToken(),
	}
	return params, nil
}

// NewController returns a new Controller.
func NewController(store *storage.Storage, pool *scheduler.WorkerPool, feedHandler *feed.Handler, opmlHandler *opml.Handler) *Controller {
	return &Controller{
		store:       store,
		pool:        pool,
		feedHandler: feedHandler,
		opmlHandler: opmlHandler,
	}
}

// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"github.com/miniflux/miniflux2/reader/feed"
	"github.com/miniflux/miniflux2/storage"
)

// Controller holds all handlers for the API.
type Controller struct {
	store       *storage.Storage
	feedHandler *feed.Handler
}

// NewController creates a new controller.
func NewController(store *storage.Storage, feedHandler *feed.Handler) *Controller {
	return &Controller{store: store, feedHandler: feedHandler}
}

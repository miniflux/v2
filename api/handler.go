// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"miniflux.app/reader/feed"
	"miniflux.app/storage"
	"miniflux.app/worker"
)

type handler struct {
	store       *storage.Storage
	pool        *worker.Pool
	feedHandler *feed.Handler
}

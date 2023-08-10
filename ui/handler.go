// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/ui"

import (
	"miniflux.app/v2/storage"
	"miniflux.app/v2/template"
	"miniflux.app/v2/worker"

	"github.com/gorilla/mux"
)

type handler struct {
	router *mux.Router
	store  *storage.Storage
	tpl    *template.Engine
	pool   *worker.Pool
}

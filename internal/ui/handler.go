// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "influxeed-engine/v2/internal/ui"

import (
	"influxeed-engine/v2/internal/storage"
	"influxeed-engine/v2/internal/template"
	"influxeed-engine/v2/internal/worker"

	"github.com/gorilla/mux"
)

type handler struct {
	router *mux.Router
	store  *storage.Storage
	tpl    *template.Engine
	pool   *worker.Pool
}

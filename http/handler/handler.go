// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package handler

import (
	"net/http"
	"time"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/locale"
	"github.com/miniflux/miniflux/storage"
	"github.com/miniflux/miniflux/template"
	"github.com/miniflux/miniflux/timer"

	"github.com/gorilla/mux"
)

// ControllerFunc is an application HTTP handler.
type ControllerFunc func(ctx *Context, request *Request, response *Response)

// Handler manages HTTP handlers.
type Handler struct {
	cfg        *config.Config
	store      *storage.Storage
	translator *locale.Translator
	template   *template.Engine
	router     *mux.Router
}

// Use is a wrapper around an HTTP handler.
func (h *Handler) Use(f ControllerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer timer.ExecutionTime(time.Now(), r.URL.Path)

		ctx := NewContext(r, h.store, h.router, h.translator)
		request := NewRequest(r)
		response := NewResponse(h.cfg, w, r, h.template)

		f(ctx, request, response)
	})
}

// NewHandler returns a new Handler.
func NewHandler(cfg *config.Config, store *storage.Storage, router *mux.Router, template *template.Engine, translator *locale.Translator) *Handler {
	return &Handler{
		cfg:        cfg,
		store:      store,
		translator: translator,
		router:     router,
		template:   template,
	}
}

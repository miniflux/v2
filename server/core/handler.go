// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package core

import (
	"net/http"
	"time"

	"github.com/miniflux/miniflux/helper"
	"github.com/miniflux/miniflux/locale"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/server/middleware"
	"github.com/miniflux/miniflux/server/template"
	"github.com/miniflux/miniflux/storage"

	"github.com/gorilla/mux"
)

// HandlerFunc is an application HTTP handler.
type HandlerFunc func(ctx *Context, request *Request, response *Response)

// Handler manages HTTP handlers and middlewares.
type Handler struct {
	store      *storage.Storage
	translator *locale.Translator
	template   *template.Engine
	router     *mux.Router
	middleware *middleware.Chain
}

// Use is a wrapper around an HTTP handler.
func (h *Handler) Use(f HandlerFunc) http.Handler {
	return h.middleware.WrapFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer helper.ExecutionTime(time.Now(), r.URL.Path)
		logger.Debug("[HTTP] %s %s", r.Method, r.URL.Path)

		ctx := NewContext(w, r, h.store, h.router)
		request := NewRequest(w, r)
		response := NewResponse(w, r, h.template)

		if ctx.IsAuthenticated() {
			h.template.SetLanguage(ctx.UserLanguage())
		} else {
			h.template.SetLanguage("en_US")
		}

		f(ctx, request, response)
	}))
}

// NewHandler returns a new Handler.
func NewHandler(store *storage.Storage, router *mux.Router, template *template.Engine, translator *locale.Translator, middleware *middleware.Chain) *Handler {
	return &Handler{
		store:      store,
		translator: translator,
		router:     router,
		template:   template,
		middleware: middleware,
	}
}

// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package handler

import (
	"net/http"
	"time"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/http/middleware"
	"github.com/miniflux/miniflux/locale"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/storage"
	"github.com/miniflux/miniflux/template"
	"github.com/miniflux/miniflux/timer"

	"github.com/gorilla/mux"
	"github.com/tomasen/realip"
)

// ControllerFunc is an application HTTP handler.
type ControllerFunc func(ctx *Context, request *Request, response *Response)

// Handler manages HTTP handlers and middlewares.
type Handler struct {
	cfg        *config.Config
	store      *storage.Storage
	translator *locale.Translator
	template   *template.Engine
	router     *mux.Router
	middleware *middleware.Chain
}

// Use is a wrapper around an HTTP handler.
func (h *Handler) Use(f ControllerFunc) http.Handler {
	return h.middleware.WrapFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer timer.ExecutionTime(time.Now(), r.URL.Path)
		logger.Debug("[HTTP] %s %s %s", realip.RealIP(r), r.Method, r.URL.Path)

		if r.Header.Get("X-Forwarded-Proto") == "https" {
			h.cfg.IsHTTPS = true
		}

		ctx := NewContext(r, h.store, h.router, h.translator)
		request := NewRequest(r)
		response := NewResponse(w, r, h.template)
		language := ctx.UserLanguage()

		if language != "" {
			h.template.SetLanguage(language)
		} else {
			h.template.SetLanguage("en_US")
		}

		f(ctx, request, response)
	}))
}

// NewHandler returns a new Handler.
func NewHandler(cfg *config.Config, store *storage.Storage, router *mux.Router, template *template.Engine, translator *locale.Translator, middleware *middleware.Chain) *Handler {
	return &Handler{
		cfg:        cfg,
		store:      store,
		translator: translator,
		router:     router,
		template:   template,
		middleware: middleware,
	}
}

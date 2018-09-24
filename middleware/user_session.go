// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware // import "miniflux.app/middleware"

import (
	"context"
	"net/http"

	"miniflux.app/http/cookie"
	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/model"

	"github.com/gorilla/mux"
)

// UserSession handles the user session middleware.
func (m *Middleware) UserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := m.getUserSessionFromCookie(r)

		if session == nil {
			logger.Debug("[Middleware:UserSession] Session not found")
			if m.isPublicRoute(r) {
				next.ServeHTTP(w, r)
			} else {
				response.Redirect(w, r, route.Path(m.router, "login"))
			}
		} else {
			logger.Debug("[Middleware:UserSession] %s", session)

			ctx := r.Context()
			ctx = context.WithValue(ctx, request.UserIDContextKey, session.UserID)
			ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)
			ctx = context.WithValue(ctx, request.UserSessionTokenContextKey, session.Token)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func (m *Middleware) isPublicRoute(r *http.Request) bool {
	route := mux.CurrentRoute(r)
	switch route.GetName() {
	case "login",
		"checkLogin",
		"stylesheet",
		"javascript",
		"oauth2Redirect",
		"oauth2Callback",
		"appIcon",
		"favicon",
		"webManifest":
		return true
	default:
		return false
	}
}

func (m *Middleware) getUserSessionFromCookie(r *http.Request) *model.UserSession {
	cookieValue := request.CookieValue(r, cookie.CookieUserSessionID)
	if cookieValue == "" {
		return nil
	}

	session, err := m.store.UserSessionByToken(cookieValue)
	if err != nil {
		logger.Error("[Middleware:UserSession] %v", err)
		return nil
	}

	return session
}

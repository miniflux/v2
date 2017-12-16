// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"net/http"

	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/server/route"
	"github.com/miniflux/miniflux/storage"

	"github.com/gorilla/mux"
)

// SessionMiddleware represents a session middleware.
type SessionMiddleware struct {
	store  *storage.Storage
	router *mux.Router
}

// Handler execute the middleware.
func (s *SessionMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := s.getSessionFromCookie(r)

		if session == nil {
			logger.Debug("[Middleware:Session] Session not found")
			if s.isPublicRoute(r) {
				next.ServeHTTP(w, r)
			} else {
				http.Redirect(w, r, route.Path(s.router, "login"), http.StatusFound)
			}
		} else {
			logger.Debug("[Middleware:Session] %s", session)
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDContextKey, session.UserID)
			ctx = context.WithValue(ctx, IsAuthenticatedContextKey, true)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func (s *SessionMiddleware) isPublicRoute(r *http.Request) bool {
	route := mux.CurrentRoute(r)
	switch route.GetName() {
	case "login", "checkLogin", "stylesheet", "javascript", "oauth2Redirect", "oauth2Callback", "appIcon", "favicon":
		return true
	default:
		return false
	}
}

func (s *SessionMiddleware) getSessionFromCookie(r *http.Request) *model.Session {
	sessionCookie, err := r.Cookie("sessionID")
	if err == http.ErrNoCookie {
		return nil
	}

	session, err := s.store.SessionByToken(sessionCookie.Value)
	if err != nil {
		logger.Error("[SessionMiddleware] %v", err)
		return nil
	}

	return session
}

// NewSessionMiddleware returns a new SessionMiddleware.
func NewSessionMiddleware(s *storage.Storage, r *mux.Router) *SessionMiddleware {
	return &SessionMiddleware{store: s, router: r}
}

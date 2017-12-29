// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"net/http"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/server/cookie"
	"github.com/miniflux/miniflux/storage"
)

// SessionMiddleware represents a session middleware.
type SessionMiddleware struct {
	cfg   *config.Config
	store *storage.Storage
}

// Handler execute the middleware.
func (s *SessionMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		session := s.getSessionValueFromCookie(r)

		if session == nil {
			logger.Debug("[Middleware:Session] Session not found")
			session, err = s.store.CreateSession()
			if err != nil {
				logger.Error("[Middleware:Session] %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, cookie.New(cookie.CookieSessionID, session.ID, s.cfg.IsHTTPS))
		} else {
			logger.Debug("[Middleware:Session] %s", session)
		}

		if r.Method == "POST" {
			formValue := r.FormValue("csrf")
			headerValue := r.Header.Get("X-Csrf-Token")

			if session.Data.CSRF != formValue && session.Data.CSRF != headerValue {
				logger.Error(`[Middleware:Session] Invalid or missing CSRF token: Form="%s", Header="%s"`, formValue, headerValue)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid or missing CSRF session!"))
				return
			}
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, SessionIDContextKey, session.ID)
		ctx = context.WithValue(ctx, CSRFContextKey, session.Data.CSRF)
		ctx = context.WithValue(ctx, OAuth2StateContextKey, session.Data.OAuth2State)
		ctx = context.WithValue(ctx, FlashMessageContextKey, session.Data.FlashMessage)
		ctx = context.WithValue(ctx, FlashErrorMessageContextKey, session.Data.FlashErrorMessage)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *SessionMiddleware) getSessionValueFromCookie(r *http.Request) *model.Session {
	sessionCookie, err := r.Cookie(cookie.CookieSessionID)
	if err == http.ErrNoCookie {
		return nil
	}

	session, err := s.store.Session(sessionCookie.Value)
	if err != nil {
		logger.Error("[Middleware:Session] %v", err)
		return nil
	}

	return session
}

// NewSessionMiddleware returns a new SessionMiddleware.
func NewSessionMiddleware(cfg *config.Config, store *storage.Storage) *SessionMiddleware {
	return &SessionMiddleware{cfg, store}
}
